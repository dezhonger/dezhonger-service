package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg Config
	db  *pgxpool.Pool
	mux *http.ServeMux
}

func New(ctx context.Context, cfg Config) (*App, error) {
	db, err := openDatabase(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	application := &App{cfg: cfg, db: db, mux: http.NewServeMux()}
	application.routes()
	return application, nil
}

func (a *App) Close() {
	a.db.Close()
}

func (a *App) Handler() http.Handler {
	return a.recoverPanic(a.securityHeaders(a.validateOrigin(a.logRequests(a.mux))))
}

func (a *App) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:              a.cfg.ListenAddr,
		Handler:           a.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errorChannel := make(chan error, 1)
	go func() {
		slog.Info("api listening", "address", a.cfg.ListenAddr)
		errorChannel <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errorChannel:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (a *App) routes() {
	a.mux.HandleFunc("GET /api/healthz", a.healthz)
	a.mux.HandleFunc("POST /api/auth/login", a.login)
	a.mux.Handle("POST /api/auth/logout", a.requireAuth(http.HandlerFunc(a.logout), true))
	a.mux.Handle("GET /api/auth/me", a.requireAuth(http.HandlerFunc(a.me), true))
	a.mux.Handle("PUT /api/account/password", a.requireAuth(http.HandlerFunc(a.changePassword), true))
	a.mux.Handle("GET /api/notes", a.requireAuth(http.HandlerFunc(a.listNotes), false))
	a.mux.Handle("PUT /api/notes/{id}", a.requireAuth(http.HandlerFunc(a.putNote), false))
	a.mux.Handle("DELETE /api/notes/{id}", a.requireAuth(http.HandlerFunc(a.deleteNote), false))
	a.mux.Handle("GET /api/admin/users", a.requireAdmin(http.HandlerFunc(a.listUsers)))
	a.mux.Handle("POST /api/admin/users", a.requireAdmin(http.HandlerFunc(a.createUser)))
	a.mux.Handle("PATCH /api/admin/users/{id}", a.requireAdmin(http.HandlerFunc(a.updateUser)))
}

func (a *App) healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.db.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database_unavailable", "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func (a *App) validateOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			origin := r.Header.Get("Origin")
			if origin != "" && a.cfg.PublicOrigin != "" && origin != a.cfg.PublicOrigin {
				writeError(w, http.StatusForbidden, "invalid_origin", "request origin is not allowed")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started))
	})
}

func (a *App) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic", "error", recovered, "stack", string(debug.Stack()))
				writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
