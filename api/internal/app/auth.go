package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{2,31}$`)

const dummyPasswordHash = "$2a$12$7CO1KfXdd7DkKNfO2XvsR.4.8QR7YhAJURJIj6ybOUFgFtDxW2CnS"

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type passwordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func normalizeUsername(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validateUsername(value string) bool {
	return usernamePattern.MatchString(value)
}

func validatePassword(value string) bool {
	length := len([]byte(value))
	return length >= 12 && length <= 72
}

func hashPassword(password string) (string, error) {
	if !validatePassword(password) {
		return "", errors.New("password must contain 12 to 72 bytes")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hash), err
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid login request")
		return
	}
	username := normalizeUsername(input.Username)

	var user User
	var passwordHash string
	err := a.db.QueryRow(r.Context(), `
		select id, username, password_hash, role, is_active, must_change_password, created_at, updated_at
		from users where username = $1
	`, username).Scan(
		&user.ID, &user.Username, &passwordHash, &user.Role, &user.IsActive,
		&user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(input.Password))
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid username or password")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to sign in")
		return
	}
	if !user.IsActive || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid username or password")
		return
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to create session")
		return
	}
	expiresAt := time.Now().Add(a.cfg.SessionTTL)
	if _, err := a.db.Exec(r.Context(), `
		insert into sessions(token_hash, user_id, expires_at) values($1, $2, $3)
	`, tokenHash[:], user.ID, expiresAt); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to create session")
		return
	}
	a.setSessionCookie(w, token, expiresAt)
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func newSessionToken() (string, [32]byte, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", [32]byte{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, sha256.Sum256([]byte(token)), nil
}

func (a *App) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   a.cfg.CookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (a *App) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   a.cfg.CookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(a.cfg.CookieName); err == nil {
		hash := sha256.Sum256([]byte(cookie.Value))
		_, _ = a.db.Exec(r.Context(), "delete from sessions where token_hash = $1", hash[:])
	}
	a.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"user": userFromContext(r.Context())})
}

func (a *App) changePassword(w http.ResponseWriter, r *http.Request) {
	var input passwordRequest
	if err := readJSON(w, r, &input); err != nil || !validatePassword(input.NewPassword) {
		writeError(w, http.StatusBadRequest, "invalid_password", "new password must contain 12 to 72 bytes")
		return
	}
	user := userFromContext(r.Context())
	var currentHash string
	if err := a.db.QueryRow(r.Context(), "select password_hash from users where id = $1", user.ID).Scan(&currentHash); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to change password")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(input.CurrentPassword)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "current password is incorrect")
		return
	}
	newHash, err := hashPassword(input.NewPassword)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_password", err.Error())
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to change password")
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck
	if _, err := tx.Exec(r.Context(), `
		update users set password_hash = $1, must_change_password = false, updated_at = now() where id = $2
	`, newHash, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to change password")
		return
	}
	if _, err := tx.Exec(r.Context(), "delete from sessions where user_id = $1", user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to change password")
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to change password")
		return
	}
	a.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "password_changed"})
}

func (a *App) requireAuth(next http.Handler, allowPasswordChange bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(a.cfg.CookieName)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "authentication_required", "authentication required")
			return
		}
		hash := sha256.Sum256([]byte(cookie.Value))
		var user User
		err = a.db.QueryRow(r.Context(), `
			select u.id, u.username, u.role, u.is_active, u.must_change_password, u.created_at, u.updated_at
			from sessions s join users u on u.id = s.user_id
			where s.token_hash = $1 and s.expires_at > now() and u.is_active = true
		`, hash[:]).Scan(
			&user.ID, &user.Username, &user.Role, &user.IsActive,
			&user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			a.clearSessionCookie(w)
			writeError(w, http.StatusUnauthorized, "authentication_required", "authentication required")
			return
		}
		if user.MustChangePassword && !allowPasswordChange {
			writeError(w, http.StatusForbidden, "password_change_required", "password change required")
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *App) requireAdmin(next http.Handler) http.Handler {
	return a.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if userFromContext(r.Context()).Role != "admin" {
			writeError(w, http.StatusForbidden, "admin_required", "administrator access required")
			return
		}
		next.ServeHTTP(w, r)
	}), false)
}

func userFromContext(ctx context.Context) User {
	user, _ := ctx.Value(userContextKey).(User)
	return user
}
