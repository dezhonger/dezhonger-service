package app

import (
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const blogImageMaxBytes int64 = 20 << 20
const blogVideoMaxBytes int64 = 200 << 20
const blogUploadMaxBytes int64 = 210 << 20
const blogMediaDeadline = 15 * time.Minute

type BlogMedia struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	MIMEType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

func blogMediaFormat(header []byte) (string, string, int64) {
	mimeType := http.DetectContentType(header)
	switch mimeType {
	case "image/jpeg":
		return mimeType, ".jpg", blogImageMaxBytes
	case "image/png":
		return mimeType, ".png", blogImageMaxBytes
	case "image/gif":
		return mimeType, ".gif", blogImageMaxBytes
	case "image/webp":
		return mimeType, ".webp", blogImageMaxBytes
	case "video/mp4":
		return mimeType, ".mp4", blogVideoMaxBytes
	case "video/webm":
		return mimeType, ".webm", blogVideoMaxBytes
	default:
		return "", "", 0
	}
}

func blogUploadName(raw string) string {
	name := filepath.Base(strings.ReplaceAll(raw, "\\", "/"))
	if name == "." || name == "/" || name == "" {
		name = "media"
	}
	runes := []rune(name)
	if len(runes) > 200 {
		name = string(runes[:200])
	}
	return name
}

func (a *App) uploadBlogMedia(w http.ResponseWriter, r *http.Request) {
	controller := http.NewResponseController(w)
	_ = controller.SetReadDeadline(time.Now().Add(blogMediaDeadline))
	_ = controller.SetWriteDeadline(time.Now().Add(blogMediaDeadline))
	postID, ok := blogPathID(w, r)
	if !ok {
		return
	}
	var exists bool
	if err := a.db.QueryRow(r.Context(), "select exists(select 1 from blog_posts where id=$1 and deleted_at is null)", postID).Scan(&exists); err != nil {
		blogInternalError(w)
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "post_not_found", "post not found")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, blogUploadMaxBytes)
	reader, err := r.MultipartReader()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_upload", "multipart file upload required")
		return
	}
	part, err := reader.NextPart()
	if err != nil || part.FormName() != "file" || part.FileName() == "" {
		writeError(w, http.StatusBadRequest, "invalid_upload", "upload exactly one file field")
		return
	}
	defer part.Close()
	header := make([]byte, 512)
	n, err := io.ReadFull(part, header)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		writeError(w, http.StatusBadRequest, "invalid_upload", "unable to read uploaded file")
		return
	}
	header = header[:n]
	mimeType, extension, maxBytes := blogMediaFormat(header)
	if maxBytes == 0 {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_media", "supported files: JPEG, PNG, GIF, WebP, MP4 and WebM")
		return
	}
	id := uuid.NewString()
	storageName := id + extension
	if err := os.MkdirAll(a.cfg.BlogMediaDir, 0700); err != nil {
		blogInternalError(w)
		return
	}
	path := filepath.Join(a.cfg.BlogMediaDir, storageName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		blogInternalError(w)
		return
	}
	keep := false
	defer func() {
		_ = file.Close()
		if !keep {
			_ = os.Remove(path)
		}
	}()
	if _, err := file.Write(header); err != nil {
		blogInternalError(w)
		return
	}
	written, err := io.Copy(file, io.LimitReader(part, maxBytes+1-int64(n)))
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "file_too_large", "file is too large")
		} else {
			writeError(w, http.StatusBadRequest, "invalid_upload", "unable to read uploaded file")
		}
		return
	}
	size := int64(n) + written
	if size > maxBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "file_too_large", "image limit is 20 MiB; video limit is 200 MiB")
		return
	}
	if _, err := reader.NextPart(); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_upload", "upload exactly one file field")
		return
	}
	if err := file.Close(); err != nil {
		blogInternalError(w)
		return
	}
	name := blogUploadName(part.FileName())
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		blogInternalError(w)
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck
	var currentID string
	err = tx.QueryRow(r.Context(), "select id from blog_posts where id=$1 and deleted_at is null for update", postID).Scan(&currentID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "post_not_found", "post not found")
		return
	}
	if err != nil {
		blogInternalError(w)
		return
	}
	if _, err := tx.Exec(r.Context(), `insert into blog_media(id,post_id,storage_name,original_name,mime_type,size) values($1,$2,$3,$4,$5,$6)`, id, postID, storageName, name, mimeType, size); err != nil {
		blogInternalError(w)
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		blogInternalError(w)
		return
	}
	keep = true
	mediaType := "image"
	if strings.HasPrefix(mimeType, "video/") {
		mediaType = "video"
	}
	writeJSON(w, http.StatusCreated, map[string]any{"media": BlogMedia{ID: id, URL: "/api/blog/media/" + id, MIMEType: mimeType, Size: size, Name: name, Type: mediaType}})
}

func (a *App) getBlogMedia(w http.ResponseWriter, r *http.Request) {
	id, ok := blogPathID(w, r)
	if !ok {
		return
	}
	var postID, storageName, name, mimeType string
	var createdAt time.Time
	err := a.db.QueryRow(r.Context(), `select m.post_id,m.storage_name,m.original_name,m.mime_type,m.created_at from blog_media m join blog_posts p on p.id=m.post_id where m.id=$1 and p.deleted_at is null`, id).Scan(&postID, &storageName, &name, &mimeType, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "media_not_found", "media not found")
		return
	}
	if err != nil {
		blogInternalError(w)
		return
	}
	admin, err := a.blogRequestIsAdmin(r)
	if err != nil {
		blogInternalError(w)
		return
	}
	post, err := scanBlogPost(a.db.QueryRow(r.Context(), "select "+blogPostColumns+" from blog_posts where id=$1 and deleted_at is null", postID))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "media_not_found", "media not found")
		return
	}
	if err != nil {
		blogInternalError(w)
		return
	}
	allowed, err := a.canReadBlogPost(r, post, admin)
	if err != nil {
		blogInternalError(w)
		return
	}
	if !admin && allowed {
		allowed = false
		for _, block := range post.Blocks {
			if block.MediaID == id {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		writeError(w, http.StatusForbidden, "article_access_required", "unlock the article before accessing its media")
		return
	}
	if filepath.Base(storageName) != storageName {
		blogInternalError(w)
		return
	}
	file, err := os.Open(filepath.Join(a.cfg.BlogMediaDir, storageName))
	if errors.Is(err, os.ErrNotExist) {
		writeError(w, http.StatusNotFound, "media_not_found", "media not found")
		return
	}
	if err != nil {
		blogInternalError(w)
		return
	}
	defer file.Close()
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(blogMediaDeadline))
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": name}))
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	http.ServeContent(w, r, name, createdAt, file)
}
