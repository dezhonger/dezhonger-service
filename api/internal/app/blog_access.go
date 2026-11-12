package app

import (
	"crypto/sha256"
	"encoding/base64"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const blogReaderTTL = 24 * time.Hour
const blogAttemptWindow = 10 * time.Minute
const blogAttemptLimit = 10
const blogAttemptMaxKeys = 4096

type blogAttempts struct {
	count int
	until time.Time
}
type blogAttemptLimiter struct {
	mu      sync.Mutex
	entries map[string]blogAttempts
}

// Bound memory and fail closed for new clients when all slots are occupied.
func (l *blogAttemptLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.entries == nil {
		l.entries = make(map[string]blogAttempts)
	}
	value, ok := l.entries[key]
	if !ok || !now.Before(value.until) {
		for k, v := range l.entries {
			if !now.Before(v.until) {
				delete(l.entries, k)
			}
		}
		if len(l.entries) >= blogAttemptMaxKeys {
			return false
		}
		value = blogAttempts{until: now.Add(blogAttemptWindow)}
	}
	if value.count >= blogAttemptLimit {
		return false
	}
	value.count++
	l.entries[key] = value
	return true
}

func blogClientIP(r *http.Request) string {
	if ip := net.ParseIP(r.Header.Get("X-Real-IP")); ip != nil {
		return ip.String()
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func (a *App) blogReaderCookieName() string {
	if a.cfg.CookieSecure {
		return "__Host-dz_blog_reader"
	}
	return "dz_blog_reader"
}

func validBlogReaderToken(value string) bool {
	if len(value) != 43 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(raw) == 32
}

func (a *App) blogReaderToken(r *http.Request) string {
	cookie, err := r.Cookie(a.blogReaderCookieName())
	if err != nil || !validBlogReaderToken(cookie.Value) {
		return ""
	}
	return cookie.Value
}

func (a *App) setBlogReaderCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: a.blogReaderCookieName(), Value: token, Path: "/", HttpOnly: true, Secure: a.cfg.CookieSecure, SameSite: http.SameSiteStrictMode, MaxAge: int(blogReaderTTL.Seconds()), Expires: time.Now().Add(blogReaderTTL)})
}

func (a *App) blogRequestIsAdmin(r *http.Request) (bool, error) {
	cookie, err := r.Cookie(a.cfg.CookieName)
	if err != nil || cookie.Value == "" {
		return false, nil
	}
	hash := sha256.Sum256([]byte(cookie.Value))
	var allowed bool
	err = a.db.QueryRow(r.Context(), `select exists(select 1 from sessions s join users u on u.id=s.user_id where s.token_hash=$1 and s.expires_at>now() and u.is_active=true and u.must_change_password=false and u.role='admin')`, hash[:]).Scan(&allowed)
	return allowed, err
}

func (a *App) canReadBlogPost(r *http.Request, post BlogPost, admin bool) (bool, error) {
	if admin {
		return true, nil
	}
	if post.Status != "published" {
		return false, nil
	}
	if !post.Protected {
		return true, nil
	}
	token := a.blogReaderToken(r)
	if token == "" {
		return false, nil
	}
	hash := sha256.Sum256([]byte(token))
	var allowed bool
	err := a.db.QueryRow(r.Context(), `select exists(select 1 from blog_reader_grants g join blog_posts p on p.id=g.post_id where g.token_hash=$1 and g.post_id=$2 and g.access_version=p.access_version and g.access_version=$3 and g.expires_at>now() and p.status='published' and p.deleted_at is null)`, hash[:], post.ID, post.accessVersion).Scan(&allowed)
	return allowed, err
}

func (a *App) unlockBlogPost(w http.ResponseWriter, r *http.Request) {
	if !a.blogLimiter.allow(blogClientIP(r), time.Now()) {
		w.Header().Set("Retry-After", strconv.Itoa(int(blogAttemptWindow.Seconds())))
		writeError(w, http.StatusTooManyRequests, "rate_limited", "too many password attempts; try again later")
		return
	}
	var input struct {
		Password string `json:"password"`
	}
	if err := readJSON(w, r, &input); err != nil || len([]byte(input.Password)) > 72 {
		writeError(w, http.StatusBadRequest, "invalid_password", "invalid article password")
		return
	}
	post, ok := a.loadBlogPost(w, r, false)
	if !ok {
		return
	}
	if !post.Protected {
		writeJSON(w, http.StatusOK, map[string]string{"status": "unlocked"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(post.passwordHash), []byte(input.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid_password", "incorrect article password")
		return
	}
	token := a.blogReaderToken(r)
	if token == "" {
		var err error
		token, _, err = newSessionToken()
		if err != nil {
			blogInternalError(w)
			return
		}
	}
	hash := sha256.Sum256([]byte(token))
	// Bind the grant to the version checked by bcrypt. A concurrent password change,
	// unpublish or delete cannot mint a grant for the replacement version.
	result, err := a.db.Exec(r.Context(), `insert into blog_reader_grants(token_hash,post_id,access_version,expires_at) select $1,id,access_version,$4 from blog_posts where id=$2 and access_version=$3 and status='published' and deleted_at is null on conflict(token_hash,post_id) do update set access_version=excluded.access_version,expires_at=excluded.expires_at`, hash[:], post.ID, post.accessVersion, time.Now().Add(blogReaderTTL))
	if err != nil {
		blogInternalError(w)
		return
	}
	if result.RowsAffected() == 0 {
		writeError(w, http.StatusConflict, "post_changed", "article access changed; reload and try again")
		return
	}
	// Expired grants carry no usable authority; remove them opportunistically.
	_, _ = a.db.Exec(r.Context(), "delete from blog_reader_grants where expires_at<now()")
	a.setBlogReaderCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]string{"status": "unlocked"})
}
