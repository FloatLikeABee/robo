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
		role := resolveRole(c.GetHeader("X-User-Role"), c.GetHeader("X-User-Roles"))
		rawPermissions := c.GetHeader("X-User-Permissions")
		token := bearerToken(c.GetHeader("Authorization"))
		var panelPerms []string
		if token != "" {
			resolvedRole, perms := resolveRoleAndPermissionsFromUsersPanel(c, token)
			panelPerms = perms
			if resolvedRole != "" {
				role = resolvedRole
			}
			if len(perms) > 0 {
				rawPermissions = strings.Join(perms, ",")
				c.Request.Header.Set("X-User-Permissions", rawPermissions)
			}
		}
		// Morph-hosted auth: any authenticated user may use SheetX (no app permission gates).
		if role == "admin" || len(panelPerms) > 0 {
			c.Next()
			return
		}
		if role == "employee" && (hasPermission(rawPermissions, tranformPermissionA) || hasPermission(rawPermissions, tranformPermissionB)) {
			c.Next()
			return
		}
		if token == "" && role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "tranform access is restricted by admin policy"})
	}
}

func resolveRoleAndPermissionsFromUsersPanel(c *gin.Context, token string) (string, []string) {
	baseURL := ""
	if hRaw, ok := c.Get("handler_instance"); ok {
		if h, ok := hRaw.(*Handler); ok && h != nil {
			baseURL = strings.TrimRight(h.Cfg.UsersPanelBaseURL, "/")
		}
	}
	if baseURL == "" {
		baseURL = strings.TrimRight(c.GetHeader("X-UsersPanel-BaseURL"), "/")
	}
	if baseURL == "" {
		return "", nil
	}

	reqUser, _ := http.NewRequest(http.MethodGet, baseURL+"/api/auth/user", nil)
	reqUser.Header.Set("Authorization", "Bearer "+token)
	userResp, err := http.DefaultClient.Do(reqUser)
	if err != nil {
		return "", nil
	}
	defer userResp.Body.Close()
	if userResp.StatusCode != http.StatusOK {
		return "", nil
	}
	userBody, _ := io.ReadAll(userResp.Body)
	var userPayload struct {
		User struct {
			Roles []string `json:"roles"`
		} `json:"user"`
	}
	if json.Unmarshal(userBody, &userPayload) != nil {
		return "", nil
	}
	role := resolveRole("", strings.Join(userPayload.User.Roles, ","))

	reqPerms, _ := http.NewRequest(http.MethodGet, baseURL+"/api/auth/permissions", nil)
	reqPerms.Header.Set("Authorization", "Bearer "+token)
	permsResp, err := http.DefaultClient.Do(reqPerms)
	if err != nil {
		return role, nil
	}
	defer permsResp.Body.Close()
	if permsResp.StatusCode != http.StatusOK {
		return role, nil
	}
	permsBody, _ := io.ReadAll(permsResp.Body)
	var permsPayload struct {
		Permissions []string `json:"permissions"`
	}
	if json.Unmarshal(permsBody, &permsPayload) != nil {
		return role, nil
	}
	return role, permsPayload.Permissions
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

