package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createAgentTokenRequest struct {
	Name          string   `json:"name"`
	Scopes        []string `json:"scopes"`
	ExpiresInDays int      `json:"expires_in_days"`
}

func (h *Handlers) requireTokenUser(c *gin.Context) (string, bool) {
	u, ok := h.userFromBearer(c)
	if !ok {
		return "", false
	}
	return u.ID, true
}

// CreateAgentToken POST /api/agent-tokens
func (h *Handlers) CreateAgentToken(c *gin.Context) {
	userID, ok := h.requireTokenUser(c)
	if !ok {
		return
	}
	var in createAgentTokenRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	tok, err := h.TranMySQL.CreateAgentToken(c.Request.Context(), userID, in.Name, in.Scopes, in.ExpiresInDays)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body := tok.Public()
	body["token"] = tok.Token
	c.JSON(http.StatusOK, body)
}

// ListAgentTokens GET /api/agent-tokens
func (h *Handlers) ListAgentTokens(c *gin.Context) {
	userID, ok := h.requireTokenUser(c)
	if !ok {
		return
	}
	list, err := h.TranMySQL.ListAgentTokens(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list tokens"})
		return
	}
	out := make([]map[string]any, 0, len(list))
	for _, tok := range list {
		out = append(out, tok.Public())
	}
	c.JSON(http.StatusOK, gin.H{"tokens": out})
}

// RevokeAgentToken DELETE /api/agent-tokens/:id
func (h *Handlers) RevokeAgentToken(c *gin.Context) {
	userID, ok := h.requireTokenUser(c)
	if !ok {
		return
	}
	err := h.TranMySQL.RevokeAgentToken(c.Request.Context(), userID, c.Param("id"))
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not revoke token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"revoked": true})
}
