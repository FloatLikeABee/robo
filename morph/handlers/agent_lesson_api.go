package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

// requestUserID is the authenticated operator. AuthzMiddleware sets auth_user_id
// and copies it onto X-User-ID; tests and internal callers may set the header.
func requestUserID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get("auth_user_id"); ok {
		if s, ok := v.(string); ok {
			if id := strings.TrimSpace(s); id != "" {
				return id
			}
		}
	}
	return strings.TrimSpace(c.GetHeader("X-User-ID"))
}

func agentLessonJSON(l db.AgentLesson) gin.H {
	return gin.H{
		"id":                l.ID,
		"trigger":           l.Trigger,
		"rule":              l.Rule,
		"source_session_id": l.SourceSessionID,
		"created_at":        l.CreatedAt,
		"enabled":           l.Enabled,
		"owner_user_id":     l.OwnerUserID,
	}
}

// ListAgentLessons GET /api/agent-lessons
// Returns every lesson owned by the caller, including disabled ones.
func (h *Handlers) ListAgentLessons(c *gin.Context) {
	if h == nil || h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sqlite not available"})
		return
	}
	rows, err := h.TranMySQL.ListAgentLessons(c.Request.Context(), requestUserID(c), false, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []db.AgentLesson{}
	}
	out := make([]gin.H, 0, len(rows))
	for _, l := range rows {
		out = append(out, agentLessonJSON(l))
	}
	c.JSON(http.StatusOK, gin.H{"lessons": out, "total": len(out)})
}

type agentLessonPatchBody struct {
	Enabled *bool `json:"enabled"`
}

// PatchAgentLesson PATCH /api/agent-lessons/:id
// Body: {"enabled": true|false}. Another user's lesson is 404.
func (h *Handlers) PatchAgentLesson(c *gin.Context) {
	if h == nil || h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sqlite not available"})
		return
	}
	var body agentLessonPatchBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if body.Enabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "enabled is required"})
		return
	}
	updated, err := h.TranMySQL.SetAgentLessonEnabled(c.Request.Context(), requestUserID(c), c.Param("id"), *body.Enabled)
	if err == sql.ErrNoRows || (err == nil && updated == nil) {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, agentLessonJSON(*updated))
}

// DeleteAgentLesson DELETE /api/agent-lessons/:id
// Another user's lesson is 404.
func (h *Handlers) DeleteAgentLesson(c *gin.Context) {
	if h == nil || h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sqlite not available"})
		return
	}
	err := h.TranMySQL.DeleteAgentLessonForOwner(c.Request.Context(), requestUserID(c), c.Param("id"))
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "lesson not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
