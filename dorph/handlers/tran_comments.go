package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"idongivaflyinfa/models"

	"github.com/gin-gonic/gin"
)

var commentEntityTypes = map[string]struct{}{
	"student": {}, "staff": {}, "school": {}, "vehicle": {}, "trip": {}, "story_post": {},
}

// ListComments returns comments for an entity as a tree (top-level with nested replies).
func (h *Handlers) ListComments(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	entityType := strings.ToLower(strings.TrimSpace(c.Query("entity_type")))
	recordIDStr := c.Query("record_id")
	if entityType == "" || recordIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_type and record_id are required"})
		return
	}
	if _, ok := commentEntityTypes[entityType]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity_type"})
		return
	}
	recordID, err := strconv.Atoi(recordIDStr)
	if err != nil || recordID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid record_id"})
		return
	}

	rows, err := h.TranMySQL.DB.Query(
		`SELECT ID, EntityType, RecordID, ParentID, AuthorUserID, Body, CreatedOn, LastUpdated
		 FROM comment WHERE EntityType = ? AND RecordID = ? ORDER BY CreatedOn ASC`,
		entityType, recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var flat []models.Comment
	for rows.Next() {
		var cm models.Comment
		var parentID interface{}
		var authorID interface{}
		if err := rows.Scan(&cm.ID, &cm.EntityType, &cm.RecordID, &parentID, &authorID, &cm.Body, &cm.CreatedOn, &cm.LastUpdated); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if parentID != nil {
			var n int
			switch v := parentID.(type) {
			case int64:
				n = int(v)
			case int32:
				n = int(v)
			case int:
				n = v
			default:
				n = 0
			}
			if n > 0 {
				cm.ParentID = &n
			}
		}
		if authorID != nil {
			var n int
			switch v := authorID.(type) {
			case int64:
				n = int(v)
			case int32:
				n = int(v)
			case int:
				n = v
			default:
				n = 0
			}
			if n > 0 {
				cm.AuthorUserID = &n
			}
		}
		flat = append(flat, cm)
	}

	tree := buildCommentTree(flat, nil)
	c.JSON(http.StatusOK, tree)
}

func buildCommentTree(flat []models.Comment, parentID *int) []models.Comment {
	var out []models.Comment
	for _, c := range flat {
		if (parentID == nil && c.ParentID == nil) || (parentID != nil && c.ParentID != nil && *c.ParentID == *parentID) {
			replies := buildCommentTree(flat, &c.ID)
			if len(replies) > 0 {
				c.Replies = replies
			}
			out = append(out, c)
		}
	}
	return out
}

// CreateComment creates a comment or reply. Body: entity_type, record_id, parent_id (optional), body.
func (h *Handlers) CreateComment(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	var in struct {
		EntityType   string `json:"entity_type"`
		RecordID     int    `json:"record_id"`
		ParentID     *int   `json:"parent_id"`
		Body         string `json:"body"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	entityType := strings.ToLower(strings.TrimSpace(in.EntityType))
	if _, ok := commentEntityTypes[entityType]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entity_type"})
		return
	}
	if in.RecordID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "record_id is required"})
		return
	}
	if strings.TrimSpace(in.Body) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}
	authorID := h.tranUserIDFromContext(c)
	res, err := h.TranMySQL.DB.Exec(
		`INSERT INTO comment (EntityType, RecordID, ParentID, AuthorUserID, Body) VALUES (?, ?, ?, ?, ?)`,
		entityType, in.RecordID, in.ParentID, authorID, in.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	id64, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id64})
}

// UpdateComment updates a comment body.
func (h *Handlers) UpdateComment(c *gin.Context) {
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
		Body string `json:"body"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(in.Body) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}
	_, err = h.TranMySQL.DB.Exec(`UPDATE comment SET Body = ? WHERE ID = ?`, in.Body, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": id})
}

// DeleteComment deletes a comment (and typically cascades to replies if desired; here we just delete one).
func (h *Handlers) DeleteComment(c *gin.Context) {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Tran SQL store not configured"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	res, err := h.TranMySQL.DB.Exec(`DELETE FROM comment WHERE ID = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": id})
}
