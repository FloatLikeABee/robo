package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// Must match plat_permissions names granted via UsersPanel role mappings (e.g. "Forms" → create_form).
	tranformPermissionA = "create_form"
	tranformPermissionB = "broadcast_form"
)

func requireWorkspaceAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		// Client identity headers are not a session. Morph strips the same set
		// before it trusts a JWT. Role and permissions come only from Morph's
		// response to the bearer.
		stripClientIdentityHeaders(c.Request)
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		role, panelPerms, ok := resolveRoleAndPermissionsFromUsersPanel(c, token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		// Morph-hosted auth: an authenticated admin, or any user Morph granted
		// permissions, may use Event Logs. Employee still needs a form permission
		// when Morph returned a role and an empty permission list.
		rawPermissions := strings.Join(panelPerms, ",")
		if role == "admin" || len(panelPerms) > 0 {
			c.Next()
			return
		}
		if role == "employee" && (hasPermission(rawPermissions, tranformPermissionA) || hasPermission(rawPermissions, tranformPermissionB)) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "tranform access is restricted by admin policy"})
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
		"X-UsersPanel-BaseURL",
	} {
		r.Header.Del(key)
	}
}

// resolveRoleAndPermissionsFromUsersPanel asks the configured Morph origin
// whether the bearer is a user. The third result is false when that call does
// not succeed. A client-supplied base URL is not an origin.
func resolveRoleAndPermissionsFromUsersPanel(c *gin.Context, token string) (string, []string, bool) {
	baseURL := ""
	if hRaw, ok := c.Get("handler_instance"); ok {
		if h, ok := hRaw.(*Handler); ok && h != nil && h.Cfg != nil {
			baseURL = strings.TrimRight(h.Cfg.UsersPanelBaseURL, "/")
		}
	}
	if baseURL == "" {
		return "", nil, false
	}

	reqUser, err := http.NewRequest(http.MethodGet, baseURL+"/api/auth/user", nil)
	if err != nil {
		return "", nil, false
	}
	reqUser.Header.Set("Authorization", "Bearer "+token)
	userResp, err := http.DefaultClient.Do(reqUser)
	if err != nil {
		return "", nil, false
	}
	defer userResp.Body.Close()
	if userResp.StatusCode != http.StatusOK {
		return "", nil, false
	}
	userBody, _ := io.ReadAll(userResp.Body)
	var userPayload struct {
		User struct {
			Roles []string `json:"roles"`
		} `json:"user"`
	}
	if json.Unmarshal(userBody, &userPayload) != nil {
		return "", nil, false
	}
	role := resolveRole("", strings.Join(userPayload.User.Roles, ","))

	reqPerms, err := http.NewRequest(http.MethodGet, baseURL+"/api/auth/permissions", nil)
	if err != nil {
		return role, nil, true
	}
	reqPerms.Header.Set("Authorization", "Bearer "+token)
	permsResp, err := http.DefaultClient.Do(reqPerms)
	if err != nil {
		return role, nil, true
	}
	defer permsResp.Body.Close()
	if permsResp.StatusCode != http.StatusOK {
		return role, nil, true
	}
	permsBody, _ := io.ReadAll(permsResp.Body)
	var permsPayload struct {
		Permissions []string `json:"permissions"`
	}
	if json.Unmarshal(permsBody, &permsPayload) != nil {
		return role, nil, true
	}
	return role, permsPayload.Permissions, true
}

func resolveRole(roleHeader, rolesHeader string) string {
	role := strings.ToLower(strings.TrimSpace(roleHeader))
	if role != "" {
		return role
	}
	out := ""
	for _, raw := range strings.Split(rolesHeader, ",") {
		r := strings.ToLower(strings.TrimSpace(raw))
		switch r {
		case "admin":
			return "admin"
		case "employee":
			out = "employee"
		case "member":
			if out == "" {
				out = "member"
			}
		// UsersPanel plat_roles.name values (JWT / GET /api/auth/user)
		case "forms", "email composer", "main panel", "sharp reports":
			if out != "admin" {
				out = "employee"
			}
		}
	}
	return out
}

func hasPermission(rawCSV, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, raw := range strings.Split(rawCSV, ",") {
		if strings.ToLower(strings.TrimSpace(raw)) == target {
			return true
		}
	}
	return false
}
