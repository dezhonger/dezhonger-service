package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBlogValidationAndPayloadPrivacy(t *testing.T) {
	tests := []struct {
		name  string
		input blogPostRequest
		valid bool
	}{
		{"empty draft", blogPostRequest{}, true},
		{"empty publication", blogPostRequest{Status: "published"}, false},
		{"blank text", blogPostRequest{Blocks: []BlogBlock{{Type: "paragraph", Text: " "}}}, false},
		{"external media", blogPostRequest{Blocks: []BlogBlock{{Type: "image", MediaID: "https://example.com/photo"}}}, false},
		{"html block type", blogPostRequest{Blocks: []BlogBlock{{Type: "html", Text: "<script>"}}}, false},
		{"text remains plain", blogPostRequest{Title: "Title", Status: "published", Blocks: []BlogBlock{{Type: "paragraph", Text: "<script>alert(1)</script>"}}}, true},
		{"200 blocks", blogPostRequest{Blocks: make([]BlogBlock, 200)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBlogPost(&tt.input)
			if (err == nil) != tt.valid {
				t.Fatalf("valid=%v err=%v", tt.valid, err)
			}
		})
	}
	secret := BlogPost{ID: "post", Title: "visible", Protected: true, Blocks: []BlogBlock{{Type: "paragraph", Text: "secret-text"}, {Type: "image", MediaID: "secret-media", Caption: "secret-caption"}}}
	for _, payload := range []any{publicBlogSummary(secret), publicBlogDetail(secret, false)} {
		raw, _ := json.Marshal(payload)
		if bytes.Contains(raw, []byte("secret")) || bytes.Contains(raw, []byte("blocks")) || bytes.Contains(raw, []byte("cover_url")) {
			t.Fatalf("locked payload leaked content: %s", raw)
		}
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"title":"x","blocks":[{"type":"paragraph","text":"ok","html":"bad"}]}`))
	recorder := httptest.NewRecorder()
	if _, ok := readBlogPostRequest(recorder, req); ok || recorder.Code != 400 {
		t.Fatal("unknown nested block property was accepted")
	}
}

func TestBlogOriginAllowlist(t *testing.T) {
	a := &App{cfg: Config{PublicOrigin: "https://dezhonger.com", BlogOrigin: "https://zmq.dezhonger.com"}}
	handler := a.validateOrigin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	tests := []struct {
		name, method, path, origin string
		want                       int
	}{
		{"blog unlock", "POST", "/api/blog/posts/id/unlock", a.cfg.BlogOrigin, 204},
		{"blog create", "POST", "/api/blog/admin/posts", a.cfg.BlogOrigin, 204},
		{"blog update", "PUT", "/api/blog/admin/posts/id", a.cfg.BlogOrigin, 204},
		{"blog delete", "DELETE", "/api/blog/admin/posts/id", a.cfg.BlogOrigin, 204},
		{"blog upload", "POST", "/api/blog/admin/posts/id/media", a.cfg.BlogOrigin, 204},
		{"blog login", "POST", "/api/auth/login", a.cfg.BlogOrigin, 204},
		{"blog logout", "POST", "/api/auth/logout", a.cfg.BlogOrigin, 204},
		{"blog cannot write notes", "PUT", "/api/notes/id", a.cfg.BlogOrigin, 403},
		{"blog cannot create users", "POST", "/api/admin/users", a.cfg.BlogOrigin, 403},
		{"blog cannot modify users", "PATCH", "/api/admin/users/id", a.cfg.BlogOrigin, 403},
		{"blog cannot change account password", "PUT", "/api/account/password", a.cfg.BlogOrigin, 403},
		{"blog prefix boundary", "POST", "/api/blogger/posts", a.cfg.BlogOrigin, 403},
		{"exact auth path", "POST", "/api/auth/login/extra", a.cfg.BlogOrigin, 403},
		{"main writes notes", "PUT", "/api/notes/id", a.cfg.PublicOrigin, 204},
		{"main creates users", "POST", "/api/admin/users", a.cfg.PublicOrigin, 204},
		{"main changes account password", "PUT", "/api/account/password", a.cfg.PublicOrigin, 204},
		{"main creates blog", "POST", "/api/blog/admin/posts", a.cfg.PublicOrigin, 204},
		{"no origin compatibility", "PUT", "/api/notes/id", "", 204},
		{"untrusted sibling", "POST", "/api/blog/posts/id/unlock", "https://evil.dezhonger.com", 403},
		{"untrusted suffix", "POST", "/api/blog/posts/id/unlock", "https://zmq.dezhonger.com.evil.test", 403},
		{"null origin", "POST", "/api/blog/posts/id/unlock", "null", 403},
		{"get unchanged", "GET", "/api/notes", "https://untrusted.example", 204},
		{"head unchanged", "HEAD", "/api/blog/media/id", "https://untrusted.example", 204},
		{"options unchanged", "OPTIONS", "/api/admin/users", "https://untrusted.example", 204},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			if test.origin != "" {
				req.Header.Set("Origin", test.origin)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != test.want {
				t.Fatalf("got HTTP %d want %d: %s", w.Code, test.want, w.Body.String())
			}
		})
	}
}

func TestBlogReaderCookieAndRateLimit(t *testing.T) {
	a := &App{cfg: Config{CookieSecure: true}}
	token, _, err := newSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	if !validBlogReaderToken(token) || validBlogReaderToken("chosen-token") {
		t.Fatal("reader token validation failed")
	}
	w := httptest.NewRecorder()
	a.setBlogReaderCookie(w, token)
	cookie := w.Result().Cookies()[0]
	if cookie.Name != "__Host-dz_blog_reader" || !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode || cookie.Domain != "" || cookie.Path != "/" {
		t.Fatalf("unsafe cookie: %#v", cookie)
	}
	var limiter blogAttemptLimiter
	now := time.Now()
	for i := 0; i < blogAttemptLimit; i++ {
		if !limiter.allow("client", now) {
			t.Fatal("limited before threshold")
		}
	}
	if limiter.allow("client", now) {
		t.Fatal("brute-force threshold not enforced")
	}
	if !limiter.allow("client", now.Add(blogAttemptWindow)) {
		t.Fatal("expired window did not reset")
	}
	var bounded blogAttemptLimiter
	for i := 0; i < blogAttemptMaxKeys; i++ {
		if !bounded.allow(fmt.Sprint(i), now) {
			t.Fatal("unexpected cap")
		}
	}
	if bounded.allow("new-client", now) || len(bounded.entries) > blogAttemptMaxKeys {
		t.Fatal("rate-limit memory was not bounded")
	}
}

func TestBlogMediaSniffAndNames(t *testing.T) {
	for _, raw := range []string{"<svg xmlns='http://www.w3.org/2000/svg'></svg>", "<html><script>alert(1)</script></html>", "plain-text"} {
		if mime, _, limit := blogMediaFormat([]byte(raw)); mime != "" || limit != 0 {
			t.Fatalf("unsafe MIME accepted: %q", mime)
		}
	}
	if mime, _, limit := blogMediaFormat([]byte("\x89PNG\r\n\x1a\n")); mime != "image/png" || limit != blogImageMaxBytes {
		t.Fatal("PNG detection failed")
	}
	if name := blogUploadName("..\\..\\photo.jpg"); name != "photo.jpg" {
		t.Fatalf("unsafe file name %q", name)
	}
}
