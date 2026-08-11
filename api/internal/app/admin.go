package app

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type createUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type updateUserRequest struct {
	Role     *string `json:"role"`
	IsActive *bool   `json:"is_active"`
	Password *string `json:"password"`
}

func (a *App) CreateAdmin(ctx context.Context, username, password string) (User, error) {
	username = normalizeUsername(username)
	if !validateUsername(username) {
		return User{}, errors.New("username must contain 3 to 32 lowercase letters, digits, dots, underscores, or hyphens")
	}
	passwordHash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}
	var user User
	err = a.db.QueryRow(ctx, `
		insert into users(username, password_hash, role, must_change_password)
		values($1, $2, 'admin', true)
		returning id, username, role, is_active, must_change_password, created_at, updated_at
	`, username, passwordHash).Scan(
		&user.ID, &user.Username, &user.Role, &user.IsActive,
		&user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (a *App) listUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), `
		select id, username, role, is_active, must_change_password, created_at, updated_at
		from users order by created_at asc
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to list users")
		return
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Role, &user.IsActive,
			&user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "unable to list users")
			return
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to list users")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var input createUserRequest
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid user request")
		return
	}
	input.Username = normalizeUsername(input.Username)
	input.Role = strings.ToLower(strings.TrimSpace(input.Role))
	if input.Role == "" {
		input.Role = "user"
	}
	if !validateUsername(input.Username) || (input.Role != "admin" && input.Role != "user") {
		writeError(w, http.StatusBadRequest, "invalid_user", "invalid username or role")
		return
	}
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_password", err.Error())
		return
	}

	var user User
	err = a.db.QueryRow(r.Context(), `
		insert into users(username, password_hash, role, must_change_password)
		values($1, $2, $3, true)
		returning id, username, role, is_active, must_change_password, created_at, updated_at
	`, input.Username, passwordHash, input.Role).Scan(
		&user.ID, &user.Username, &user.Role, &user.IsActive,
		&user.MustChangePassword, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			writeError(w, http.StatusConflict, "username_exists", "username already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to create user")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": user})
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	targetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_user", "invalid user id")
		return
	}
	var input updateUserRequest
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid user update")
		return
	}
	if input.Role == nil && input.IsActive == nil && input.Password == nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "no user changes supplied")
		return
	}

	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to update user")
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck

	var target User
	var existingHash string
	err = tx.QueryRow(r.Context(), `
		select id, username, password_hash, role, is_active, must_change_password, created_at, updated_at
		from users where id = $1 for update
	`, targetID).Scan(
		&target.ID, &target.Username, &existingHash, &target.Role, &target.IsActive,
		&target.MustChangePassword, &target.CreatedAt, &target.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "user_not_found", "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to update user")
		return
	}

	nextRole := target.Role
	if input.Role != nil {
		nextRole = strings.ToLower(strings.TrimSpace(*input.Role))
		if nextRole != "admin" && nextRole != "user" {
			writeError(w, http.StatusBadRequest, "invalid_role", "role must be admin or user")
			return
		}
	}
	nextActive := target.IsActive
	if input.IsActive != nil {
		nextActive = *input.IsActive
	}
	currentUser := userFromContext(r.Context())
	if target.ID == currentUser.ID && (!nextActive || nextRole != "admin") {
		writeError(w, http.StatusConflict, "cannot_lock_self", "you cannot disable or demote your own administrator account")
		return
	}
	if target.Role == "admin" && target.IsActive && (!nextActive || nextRole != "admin") {
		var otherAdmins int
		if err := tx.QueryRow(r.Context(), `
			select count(*) from users where role = 'admin' and is_active = true and id <> $1
		`, targetID).Scan(&otherAdmins); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "unable to update user")
			return
		}
		if otherAdmins == 0 {
			writeError(w, http.StatusConflict, "last_admin", "the last active administrator cannot be disabled or demoted")
			return
		}
	}

	nextHash := existingHash
	mustChange := target.MustChangePassword
	passwordChanged := false
	if input.Password != nil {
		nextHash, err = hashPassword(*input.Password)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_password", err.Error())
			return
		}
		mustChange = true
		passwordChanged = true
	}

	err = tx.QueryRow(r.Context(), `
		update users
		set role = $1, is_active = $2, password_hash = $3,
		    must_change_password = $4, updated_at = now()
		where id = $5
		returning id, username, role, is_active, must_change_password, created_at, updated_at
	`, nextRole, nextActive, nextHash, mustChange, targetID).Scan(
		&target.ID, &target.Username, &target.Role, &target.IsActive,
		&target.MustChangePassword, &target.CreatedAt, &target.UpdatedAt,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to update user")
		return
	}
	if passwordChanged || !nextActive {
		if _, err := tx.Exec(r.Context(), "delete from sessions where user_id = $1", targetID); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "unable to update user")
			return
		}
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to update user")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": target})
}
