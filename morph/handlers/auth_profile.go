package handlers

import (
	"net/http"
	"strings"

	"idongivaflyinfa/auth"

	"github.com/gin-gonic/gin"
)

type patchAuthMeBody struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	CurrentPassword string `json:"current_password"`
}

// PatchMorphAuthMe PATCH /api/auth/me — self-service username/password update.
func (h *Handlers) PatchMorphAuthMe(c *gin.Context) {
	u, ok := h.userFromBearer(c)
	if !ok {
		return
	}
	var body patchAuthMeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	updated, err := h.TranMySQL.UpdatePlatUserCredentials(
		c.Request.Context(),
		u.ID,
		strings.TrimSpace(body.Username),
		strings.TrimSpace(body.Password),
		body.CurrentPassword,
	)
	if err != nil {
		msg := err.Error()
		st := http.StatusBadRequest
		switch {
		case strings.Contains(msg, "current password"):
			st = http.StatusUnauthorized
		case strings.Contains(msg, "already taken"):
			st = http.StatusConflict
		}
		c.JSON(st, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user":        updated.Public(),
		"permissions": auth.FullPermissions(),
	})
}
