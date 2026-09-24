package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"idongivaflyinfa/auth"
	"idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

// trustedLessonUserID is the platform user id from a verified bearer token
// whose subject still exists in plat_users. Client X-User-ID and the
// auth_user_id value middleware copies from that header are not accepted.
// ok is false when no trusted identity exists. This does not write a response.
func (h *Handlers) trustedLessonUserID(c *gin.Context) (string, bool) {
	if h == nil || c == nil || h.TranMySQL == nil {
		return "", false
	}
	token := bearerToken(c.GetHeader("Authorization"))
	if token == "" {
		return "", false
	}
	claims, err := auth.DecodeToken(h.jwtCfg, token)
	if err != nil || claims == nil {
		return "", false
	}
	id := strings.TrimSpace(claims.Subject)
	if id == "" {
		return "", false
	}
	u, err := h.TranMySQL.GetPlatUserByID(c.Request.Context(), id)
	if err != nil || u == nil || strings.TrimSpace(u.ID) == "" {
		return "", false
	}
	return u.ID, true
}

func (h *Handlers) requireLessonUser(c *gin.Context) (string, bool) {
	if c == nil {
		return "", false
	}
	// A header-only caller is unauthenticated even when the store is down.
	// 503 is only for a presented bearer that cannot be checked.
	if bearerToken(c.GetHeader("Authorization")) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return "", false
	}
	if h == nil || h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sqlite not available"})
		return "", false
	}
	id, ok := h.trustedLessonUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return "", false
	}
	return id, true
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
	userID, ok := h.requireLessonUser(c)
	if !ok {
		return
	}
	rows, err := h.TranMySQL.ListAgentLessons(c.Request.Context(), userID, false, 0)
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
	userID, ok := h.requireLessonUser(c)
	if !ok {
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
	updated, err := h.TranMySQL.SetAgentLessonEnabled(c.Request.Context(), userID, c.Param("id"), *body.Enabled)
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
	userID, ok := h.requireLessonUser(c)
	if !ok {
		return
	}
	err := h.TranMySQL.DeleteAgentLessonForOwner(c.Request.Context(), userID, c.Param("id"))
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
