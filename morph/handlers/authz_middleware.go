package handlers

import (
	"net/http"
	"strings"

	"idongivaflyinfa/auth"

	"github.com/gin-gonic/gin"
)

const (
	roleAdmin    = "admin"
	roleEmployee = "employee"
)

type userScope struct {
	userID  string
	role    string
	email   string
	isAdmin bool
}

func (h *Handlers) AuthzMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}
		// Public auth endpoints (login). /api/auth/me|user|permissions require Bearer below.
		if path == "/api/auth/login" {
			c.Next()
			return
		}
		if strings.HasPrefix(path, "/api/tran/public/") {
			c.Next()
			return
		}
		if strings.HasPrefix(path, "/api/auth/") {
			// Still require auth for me/user/permissions — handlers validate Bearer.
			// Let them run without AuthzMiddleware abort so they can return their own errors.
			if path == "/api/auth/me" || path == "/api/auth/user" || path == "/api/auth/permissions" {
				c.Next()
				return
			}
			c.Next()
			return
		}

		scope, ok := h.resolveUserScope(c)

		// Morph Data surfaces are usable without login. When a Morph AI session is present,
		// attach identity; otherwise continue anonymously.
		if isOpenMorphDataAPI(path) {
			if ok {
				h.applyAuthScope(c, scope)
			}
			c.Next()
			return
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		h.applyAuthScope(c, scope)

		// Admin-only surfaces.
		if strings.HasPrefix(path, "/api/admin/") || strings.HasPrefix(path, "/api/data-collector") {
			if !scope.isAdmin {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
				return
			}
		}

		// Every authenticated user has full product access.
		c.Next()
	}
}

func isOpenMorphDataAPI(path string) bool {
	prefixes := []string{
		"/api/tran/",
		"/api/forms/",
		"/api/knowledge/",
		"/api/graph/",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func (h *Handlers) applyAuthScope(c *gin.Context, scope userScope) {
	c.Set("auth_user_id", scope.userID)
	c.Set("auth_role", scope.role)
	c.Set("auth_email", scope.email)
	c.Set("auth_is_admin", scope.isAdmin)
	c.Request.Header.Set("X-User-ID", scope.userID)
	if scope.email != "" {
		c.Request.Header.Set("X-User-Email", scope.email)
	}
	if scope.isAdmin {
		c.Request.Header.Set("X-User-Role", "admin")
	} else {
		c.Request.Header.Set("X-User-Role", "employee")
	}
}

func (h *Handlers) resolveUserScope(c *gin.Context) (userScope, bool) {
	token := bearerToken(c.GetHeader("Authorization"))
	if token != "" && h.TranMySQL != nil {
		claims, err := auth.DecodeToken(h.jwtCfg, token)
		if err != nil {
			return userScope{}, false
		}
		u, err := h.TranMySQL.GetPlatUserByID(c.Request.Context(), claims.Subject)
		if err != nil {
			return userScope{}, false
		}
		isAdmin := u.IsAdmin()
		role := roleEmployee
		if isAdmin {
			role = roleAdmin
		}
		return userScope{
			userID:  u.ID,
			role:   role,
			email:  u.Email,
			isAdmin: isAdmin,
		}, true
	}

	// Legacy header fallback for internal tool calls.
	userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if userID == "" {
		return userScope{}, false
	}
	roleHdr := strings.ToLower(strings.TrimSpace(c.GetHeader("X-User-Role")))
	isAdmin := roleHdr == roleAdmin || strings.Contains(strings.ToLower(c.GetHeader("X-User-Roles")), "admin")
	role := roleEmployee
	if isAdmin {
		role = roleAdmin
	}
	return userScope{
		userID:  userID,
		role:   role,
		email:  strings.TrimSpace(c.GetHeader("X-User-Email")),
		isAdmin: isAdmin,
	}, true
}

func (h *Handlers) requireAdmin(c *gin.Context) bool {
	v, ok := c.Get("auth_is_admin")
	if ok {
		if b, ok := v.(bool); ok && b {
			return true
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
	return false
}
