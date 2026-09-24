package handlers

import (
	"context"
	"net/http"
	"strings"

	"idongivaflyinfa/auth"
	"idongivaflyinfa/publish"

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

// UserScope is the caller resolved from a Morph JWT.
// The token selects the plat_users row. Role is that row's current role.
// Request identity headers are not consulted.
type UserScope struct {
	UserID  string
	Role    string
	Email   string
	IsAdmin bool
}

// ResolveUserScope returns the session for r. A missing token, an invalid
// token, an unknown user, or a nil store is not a session.
func (h *Handlers) ResolveUserScope(r *http.Request) (UserScope, bool) {
	if h == nil || r == nil || h.TranMySQL == nil {
		return UserScope{}, false
	}
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		return UserScope{}, false
	}
	claims, err := auth.DecodeToken(h.jwtCfg, token)
	if err != nil {
		return UserScope{}, false
	}
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	u, err := h.TranMySQL.GetPlatUserByID(ctx, claims.Subject)
	if err != nil || u == nil {
		return UserScope{}, false
	}
	isAdmin := u.IsAdmin()
	role := roleEmployee
	if isAdmin {
		role = roleAdmin
	}
	return UserScope{
		UserID:  u.ID,
		Role:    role,
		Email:   u.Email,
		IsAdmin: isAdmin,
	}, true
}

func (h *Handlers) AuthzMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// CORS in main.go answers OPTIONS with 204 before this runs.
		// Pass through here too, so a reorder cannot turn preflight into 401.
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}
		stripClientIdentityHeaders(c.Request)
		// Public auth endpoints (login, invite redeem). /api/auth/me|user|permissions validate Bearer in handlers.
		if path == "/api/auth/login" || path == "/api/invite/redeem" {
			c.Next()
			return
		}
		// Published HTML is an explicit GET/HEAD allowlist, not the whole /api/tran/public/ prefix.
		if isPublicMorphRead(c.Request.Method, path) {
			c.Next()
			return
		}
		if strings.HasPrefix(path, "/api/auth/") {
			// Let auth handlers run without AuthzMiddleware abort so they can return their own errors.
			if path == "/api/auth/me" || path == "/api/auth/user" || path == "/api/auth/permissions" {
				c.Next()
				return
			}
			c.Next()
			return
		}

		scope, ok := h.resolveUserScope(c)

		// Private Morph data. Every method needs a session, including GET and HEAD.
		// Published pages return earlier through publish.PageRoute.
		if isMorphDataAPI(path) {
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
				return
			}
			h.applyAuthScope(c, scope)
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

// morphDataAPIPrefixes are private. Every method requires a session.
// Published HTML is publish.PageRoute, not these prefixes.
var morphDataAPIPrefixes = []string{
	"/api/tran/",
	"/api/forms/",
	"/api/knowledge/",
	"/api/graph/",
}

func isMorphDataAPI(path string) bool {
	for _, p := range morphDataAPIPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// isPublicMorphRead is the allowlist for published HTML.
// The rule lives in publish.PageRoute so MCP can call the same package.
func isPublicMorphRead(method, path string) bool {
	return publish.PageRoute(method, path)
}

// requireSessionUserID is the chat user id. An empty id is not admin.
func requireSessionUserID(c *gin.Context) (string, bool) {
	id := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if id == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return "", false
	}
	return id, true
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

func stripClientIdentityHeaders(r *http.Request) {
	if r == nil {
		return
	}
	for _, key := range []string{
		"X-User-ID",
		"X-User-Role",
		"X-User-Roles",
		"X-User-Email",
		"X-User-Permissions",
	} {
		r.Header.Del(key)
	}
}

func (h *Handlers) resolveUserScope(c *gin.Context) (userScope, bool) {
	scope, ok := h.ResolveUserScope(c.Request)
	if !ok {
		return userScope{}, false
	}
	return userScope{
		userID:  scope.UserID,
		role:    scope.Role,
		email:   scope.Email,
		isAdmin: scope.IsAdmin,
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
