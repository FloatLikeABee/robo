package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/formsx/backend/internal/config"
	"github.com/gin-gonic/gin"
)

func TestHeaderRoleWithoutBearerIs401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{Cfg: &config.Config{UsersPanelBaseURL: "http://127.0.0.1:9"}}
	r := gin.New()
	h.Register(r.Group(""))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events-info", nil)
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-User-Roles", "admin")
	req.Header.Set("X-User-Permissions", "create_form,broadcast_form")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401, body %s", w.Code, w.Body.String())
	}
}

func TestSpoofedMorphBaseDoesNotAcceptHeaderRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"user":{"roles":["admin"]},"permissions":["create_form"]}`))
	}))
	t.Cleanup(attacker.Close)
	morph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(morph.Close)

	h := &Handler{Cfg: &config.Config{UsersPanelBaseURL: morph.URL}}
	r := gin.New()
	h.Register(r.Group(""))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events-info", nil)
	req.Header.Set("Authorization", "Bearer not-a-session")
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-UsersPanel-BaseURL", attacker.URL)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401, body %s", w.Code, w.Body.String())
	}
}

func TestMorphBearerAllowsProtectedRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	morph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer good-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/auth/user":
			_, _ = w.Write([]byte(`{"user":{"roles":["admin"]}}`))
		case "/api/auth/permissions":
			_, _ = w.Write([]byte(`{"permissions":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(morph.Close)

	h := &Handler{Cfg: &config.Config{UsersPanelBaseURL: morph.URL}}
	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set("handler_instance", h)
		c.Next()
	})
	api.Use(requireWorkspaceAccess())
	api.GET("/events-info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events-info", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	req.Header.Set("X-User-Role", "member")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("validated bearer status %d, body %s", w.Code, w.Body.String())
	}
}
