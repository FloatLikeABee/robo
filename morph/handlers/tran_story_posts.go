package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	storyPostListCols = `sp.ID, sp.title, sp.content, sp.author_user_id, sp.created_on, sp.last_updated`
	storyPostEntity   = "story_post"
)

var allowedStoryPostWrite = map[string]struct{}{
	"title": {}, "content": {},
}

type storyPostRow struct {
	ID            int        `json:"id"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	AuthorUserID  int        `json:"author_user_id"`
	AuthorName    string     `json:"author_name,omitempty"`
	CreatedOn     *time.Time `json:"created_on,omitempty"`
	LastUpdated   *time.Time `json:"last_updated,omitempty"`
	CommentCount  int        `json:"comment_count"`
	Attachments   []entityAttachmentOut `json:"attachments,omitempty"`
}

func parseStoryPostRow(scanner interface {
	Scan(dest ...interface{}) error
}, includeCommentCount bool) (storyPostRow, error) {
	var r storyPostRow
	var createdOn, updatedOn sql.NullTime
	dest := []interface{}{&r.ID, &r.Title, &r.Content, &r.AuthorUserID, &createdOn, &updatedOn}
	if includeCommentCount {
		dest = append(dest, &r.CommentCount)
	}
	if err := scanner.Scan(dest...); err != nil {
		return r, err
	}
	if createdOn.Valid {
		ts := createdOn.Time
		r.CreatedOn = &ts
	}
	if updatedOn.Valid {
		ts := updatedOn.Time
		r.LastUpdated = &ts
	}
	return r, nil
}

func (h *Handlers) tranUserDisplayName(ctx context.Context, userID int) string {
	if h == nil || h.TranMySQL == nil || userID <= 0 {
		return ""
	}
	var fn, ln, em sql.NullString
	err := h.TranMySQL.DB.QueryRowContext(ctx,
		"SELECT FirstName, LastName, Email FROM `User` WHERE UserID = ? LIMIT 1", userID).
		Scan(&fn, &ln, &em)
	if err != nil {
		return ""
	}
	name := strings.TrimSpace(strings.TrimSpace(fn.String) + " " + strings.TrimSpace(ln.String))
	if name != "" {
		return name
	}
	if em.Valid && strings.TrimSpace(em.String) != "" {
		return strings.TrimSpace(em.String)
	}
	return ""
}

func (h *Handlers) enrichStoryPostRow(ctx context.Context, r *storyPostRow, withAttachments bool) {
	if r == nil {
		return
	}
	r.AuthorName = h.tranUserDisplayName(ctx, r.AuthorUserID)
	if withAttachments {
		list, err := h.listEntityAttachments(ctx, "story-posts", entityAttachmentStoryPost, r.ID)
		if err != nil || list == nil {
			r.Attachments = []entityAttachmentOut{}
		} else {
			r.Attachments = list
		}
	}
}

func (h *Handlers) ListStoryPosts(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	q := `
		SELECT sp.ID, sp.title, sp.content, sp.author_user_id, sp.created_on, sp.last_updated,
		       (SELECT COUNT(*) FROM ` + "`comment`" + ` c WHERE c.EntityType = ? AND c.RecordID = sp.ID) AS comment_count
		FROM StoryPost sp
		ORDER BY sp.created_on DESC, sp.ID DESC
		LIMIT 500`
	rows, err := h.TranMySQL.DB.Query(q, storyPostEntity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	ctx := c.Request.Context()
	list := make([]storyPostRow, 0, 64)
	for rows.Next() {
		r, err := parseStoryPostRow(rows, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		h.enrichStoryPostRow(ctx, &r, true)
		list = append(list, r)
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handlers) GetStoryPost(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	row := h.TranMySQL.DB.QueryRow(`
		SELECT sp.ID, sp.title, sp.content, sp.author_user_id, sp.created_on, sp.last_updated,
		       (SELECT COUNT(*) FROM ` + "`comment`" + ` c WHERE c.EntityType = ? AND c.RecordID = sp.ID) AS comment_count
		FROM StoryPost sp WHERE sp.ID = ?`, storyPostEntity, id)
	item, err := parseStoryPostRow(row, true)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story post not found"})
		return
	}
	h.enrichStoryPostRow(c.Request.Context(), &item, true)
	c.JSON(http.StatusOK, item)
}

func (h *Handlers) GetStoryPostFull(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, `
		SELECT
			ID AS id,
			title AS title,
			content AS content,
			author_user_id AS author_user_id,
			created_on AS created_on,
			last_updated AS last_updated
		FROM StoryPost
		WHERE ID = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "story post not found"})
		return
	}
	authorID := 0
	switch v := m["author_user_id"].(type) {
	case int64:
		authorID = int(v)
	case int:
		authorID = v
	case float64:
		authorID = int(v)
	}
	if authorID > 0 {
		m["author_name"] = h.tranUserDisplayName(c.Request.Context(), authorID)
	}
	var commentCount int
	_ = h.TranMySQL.DB.QueryRow(`
		SELECT COUNT(*) FROM comment WHERE EntityType = ? AND RecordID = ?`, storyPostEntity, id).Scan(&commentCount)
	m["comment_count"] = commentCount
	h.attachEntityAttachmentsToRow(c.Request.Context(), "story-posts", entityAttachmentStoryPost, id, m)
	c.JSON(http.StatusOK, m)
}

func (h *Handlers) CreateStoryPost(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in map[string]interface{}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	title, _ := in["title"].(string)
	content, _ := in["content"].(string)
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}
	authorID := h.tranUserIDFromContext(c)
	res, err := h.TranMySQL.DB.Exec(`
		INSERT INTO StoryPost (title, content, author_user_id) VALUES (?, ?, ?)`,
		title, content, authorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	id := int(id64)
	m, ok, err := querySingleRowMap(h.TranMySQL.DB, `
		SELECT
			ID AS id,
			title AS title,
			content AS content,
			author_user_id AS author_user_id,
			created_on AS created_on,
			last_updated AS last_updated
		FROM StoryPost
		WHERE ID = ?`, id)
	if err != nil || !ok {
		c.JSON(http.StatusOK, gin.H{"id": id, "title": title, "content": content, "author_user_id": authorID})
		return
	}
	m["author_name"] = h.tranUserDisplayName(c.Request.Context(), authorID)
	m["comment_count"] = 0
	h.attachEntityAttachmentsToRow(c.Request.Context(), "story-posts", entityAttachmentStoryPost, id, m)
	c.JSON(http.StatusOK, m)
}

func (h *Handlers) UpdateStoryPost(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var in map[string]interface{}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	var curTitle, curContent string
	if err := h.TranMySQL.DB.QueryRow("SELECT title, content FROM StoryPost WHERE ID = ?", id).Scan(&curTitle, &curContent); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "story post not found"})
		return
	}
	sets := make([]string, 0, 2)
	args := make([]interface{}, 0, 3)
	for k, v := range in {
		col := strings.ToLower(strings.TrimSpace(k))
		if _, ok := allowedStoryPostWrite[col]; !ok {
			continue
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		switch col {
		case "title":
			if s == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "title cannot be empty"})
				return
			}
			sets = append(sets, "title = ?")
			args = append(args, s)
		case "content":
			if s == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "content cannot be empty"})
				return
			}
			sets = append(sets, "content = ?")
			args = append(args, s)
		}
	}
	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid fields to update"})
		return
	}
	args = append(args, id)
	if _, err := h.TranMySQL.DB.Exec("UPDATE StoryPost SET "+strings.Join(sets, ", ")+" WHERE ID = ?", args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.GetStoryPostFull(c)
}

func (h *Handlers) DeleteStoryPost(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.TranMySQL.DB.Exec("DELETE FROM StoryPost WHERE ID = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "story post not found"})
		return
	}
	ctx := context.Background()
	h.purgeEntityAttachments(ctx, entityAttachmentStoryPost, id)
	_, _ = h.TranMySQL.DB.ExecContext(ctx, "DELETE FROM comment WHERE EntityType = ? AND RecordID = ?", storyPostEntity, id)
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
