package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type BlogBlock struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	MediaID string `json:"media_id,omitempty"`
	Caption string `json:"caption,omitempty"`
}

type BlogPost struct {
	ID            string      `json:"id"`
	Title         string      `json:"title"`
	Blocks        []BlogBlock `json:"blocks"`
	Status        string      `json:"status"`
	Protected     bool        `json:"protected"`
	Version       int64       `json:"version"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	PublishedAt   *time.Time  `json:"published_at"`
	passwordHash  string
	accessVersion int64
}

type blogPostRequest struct {
	Title    string      `json:"title"`
	Blocks   []BlogBlock `json:"blocks"`
	Status   string      `json:"status"`
	Password *string     `json:"password"`
	Version  *int64      `json:"version"`
}

const blogPostColumns = `id, title, blocks, status, password_hash, access_version, version, created_at, updated_at, published_at`

func (a *App) blogRoutes() {
	a.mux.HandleFunc("GET /api/blog/posts", a.listBlogPosts)
	a.mux.HandleFunc("GET /api/blog/posts/{id}", a.getBlogPost)
	a.mux.HandleFunc("POST /api/blog/posts/{id}/unlock", a.unlockBlogPost)
	a.mux.HandleFunc("GET /api/blog/media/{id}", a.getBlogMedia)
	a.mux.Handle("GET /api/blog/admin/posts", a.requireAdmin(http.HandlerFunc(a.listAdminBlogPosts)))
	a.mux.Handle("POST /api/blog/admin/posts", a.requireAdmin(http.HandlerFunc(a.createBlogPost)))
	a.mux.Handle("GET /api/blog/admin/posts/{id}", a.requireAdmin(http.HandlerFunc(a.getAdminBlogPost)))
	a.mux.Handle("PUT /api/blog/admin/posts/{id}", a.requireAdmin(http.HandlerFunc(a.updateBlogPost)))
	a.mux.Handle("DELETE /api/blog/admin/posts/{id}", a.requireAdmin(http.HandlerFunc(a.deleteBlogPost)))
	a.mux.Handle("POST /api/blog/admin/posts/{id}/media", a.requireAdmin(http.HandlerFunc(a.uploadBlogMedia)))
}

func scanBlogPost(row pgx.Row) (BlogPost, error) {
	var p BlogPost
	var raw []byte
	err := row.Scan(&p.ID, &p.Title, &raw, &p.Status, &p.passwordHash, &p.accessVersion, &p.Version, &p.CreatedAt, &p.UpdatedAt, &p.PublishedAt)
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(raw, &p.Blocks); err != nil {
		return p, err
	}
	if p.Blocks == nil {
		p.Blocks = []BlogBlock{}
	}
	p.Protected = p.passwordHash != ""
	return p, nil
}

func validateBlogPost(input *blogPostRequest) error {
	input.Title = strings.TrimSpace(input.Title)
	if utf8.RuneCountInString(input.Title) > 200 {
		return errors.New("title may contain at most 200 characters")
	}
	if input.Status == "" {
		input.Status = "draft"
	}
	if input.Status != "draft" && input.Status != "published" {
		return errors.New("status must be draft or published")
	}
	if len(input.Blocks) >= 200 {
		return errors.New("a post must contain fewer than 200 blocks")
	}
	if input.Blocks == nil {
		input.Blocks = []BlogBlock{}
	}
	for i := range input.Blocks {
		b := &input.Blocks[i]
		switch b.Type {
		case "paragraph", "heading", "quote", "list":
			b.Text = strings.TrimSpace(b.Text)
			if b.Text == "" || b.MediaID != "" || b.Caption != "" {
				return errors.New("text blocks require text and cannot contain media or captions")
			}
		case "image", "video":
			parsed, err := uuid.Parse(b.MediaID)
			if err != nil || b.Text != "" || utf8.RuneCountInString(b.Caption) > 2000 {
				return errors.New("media blocks require a media id and a caption of at most 2000 characters")
			}
			b.MediaID = parsed.String()
		default:
			return errors.New("unsupported block type")
		}
	}
	if input.Status == "published" && (input.Title == "" || len(input.Blocks) == 0) {
		return errors.New("published posts require a title and at least one block")
	}
	if input.Password != nil && len([]byte(*input.Password)) > 72 {
		return errors.New("post password may contain at most 72 bytes")
	}
	return nil
}

func blogPagination(r *http.Request) (int, int, error) {
	page, limit := 1, 12
	var err error
	if value := r.URL.Query().Get("page"); value != "" {
		page, err = strconv.Atoi(value)
		if err != nil {
			return 0, 0, errors.New("invalid page")
		}
	}
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil {
			return 0, 0, errors.New("invalid limit")
		}
	}
	if page < 1 || page > 1000000 || limit < 1 || limit > 100 {
		return 0, 0, errors.New("page must be 1 to 1000000 and limit must be 1 to 100")
	}
	return page, limit, nil
}

func (a *App) listBlogPosts(w http.ResponseWriter, r *http.Request) { a.listBlogPostsPage(w, r, false) }
func (a *App) listAdminBlogPosts(w http.ResponseWriter, r *http.Request) {
	a.listBlogPostsPage(w, r, true)
}

func (a *App) listBlogPostsPage(w http.ResponseWriter, r *http.Request, admin bool) {
	page, limit, err := blogPagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}
	where := "deleted_at is null"
	if admin {
		switch r.URL.Query().Get("status") {
		case "":
		case "draft":
			where += " and status = 'draft'"
		case "published":
			where += " and status = 'published'"
		default:
			writeError(w, http.StatusBadRequest, "invalid_status", "status must be draft or published")
			return
		}
	}
	order := "published_at desc, id desc"
	if !admin {
		where += " and status = 'published'"
	} else {
		order = "updated_at desc, id desc"
	}
	var total int64
	if err := a.db.QueryRow(r.Context(), "select count(*) from blog_posts where "+where).Scan(&total); err != nil {
		blogInternalError(w)
		return
	}
	if !admin {
		a.listPublicBlogSummaries(w, r, page, limit, total)
		return
	}
	rows, err := a.db.Query(r.Context(), "select "+blogPostColumns+" from blog_posts where "+where+" order by "+order+" limit $1 offset $2", limit, (page-1)*limit)
	if err != nil {
		blogInternalError(w)
		return
	}
	defer rows.Close()
	posts := make([]any, 0)
	for rows.Next() {
		post, err := scanBlogPost(rows)
		if err != nil {
			blogInternalError(w)
			return
		}
		if admin {
			posts = append(posts, post)
		} else {
			posts = append(posts, publicBlogSummary(post))
		}
	}
	if rows.Err() != nil {
		blogInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"posts": posts, "total": total, "page": page, "limit": limit})
}

// Fetch bounded public summaries without loading entire article bodies or any
// protected content into the list response path.
func (a *App) listPublicBlogSummaries(w http.ResponseWriter, r *http.Request, page, limit int, total int64) {
	rows, err := a.db.Query(r.Context(), `select id,title,password_hash<>'',published_at,updated_at,
 case when password_hash='' then coalesce((select left(b->>'text',161) from jsonb_array_elements(blocks) b where coalesce(b->>'text','')<>'' limit 1),'') else '' end,
 case when password_hash='' then coalesce((select b->>'media_id' from jsonb_array_elements(blocks) b where b->>'type'='image' limit 1),'') else '' end
 from blog_posts where status='published' and deleted_at is null order by published_at desc,id desc limit $1 offset $2`, limit, (page-1)*limit)
	if err != nil {
		blogInternalError(w)
		return
	}
	defer rows.Close()
	posts := make([]any, 0)
	for rows.Next() {
		var post BlogPost
		var excerpt, coverID string
		if err := rows.Scan(&post.ID, &post.Title, &post.Protected, &post.PublishedAt, &post.UpdatedAt, &excerpt, &coverID); err != nil {
			blogInternalError(w)
			return
		}
		if excerpt != "" {
			post.Blocks = append(post.Blocks, BlogBlock{Type: "paragraph", Text: excerpt})
		}
		if coverID != "" {
			post.Blocks = append(post.Blocks, BlogBlock{Type: "image", MediaID: coverID})
		}
		posts = append(posts, publicBlogSummary(post))
	}
	if rows.Err() != nil {
		blogInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"posts": posts, "total": total, "page": page, "limit": limit})
}

func publicBlogSummary(post BlogPost) map[string]any {
	result := map[string]any{"id": post.ID, "title": post.Title, "excerpt": "", "protected": post.Protected, "published_at": post.PublishedAt, "updated_at": post.UpdatedAt}
	if post.Protected {
		return result
	}
	for _, b := range post.Blocks {
		if b.Text != "" && result["excerpt"] == "" {
			text := []rune(b.Text)
			if len(text) > 160 {
				text = append(text[:160], '…')
			}
			result["excerpt"] = string(text)
		}
		if b.Type == "image" && result["cover_url"] == nil {
			result["cover_url"] = "/api/blog/media/" + b.MediaID
		}
	}
	return result
}

func publicBlogDetail(post BlogPost, allowed bool) map[string]any {
	result := map[string]any{"id": post.ID, "title": post.Title, "protected": post.Protected, "published_at": post.PublishedAt, "updated_at": post.UpdatedAt}
	if allowed {
		result["blocks"] = post.Blocks
	}
	return map[string]any{"post": result, "locked": !allowed}
}

func blogPathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "post_not_found", "post or media not found")
		return "", false
	}
	return id.String(), true
}

func (a *App) loadBlogPost(w http.ResponseWriter, r *http.Request, admin bool) (BlogPost, bool) {
	id, ok := blogPathID(w, r)
	if !ok {
		return BlogPost{}, false
	}
	where := "id = $1 and deleted_at is null"
	if !admin {
		where += " and status = 'published'"
	}
	post, err := scanBlogPost(a.db.QueryRow(r.Context(), "select "+blogPostColumns+" from blog_posts where "+where, id))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "post_not_found", "post not found")
		return post, false
	}
	if err != nil {
		blogInternalError(w)
		return post, false
	}
	return post, true
}

func (a *App) getBlogPost(w http.ResponseWriter, r *http.Request) {
	post, ok := a.loadBlogPost(w, r, false)
	if !ok {
		return
	}
	allowed, err := a.canReadBlogPost(r, post, false)
	if err != nil {
		blogInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, publicBlogDetail(post, allowed))
}

func (a *App) getAdminBlogPost(w http.ResponseWriter, r *http.Request) {
	post, ok := a.loadBlogPost(w, r, true)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"post": post})
}

func readBlogPostRequest(w http.ResponseWriter, r *http.Request) (blogPostRequest, bool) {
	var input blogPostRequest
	if err := readJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_post", "invalid post body or unsupported fields")
		return input, false
	}
	if err := validateBlogPost(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_post", err.Error())
		return input, false
	}
	return input, true
}

func blogPasswordHash(password *string, previous string) (string, error) {
	if password == nil {
		return previous, nil
	}
	if *password == "" {
		return "", nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
	return string(hash), err
}

var errInvalidBlogMedia = errors.New("media must belong to this post and match its block type")

func validateBlogMediaReferences(r *http.Request, tx pgx.Tx, postID string, blocks []BlogBlock) error {
	for _, b := range blocks {
		if b.MediaID == "" {
			continue
		}
		var mime string
		err := tx.QueryRow(r.Context(), "select mime_type from blog_media where id = $1 and post_id = $2", b.MediaID, postID).Scan(&mime)
		if errors.Is(err, pgx.ErrNoRows) {
			return errInvalidBlogMedia
		}
		if err != nil {
			return err
		}
		if (b.Type == "image" && !strings.HasPrefix(mime, "image/")) || (b.Type == "video" && !strings.HasPrefix(mime, "video/")) {
			return errInvalidBlogMedia
		}
	}
	return nil
}

func (a *App) createBlogPost(w http.ResponseWriter, r *http.Request) {
	input, ok := readBlogPostRequest(w, r)
	if !ok {
		return
	}
	for _, b := range input.Blocks {
		if b.MediaID != "" {
			writeError(w, http.StatusBadRequest, "invalid_media", "create a draft before uploading media")
			return
		}
	}
	hash, err := blogPasswordHash(input.Password, "")
	if err != nil {
		blogInternalError(w)
		return
	}
	raw, _ := json.Marshal(input.Blocks)
	post, err := scanBlogPost(a.db.QueryRow(r.Context(), `insert into blog_posts(author_id,title,blocks,status,password_hash,published_at) values($1,$2,$3,$4,$5,case when $4 = 'published' then now() else null end) returning `+blogPostColumns, userFromContext(r.Context()).ID, input.Title, raw, input.Status, hash))
	if err != nil {
		blogInternalError(w)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"post": post})
}

func (a *App) updateBlogPost(w http.ResponseWriter, r *http.Request) {
	id, ok := blogPathID(w, r)
	if !ok {
		return
	}
	input, ok := readBlogPostRequest(w, r)
	if !ok {
		return
	}
	if input.Version == nil || *input.Version < 1 {
		writeError(w, http.StatusBadRequest, "version_required", "post version is required")
		return
	}
	// Password hashing precedes the transaction to avoid holding a row lock during bcrypt.
	hash, err := blogPasswordHash(input.Password, "")
	if err != nil {
		blogInternalError(w)
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		blogInternalError(w)
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck
	current, err := scanBlogPost(tx.QueryRow(r.Context(), "select "+blogPostColumns+" from blog_posts where id = $1 and deleted_at is null for update", id))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "post_not_found", "post not found")
		return
	}
	if err != nil {
		blogInternalError(w)
		return
	}
	if current.Version != *input.Version {
		writeError(w, http.StatusConflict, "version_conflict", "post changed; reload before saving")
		return
	}
	if err := validateBlogMediaReferences(r, tx, id, input.Blocks); err != nil {
		if errors.Is(err, errInvalidBlogMedia) {
			writeError(w, http.StatusBadRequest, "invalid_media", err.Error())
		} else {
			blogInternalError(w)
		}
		return
	}
	if input.Password == nil {
		hash = current.passwordHash
	}
	accessVersion := current.accessVersion
	if input.Password != nil || input.Status != current.Status {
		accessVersion++
	}
	raw, _ := json.Marshal(input.Blocks)
	post, err := scanBlogPost(tx.QueryRow(r.Context(), `update blog_posts set title=$2,blocks=$3,status=$4,password_hash=$5,access_version=$6,version=version+1,updated_at=now(),published_at=case when $4='published' then coalesce(published_at,now()) else published_at end where id=$1 returning `+blogPostColumns, id, input.Title, raw, input.Status, hash, accessVersion))
	if err != nil {
		blogInternalError(w)
		return
	}
	if accessVersion != current.accessVersion {
		if _, err := tx.Exec(r.Context(), "delete from blog_reader_grants where post_id=$1", id); err != nil {
			blogInternalError(w)
			return
		}
	}
	if err := tx.Commit(r.Context()); err != nil {
		blogInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"post": post})
}

func (a *App) deleteBlogPost(w http.ResponseWriter, r *http.Request) {
	id, ok := blogPathID(w, r)
	if !ok {
		return
	}
	tx, err := a.db.Begin(r.Context())
	if err != nil {
		blogInternalError(w)
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck
	result, err := tx.Exec(r.Context(), "update blog_posts set deleted_at=now(),updated_at=now(),access_version=access_version+1,version=version+1 where id=$1 and deleted_at is null", id)
	if err != nil {
		blogInternalError(w)
		return
	}
	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "post_not_found", "post not found")
		return
	}
	if _, err := tx.Exec(r.Context(), "delete from blog_reader_grants where post_id=$1", id); err != nil {
		blogInternalError(w)
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		blogInternalError(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func blogInternalError(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, "internal_error", "unable to process article request")
}
