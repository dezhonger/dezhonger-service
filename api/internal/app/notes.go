package app

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type noteRequest struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
	Category string   `json:"category"`
	Status   string   `json:"status"`
	IsPinned bool     `json:"is_pinned"`
}

var validNoteStatuses = map[string]bool{
	"inbox": true, "todo": true, "doing": true, "done": true, "archived": true,
}

func normalizeTags(tags []string) ([]string, error) {
	if len(tags) > 20 {
		return nil, errors.New("a note may contain at most 20 tags")
	}
	result := make([]string, 0, len(tags))
	seen := make(map[string]bool)
	for _, raw := range tags {
		tag := strings.TrimSpace(raw)
		if tag == "" {
			continue
		}
		if len([]rune(tag)) > 32 {
			return nil, errors.New("tags may contain at most 32 characters")
		}
		key := strings.ToLower(tag)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, tag)
	}
	return result, nil
}

func validateNote(input *noteRequest) error {
	if len([]rune(input.Title)) > 200 {
		return errors.New("title is too long")
	}
	if len([]rune(input.Content)) > 2_000_000 {
		return errors.New("content is too long")
	}
	input.Category = strings.TrimSpace(input.Category)
	if len([]rune(input.Category)) > 50 {
		return errors.New("category is too long")
	}
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.Status == "" {
		input.Status = "inbox"
	}
	if !validNoteStatuses[input.Status] {
		return errors.New("invalid note status")
	}
	tags, err := normalizeTags(input.Tags)
	if err != nil {
		return err
	}
	input.Tags = tags
	return nil
}

func scanNote(row pgx.Row) (Note, error) {
	var note Note
	err := row.Scan(
		&note.ID, &note.UserID, &note.Title, &note.Content, &note.Tags,
		&note.Category, &note.Status, &note.IsPinned, &note.CreatedAt, &note.UpdatedAt,
	)
	if note.Tags == nil {
		note.Tags = []string{}
	}
	return note, err
}

func (a *App) listNotes(w http.ResponseWriter, r *http.Request) {
	user := userFromContext(r.Context())
	rows, err := a.db.Query(r.Context(), `
		select id, user_id, title, content, tags, category, status, is_pinned, created_at, updated_at
		from notes where user_id = $1 order by is_pinned desc, updated_at desc
	`, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to list notes")
		return
	}
	defer rows.Close()

	notes := make([]Note, 0)
	for rows.Next() {
		var note Note
		if err := rows.Scan(
			&note.ID, &note.UserID, &note.Title, &note.Content, &note.Tags,
			&note.Category, &note.Status, &note.IsPinned, &note.CreatedAt, &note.UpdatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "unable to list notes")
			return
		}
		if note.Tags == nil {
			note.Tags = []string{}
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to list notes")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"notes": notes})
}

func (a *App) putNote(w http.ResponseWriter, r *http.Request) {
	noteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_note", "invalid note id")
		return
	}
	var input noteRequest
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid note request")
		return
	}
	if err := validateNote(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_note", err.Error())
		return
	}
	user := userFromContext(r.Context())
	note, err := scanNote(a.db.QueryRow(r.Context(), `
		insert into notes(id, user_id, title, content, tags, category, status, is_pinned)
		values($1, $2, $3, $4, $5, $6, $7, $8)
		on conflict (id) do update
		set title = excluded.title,
		    content = excluded.content,
		    tags = excluded.tags,
		    category = excluded.category,
		    status = excluded.status,
		    is_pinned = excluded.is_pinned,
		    updated_at = now()
		where notes.user_id = excluded.user_id
		returning id, user_id, title, content, tags, category, status, is_pinned, created_at, updated_at
	`, noteID, user.ID, input.Title, input.Content, input.Tags, input.Category, input.Status, input.IsPinned))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusConflict, "note_conflict", "note id belongs to another user")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to save note")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"note": note})
}

func (a *App) deleteNote(w http.ResponseWriter, r *http.Request) {
	noteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_note", "invalid note id")
		return
	}
	user := userFromContext(r.Context())
	result, err := a.db.Exec(r.Context(), "delete from notes where id = $1 and user_id = $2", noteID, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to delete note")
		return
	}
	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "note_not_found", "note not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
