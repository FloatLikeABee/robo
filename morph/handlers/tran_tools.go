package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

// tranUserIDFromContext resolves Tran `User.UserID` for tables keyed by that column.
// Order: ?user_id=, Tran User row by auth_email (UsersPanel profile), numeric X-User-ID,
// else 1 for backwards-compatible demo installs.
func (h *Handlers) tranUserIDFromContext(c *gin.Context) int {
	if id := c.Query("user_id"); id != "" {
		if n, err := strconv.Atoi(id); err == nil && n > 0 {
			return n
		}
	}
	if h.TranMySQL != nil {
		if v, ok := c.Get("auth_email"); ok {
			if s, ok := v.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					var uid int
					err := h.TranMySQL.DB.QueryRow(
						"SELECT UserID FROM `User` WHERE Email IS NOT NULL AND LOWER(TRIM(Email)) = LOWER(?) AND Deactivated = 0 LIMIT 1",
						s).Scan(&uid)
					if err == nil && uid > 0 {
						return uid
					}
				}
			}
		}
	}
	if id := c.GetHeader("X-User-ID"); id != "" {
		if n, err := strconv.Atoi(id); err == nil && n > 0 {
			return n
		}
	}
	return 1
}

// ListToolNotes returns notes for the current user: as author (mine) or as target (inbox). Query: box=mine|inbox|public, user_id (current user).
func (h *Handlers) ListToolNotes(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	userID := h.tranUserIDFromContext(c)
	box := c.DefaultQuery("box", "mine")
	var rows *sql.Rows
	var err error
	switch box {
	case "mine":
		rows, err = h.TranMySQL.DB.Query(
			`SELECT ID, AuthorUserID, TargetType, TargetUserID, IsPrivate, Title, Body, ReadAt, CreatedOn, LastUpdated
			 FROM tool_note WHERE AuthorUserID = ? ORDER BY CreatedOn DESC LIMIT 200`, userID)
	case "inbox":
		rows, err = h.TranMySQL.DB.Query(
			`SELECT ID, AuthorUserID, TargetType, TargetUserID, IsPrivate, Title, Body, ReadAt, CreatedOn, LastUpdated
			 FROM tool_note WHERE TargetType = 'user' AND TargetUserID = ? ORDER BY CreatedOn DESC LIMIT 200`, userID)
	case "public":
		rows, err = h.TranMySQL.DB.Query(
			`SELECT ID, AuthorUserID, TargetType, TargetUserID, IsPrivate, Title, Body, ReadAt, CreatedOn, LastUpdated
			 FROM tool_note WHERE TargetType = 'public' AND IsPrivate = 0 ORDER BY CreatedOn DESC LIMIT 200`)
	default:
		rows, err = h.TranMySQL.DB.Query(
			`SELECT ID, AuthorUserID, TargetType, TargetUserID, IsPrivate, Title, Body, ReadAt, CreatedOn, LastUpdated
			 FROM tool_note WHERE AuthorUserID = ? ORDER BY CreatedOn DESC LIMIT 200`, userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var list []models.ToolNote
	for rows.Next() {
		var n models.ToolNote
		if err := rows.Scan(&n.ID, &n.AuthorUserID, &n.TargetType, &n.TargetUserID, &n.IsPrivate, &n.Title, &n.Body, &n.ReadAt, &n.CreatedOn, &n.LastUpdated); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		list = append(list, n)
	}
	c.JSON(http.StatusOK, list)
}

// GetToolNote returns one note by ID.
func (h *Handlers) GetToolNote(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var n models.ToolNote
	row := h.TranMySQL.DB.QueryRow(
		`SELECT ID, AuthorUserID, TargetType, TargetUserID, IsPrivate, Title, Body, ReadAt, CreatedOn, LastUpdated
		 FROM tool_note WHERE ID = ?`, id)
	if err := row.Scan(&n.ID, &n.AuthorUserID, &n.TargetType, &n.TargetUserID, &n.IsPrivate, &n.Title, &n.Body, &n.ReadAt, &n.CreatedOn, &n.LastUpdated); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}
	c.JSON(http.StatusOK, n)
}

// CreateToolNote creates a note/message. Body: author_user_id (or use current), target_type (self|user|public), target_user_id?, is_private, title, body.
func (h *Handlers) CreateToolNote(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in struct {
		AuthorUserID *int   `json:"author_user_id"`
		TargetType   string `json:"target_type"`
		TargetUserID *int   `json:"target_user_id"`
		IsPrivate    *bool  `json:"is_private"`
		Title        string `json:"title"`
		Body         string `json:"body"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	authorID := h.tranUserIDFromContext(c)
	if in.AuthorUserID != nil && *in.AuthorUserID > 0 {
		authorID = *in.AuthorUserID
	}
	targetType := in.TargetType
	if targetType == "" {
		targetType = "self"
	}
	if targetType != "self" && targetType != "user" && targetType != "public" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_type must be self, user, or public"})
		return
	}
	isPrivate := true
	if in.IsPrivate != nil {
		isPrivate = *in.IsPrivate
	}
	var targetUserID interface{}
	if targetType == "user" && in.TargetUserID != nil {
		targetUserID = *in.TargetUserID
	} else {
		targetUserID = nil
	}
	res, err := h.TranMySQL.DB.Exec(
		`INSERT INTO tool_note (AuthorUserID, TargetType, TargetUserID, IsPrivate, Title, Body) VALUES (?, ?, ?, ?, ?, ?)`,
		authorID, targetType, targetUserID, isPrivate, in.Title, in.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id64})
}

// UpdateToolNote updates title/body.
func (h *Handlers) UpdateToolNote(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var in struct {
		Title *string `json:"title"`
		Body  *string `json:"body"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if in.Title == nil && in.Body == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title or body required"})
		return
	}
	if in.Title != nil && in.Body != nil {
		_, err = h.TranMySQL.DB.Exec(`UPDATE tool_note SET Title = ?, Body = ? WHERE ID = ?`, *in.Title, *in.Body, id)
	} else if in.Title != nil {
		_, err = h.TranMySQL.DB.Exec(`UPDATE tool_note SET Title = ? WHERE ID = ?`, *in.Title, id)
	} else {
		_, err = h.TranMySQL.DB.Exec(`UPDATE tool_note SET Body = ? WHERE ID = ?`, *in.Body, id)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": id})
}

// DeleteToolNote deletes a note.
func (h *Handlers) DeleteToolNote(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.TranMySQL.DB.Exec(`DELETE FROM tool_note WHERE ID = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}

// MarkToolNoteRead sets ReadAt for a note (inbox).
func (h *Handlers) MarkToolNoteRead(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	now := time.Now()
	_, err = h.TranMySQL.DB.Exec(`UPDATE tool_note SET ReadAt = ? WHERE ID = ? AND ReadAt IS NULL`, now, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"read_at": now})
}
