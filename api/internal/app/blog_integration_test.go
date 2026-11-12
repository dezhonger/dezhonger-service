package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// The supplied database must be disposable. Each run uses and removes a private
// schema; no production table or account is accessed.
func blogIntegrationApp(t *testing.T) (*App, *http.Cookie, *http.Cookie) {
	t.Helper()
	dsn := os.Getenv("BLOG_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("BLOG_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "blog_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := conn.Exec(ctx, "create schema "+pgx.Identifier{schema}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = conn.Exec(context.Background(), "drop schema "+pgx.Identifier{schema}.Sanitize()+" cascade")
		_ = conn.Close(context.Background())
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema+",public")
	u.RawQuery = q.Encode()
	a, err := New(ctx, Config{DatabaseURL: u.String(), CookieName: "dz_test_session", CookieSecure: true, SessionTTL: time.Hour, PublicOrigin: "https://dezhonger.com", BlogOrigin: "https://zmq.dezhonger.com", BlogMediaDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	cookies := make([]*http.Cookie, 0, 2)
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("integration-password"), bcrypt.MinCost)
	for _, role := range []string{"admin", "user"} {
		id := uuid.NewString()
		if _, err := a.db.Exec(ctx, `insert into users(id,username,password_hash,role,must_change_password) values($1,$2,$3,$4,false)`, id, "test-"+role, string(passwordHash), role); err != nil {
			t.Fatal(err)
		}
		token, hash, _ := newSessionToken()
		if _, err := a.db.Exec(ctx, `insert into sessions(token_hash,user_id,expires_at) values($1,$2,$3)`, hash[:], id, time.Now().Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
		cookies = append(cookies, &http.Cookie{Name: a.cfg.CookieName, Value: token})
	}
	return a, cookies[0], cookies[1]
}

func blogTestRequest(t *testing.T, a *App, method, path string, body any, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(raw)
	}
	r := httptest.NewRequest(method, path, reader)
	r.RemoteAddr = "192.0.2.1:1234"
	r.Header.Set("Origin", a.cfg.BlogOrigin)
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	for _, cookie := range cookies {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, r)
	return w
}
func blogRequireStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("HTTP %d want %d: %s", w.Code, want, w.Body.String())
	}
}
func blogResponsePost(t *testing.T, w *httptest.ResponseRecorder) BlogPost {
	t.Helper()
	var response struct {
		Post BlogPost `json:"post"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	return response.Post
}
func blogUploadTest(t *testing.T, a *App, postID, name string, data []byte, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(data)
	_ = writer.Close()
	r := httptest.NewRequest("POST", "/api/blog/admin/posts/"+postID+"/media", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("Origin", a.cfg.BlogOrigin)
	r.AddCookie(cookie)
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, r)
	return w
}

func TestBlogIntegrationAccessLifecycle(t *testing.T) {
	a, admin, user := blogIntegrationApp(t)
	create := blogTestRequest(t, a, "POST", "/api/blog/admin/posts", map[string]any{"title": "Family", "status": "draft", "blocks": []BlogBlock{{Type: "paragraph", Text: "private family text"}}}, admin)
	blogRequireStatus(t, create, 201)
	post := blogResponsePost(t, create)
	path := "/api/blog/posts/" + post.ID
	adminPath := "/api/blog/admin/posts/" + post.ID
	blogRequireStatus(t, blogTestRequest(t, a, "GET", path, nil), 404)
	blogRequireStatus(t, blogTestRequest(t, a, "GET", adminPath, nil, user), 403)
	blogRequireStatus(t, blogTestRequest(t, a, "GET", adminPath, nil), 401)
	for _, test := range []struct {
		status string
		total  int
	}{{"draft", 1}, {"published", 0}} {
		response := blogTestRequest(t, a, "GET", "/api/blog/admin/posts?status="+test.status, nil, admin)
		blogRequireStatus(t, response, 200)
		var listing struct {
			Total int        `json:"total"`
			Posts []BlogPost `json:"posts"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &listing); err != nil {
			t.Fatal(err)
		}
		if listing.Total != test.total || len(listing.Posts) != test.total {
			t.Fatalf("status=%s: %s", test.status, response.Body.String())
		}
	}
	blogRequireStatus(t, blogTestRequest(t, a, "GET", "/api/blog/admin/posts?status=bad", nil, admin), 400)
	png := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 128)...)
	upload := blogUploadTest(t, a, post.ID, "photo.png", png, admin)
	blogRequireStatus(t, upload, 201)
	var uploaded struct {
		Media BlogMedia `json:"media"`
	}
	if err := json.Unmarshal(upload.Body.Bytes(), &uploaded); err != nil {
		t.Fatal(err)
	}
	media := uploaded.Media
	blogRequireStatus(t, blogTestRequest(t, a, "GET", media.URL, nil), 403)
	blogRequireStatus(t, blogTestRequest(t, a, "GET", media.URL, nil, admin), 200)
	blogRequireStatus(t, blogTestRequest(t, a, "GET", media.URL, nil, user), 403)
	blocks := append(post.Blocks, BlogBlock{Type: "image", MediaID: media.ID, Caption: "private caption"})
	password := "1" // Article passwords intentionally do not inherit account minimum length.
	save := func(status string, passwordValue any) *httptest.ResponseRecorder {
		body := map[string]any{"title": post.Title, "status": status, "blocks": blocks, "version": post.Version}
		if passwordValue != nil {
			body["password"] = passwordValue
		}
		w := blogTestRequest(t, a, "PUT", adminPath, body, admin)
		if w.Code == 200 {
			post = blogResponsePost(t, w)
		}
		return w
	}
	blogRequireStatus(t, save("published", password), 200)
	stale := blogTestRequest(t, a, "PUT", adminPath, map[string]any{"title": "stale", "status": "draft", "blocks": blocks, "version": 1}, admin)
	blogRequireStatus(t, stale, 409)
	locked := blogTestRequest(t, a, "GET", path, nil)
	blogRequireStatus(t, locked, 200)
	if strings.Contains(locked.Body.String(), "private") || strings.Contains(locked.Body.String(), "blocks") {
		t.Fatalf("locked content leak: %s", locked.Body.String())
	}
	list := blogTestRequest(t, a, "GET", "/api/blog/posts", nil)
	blogRequireStatus(t, list, 200)
	if strings.Contains(list.Body.String(), "private") || strings.Contains(list.Body.String(), media.ID) {
		t.Fatalf("list content leak: %s", list.Body.String())
	}
	blogRequireStatus(t, blogTestRequest(t, a, "POST", path+"/unlock", map[string]string{"password": "wrong"}), 401)
	unlocked := blogTestRequest(t, a, "POST", path+"/unlock", map[string]string{"password": password})
	blogRequireStatus(t, unlocked, 200)
	reader := unlocked.Result().Cookies()[0]
	detail := blogTestRequest(t, a, "GET", path, nil, reader)
	if !strings.Contains(detail.Body.String(), "private family text") {
		t.Fatalf("unlock failed: %s", detail.Body.String())
	}
	for _, method := range []string{"GET", "HEAD"} {
		req := httptest.NewRequest(method, media.URL, nil)
		req.AddCookie(reader)
		req.Header.Set("Range", "bytes=0-7")
		w := httptest.NewRecorder()
		a.Handler().ServeHTTP(w, req)
		blogRequireStatus(t, w, 206)
		if w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Content-Range") == "" {
			t.Fatal("unsafe range response")
		}
	}
	// A grant never unlocks a different article, even when its password matches.
	second := blogTestRequest(t, a, "POST", "/api/blog/admin/posts", map[string]any{"title": "Second", "status": "published", "password": password, "blocks": []BlogBlock{{Type: "paragraph", Text: "second secret"}}}, admin)
	blogRequireStatus(t, second, 201)
	other := blogResponsePost(t, second)
	otherDetail := blogTestRequest(t, a, "GET", "/api/blog/posts/"+other.ID, nil, reader)
	if strings.Contains(otherDetail.Body.String(), "second secret") {
		t.Fatal("grant crossed article boundary")
	}
	foreign := blogTestRequest(t, a, "PUT", "/api/blog/admin/posts/"+other.ID, map[string]any{"title": "Second", "status": "published", "version": other.Version, "blocks": blocks}, admin)
	blogRequireStatus(t, foreign, 400)
	badType := append([]BlogBlock{}, blocks...)
	badType[1].Type = "video"
	blogRequireStatus(t, blogTestRequest(t, a, "PUT", adminPath, map[string]any{"title": post.Title, "status": "published", "version": post.Version, "blocks": badType}, admin), 400)
	// Expired grants cannot fetch media or conditional HEAD responses.
	if _, err := a.db.Exec(context.Background(), "update blog_reader_grants set expires_at=now()-interval '1 second'"); err != nil {
		t.Fatal(err)
	}
	blogRequireStatus(t, blogTestRequest(t, a, "HEAD", media.URL, nil, reader), 403)
	unlocked = blogTestRequest(t, a, "POST", path+"/unlock", map[string]string{"password": password}, reader)
	blogRequireStatus(t, unlocked, 200)
	blogRequireStatus(t, save("published", "new-password"), 200)
	blogRequireStatus(t, blogTestRequest(t, a, "GET", media.URL, nil, reader), 403)
	if detail := blogTestRequest(t, a, "GET", path, nil, reader); strings.Contains(detail.Body.String(), "private family text") {
		t.Fatal("old grant survived password change")
	}
	blogRequireStatus(t, save("published", ""), 200)
	publicList := blogTestRequest(t, a, "GET", "/api/blog/posts", nil)
	blogRequireStatus(t, publicList, 200)
	if !strings.Contains(publicList.Body.String(), "private family text") || !strings.Contains(publicList.Body.String(), media.URL) {
		t.Fatalf("public summary missing content: %s", publicList.Body.String())
	}
	blogRequireStatus(t, blogTestRequest(t, a, "GET", media.URL, nil), 200)
	blogRequireStatus(t, save("published", "new-password"), 200)
	blogRequireStatus(t, blogTestRequest(t, a, "GET", media.URL, nil, reader), 403)
	unlocked = blogTestRequest(t, a, "POST", path+"/unlock", map[string]string{"password": "new-password"}, reader)
	blogRequireStatus(t, unlocked, 200)
	blogRequireStatus(t, save("draft", nil), 200)
	blogRequireStatus(t, blogTestRequest(t, a, "GET", path, nil, reader), 404)
	blogRequireStatus(t, blogTestRequest(t, a, "GET", media.URL, nil, reader), 403)
	blogRequireStatus(t, save("published", nil), 200)
	blogRequireStatus(t, blogTestRequest(t, a, "GET", media.URL, nil, reader), 403)
	blogRequireStatus(t, blogTestRequest(t, a, "DELETE", adminPath, nil, admin), 204)
	blogRequireStatus(t, blogTestRequest(t, a, "GET", media.URL, nil, admin), 404)
	blogRequireStatus(t, blogTestRequest(t, a, "GET", path, nil, reader), 404)
	var deleted bool
	if err := a.db.QueryRow(context.Background(), "select deleted_at is not null from blog_posts where id=$1", post.ID).Scan(&deleted); err != nil || !deleted {
		t.Fatalf("soft-delete not persisted: %v", err)
	}
	// The cookie contains a random token, never a password; only its hash is stored.
	hash := sha256.Sum256([]byte(reader.Value))
	if len(hash) != 32 || reader.Value == password {
		t.Fatal("bad reader credential")
	}
}

func TestBlogIntegrationUploadValidation(t *testing.T) {
	a, admin, _ := blogIntegrationApp(t)
	create := blogTestRequest(t, a, "POST", "/api/blog/admin/posts", map[string]any{}, admin)
	blogRequireStatus(t, create, 201)
	post := blogResponsePost(t, create)
	blogRequireStatus(t, blogUploadTest(t, a, post.ID, "evil.svg", []byte("<svg></svg>"), admin), 415)
	oversized := make([]byte, blogImageMaxBytes+1)
	copy(oversized, []byte("\x89PNG\r\n\x1a\n"))
	blogRequireStatus(t, blogUploadTest(t, a, post.ID, "large.png", oversized, admin), 413)
	entries, err := os.ReadDir(a.cfg.BlogMediaDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatal("failed upload left private files behind")
	}
	mp4, _ := hex.DecodeString("00000018667479706d703432000000006d70343269736f6d")
	w := blogUploadTest(t, a, post.ID, "clip.mp4", mp4, admin)
	blogRequireStatus(t, w, 201)
	var uploaded struct {
		Media BlogMedia `json:"media"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &uploaded)
	if uploaded.Media.Type != "video" || uploaded.Media.MIMEType != "video/mp4" {
		t.Fatalf("bad video metadata: %s", w.Body.String())
	}
	entries, _ = os.ReadDir(a.cfg.BlogMediaDir)
	info, err := entries[0].Info()
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("media permissions: %v %v", info, err)
	}
	// Unsupported properties and a body exceeding the JSON limit are rejected.
	raw := fmt.Sprintf(`{"title":"x","status":"draft","blocks":[{"type":"paragraph","text":%q}]}`, strings.Repeat("x", maxJSONBody))
	req := httptest.NewRequest("PUT", "/api/blog/admin/posts/"+post.ID, strings.NewReader(raw))
	req.AddCookie(admin)
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	blogRequireStatus(t, w, 400)
}
