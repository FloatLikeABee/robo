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

		// Former open data APIs. Unsafe methods need a session. GET/HEAD stay
		// reachable without one so MorphNotes can still list records. Anything
		// under /api/tran/public/ that is not on the allowlist is not a private read.
		if isMorphDataAPI(path) {
			if strings.HasPrefix(path, "/api/tran/public/") || !isSafeMethod(c.Request.Method) {
				if !ok {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
					return
				}
			}
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

// morphDataAPIPrefixes used to skip auth entirely (isOpenMorphDataAPI).
// Writes on these prefixes now require a session. Reads do not.
var morphDataAPIPrefixes = []string{
	"/api/tran/",
	"/api/forms/",
	"/api/knowledge/",
	"/api/graph/",
}

// publicMorphReadKinds are the only unauthenticated published pages.
// A new public route must be added here and registered as GET.
var publicMorphReadKinds = map[string]struct{}{
	"big-notes": {},
	"timelines": {},
	"research":  {},
}

func isMorphDataAPI(path string) bool {
	for _, p := range morphDataAPIPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}

// isPublicMorphRead is the allowlist for published HTML.
// Only GET and HEAD of /api/tran/public/{kind}/{slug} match.
// Other methods and unknown kinds under /api/tran/public/ are not public.
func isPublicMorphRead(method, path string) bool {
	if method != http.MethodGet && method != http.MethodHead {
		return false
	}
	rest, ok := strings.CutPrefix(path, "/api/tran/public/")
	if !ok || rest == "" || strings.Contains(rest, "..") {
		return false
	}
	kind, slug, ok := strings.Cut(rest, "/")
	if !ok {
		return false
	}
	if _, known := publicMorphReadKinds[kind]; !known {
		return false
	}
	slug = strings.Trim(slug, "/")
	if slug == "" || strings.Contains(slug, "/") {
		return false
	}
	return true
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
			role:    role,
			email:   u.Email,
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
		role:    role,
		email:   strings.TrimSpace(c.GetHeader("X-User-Email")),
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
