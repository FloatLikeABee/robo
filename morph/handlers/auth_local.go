package handlers

import (
	"net/http"
	"strings"

	"idongivaflyinfa/auth"
	"idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handlers) requireAuthDB(c *gin.Context) bool {
	if h.TranMySQL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth database unavailable"})
		return false
	}
	return true
}

// MorphAuthLogin POST /api/auth/login — local username or email + password, JWT session.
func (h *Handlers) MorphAuthLogin(c *gin.Context) {
	if !h.requireAuthDB(c) {
		return
	}
	var in loginRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	login := strings.TrimSpace(in.Email)
	if login == "" {
		login = strings.TrimSpace(in.Username)
	}
	if login == "" || in.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username/email and password required"})
		return
	}
	u, err := h.TranMySQL.GetPlatUserByLogin(c.Request.Context(), login)
	if err != nil || !db.VerifyPassword(u.PasswordHash, in.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username/email or password"})
		return
	}
	token, err := auth.EncodeToken(h.jwtCfg, u.ID, u.Email, u.Username, u.Roles, u.DefaultChannelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":       token,
		"user":        u.Public(),
		"permissions": auth.FullPermissions(),
	})
}

// MorphAuthMe GET /api/auth/me
func (h *Handlers) MorphAuthMe(c *gin.Context) {
	u, ok := h.userFromBearer(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user":        u.Public(),
		"permissions": auth.FullPermissions(),
	})
}

// MorphAuthUser GET /api/auth/user — UsersPanel-compatible session endpoint.
func (h *Handlers) MorphAuthUser(c *gin.Context) {
	u, ok := h.userFromBearer(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": u.Public()})
}

// MorphAuthPermissions GET /api/auth/permissions — always full access for any authed user.
func (h *Handlers) MorphAuthPermissions(c *gin.Context) {
	if _, ok := h.userFromBearer(c); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"permissions": auth.FullPermissions()})
}

func (h *Handlers) userFromBearer(c *gin.Context) (*db.PlatUser, bool) {
	if !h.requireAuthDB(c) {
		return nil, false
	}
	token := bearerToken(c.GetHeader("Authorization"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
		return nil, false
	}
	claims, err := auth.DecodeToken(h.jwtCfg, token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return nil, false
	}
	u, err := h.TranMySQL.GetPlatUserByID(c.Request.Context(), claims.Subject)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return nil, false
	}
	return u, true
}

func bearerToken(h string) string {
	h = strings.TrimSpace(h)
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
