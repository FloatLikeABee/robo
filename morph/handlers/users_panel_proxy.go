package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// UsersPanelProxy used to forward to UsersPanel. Auth now lives on Morph; when
// USERS_PANEL_BASE_URL is unset, rewrite /api/users-panel/api/auth/* to Morph auth
// and stub other paths (messaging etc.).
func (h *Handlers) UsersPanelProxy(c *gin.Context) {
	trimmed := strings.TrimPrefix(c.Param("filepath"), "/")
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	// /api/users-panel/api/auth/login → local Morph auth
	if strings.HasPrefix(trimmed, "/api/auth/") || trimmed == "/api/auth" {
		c.Request.URL.Path = trimmed
		switch {
		case strings.HasSuffix(trimmed, "/login") && c.Request.Method == http.MethodPost:
			h.MorphAuthLogin(c)
			return
		case strings.HasSuffix(trimmed, "/user") && c.Request.Method == http.MethodGet:
			h.MorphAuthUser(c)
			return
		case strings.HasSuffix(trimmed, "/permissions") && c.Request.Method == http.MethodGet:
			h.MorphAuthPermissions(c)
			return
		case strings.HasSuffix(trimmed, "/me") && c.Request.Method == http.MethodGet:
			h.MorphAuthMe(c)
			return
		}
	}
	if strings.HasPrefix(trimmed, "/api/messages") {
		c.Params = append(c.Params, gin.Param{Key: "filepath", Value: trimmed})
		h.MessagesStub(c)
		return
	}

	base := strings.TrimSpace(h.usersPanelBaseURL)
	if base == "" {
		c.JSON(http.StatusGone, gin.H{"error": "UsersPanel removed; use Morph /api/auth/* and /api/admin/users"})
		return
	}
	target, err := url.Parse(base)
	if err != nil || target.Scheme == "" || target.Host == "" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid users panel URL"})
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	orig := proxy.Director
	proxy.Director = func(req *http.Request) {
		orig(req)
		trimmed := strings.TrimPrefix(req.URL.Path, "/api/users-panel")
		if trimmed == "" {
			trimmed = "/"
		} else if !strings.HasPrefix(trimmed, "/") {
			trimmed = "/" + trimmed
		}
		req.URL.Path = trimmed
		req.URL.RawQuery = c.Request.URL.RawQuery
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}
