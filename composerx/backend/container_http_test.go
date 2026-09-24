package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newContentMakerRouter(uiDir string) *gin.Engine {
	return newContentMakerRouterAt(uiDir, "http://127.0.0.1:9")
}

func newContentMakerRouterAt(uiDir, panelURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	uiPaths := map[string]struct{}{}
	r := gin.New()
	r.Use(requireTranmailAccess(panelURL, uiPaths))
	r.GET("/health", (&App{}).handleHealth)
	r.GET("/templates", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.GET("/emails", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	mountContentMakerUI(r, uiDir, uiPaths)
	return r
}

func TestHTTPListenPort(t *testing.T) {
	t.Run("composerx port wins", func(t *testing.T) {
		t.Setenv("COMPOSERX_PORT", "8043")
		t.Setenv("PORT", "9090")
		if got := httpListenPort(); got != "8043" {
			t.Fatalf("listen port %q", got)
		}
	})
	t.Run("port fallback", func(t *testing.T) {
		t.Setenv("COMPOSERX_PORT", "")
		t.Setenv("PORT", "18043")
		if got := httpListenPort(); got != "18043" {
			t.Fatalf("listen port %q", got)
		}
	})
	t.Run("default", func(t *testing.T) {
		t.Setenv("COMPOSERX_PORT", "")
		t.Setenv("PORT", "")
		if got := httpListenPort(); got != "8043" {
			t.Fatalf("listen port %q", got)
		}
	})
}

func TestUIMountDoesNotRequireAuth(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<title>Content Maker</title>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "favicon.svg"), []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("console.log(1)"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := newContentMakerRouter(dir)
	for _, path := range []string{"/", "/favicon.svg", "/assets/app.js"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status %d", path, rec.Code)
		}
	}
	if rec := get(r, "/"); !strings.Contains(rec.Body.String(), "Content Maker") {
		t.Fatalf("GET / body %q", rec.Body.String())
	}

	if rec := get(r, "/templates"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /templates status %d", rec.Code)
	}

	rec := get(r, "/health")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /health status %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("health status %v", body["status"])
	}
}

func TestHeaderOnlyAdminIs401(t *testing.T) {
	r := newContentMakerRouter(t.TempDir())
	for _, path := range []string{"/templates", "/emails"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("X-User-Role", "admin")
		req.Header.Set("X-User-Roles", "admin")
		req.Header.Set("X-User-Permissions", "compose_email")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s status %d, want 401", path, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/templates", nil)
	req.Header.Set("Authorization", "Bearer not-a-token")
	req.Header.Set("X-User-Role", "admin")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad bearer status %d, want 401", rec.Code)
	}
}

func TestMorphBearerIgnoresSpoofedRole(t *testing.T) {
	allow := true
	morph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer good" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/permissions") {
			if allow {
				_, _ = w.Write([]byte(`{"permissions":["compose_email"]}`))
			} else {
				_, _ = w.Write([]byte(`{"permissions":[]}`))
			}
			return
		}
		if allow {
			_, _ = w.Write([]byte(`{"user":{"roles":["employee"]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"user":{"roles":["member"]}}`))
	}))
	t.Cleanup(morph.Close)

	r := newContentMakerRouterAt(t.TempDir(), morph.URL)
	req := httptest.NewRequest(http.MethodGet, "/templates", nil)
	req.Header.Set("Authorization", "Bearer good")
	req.Header.Set("X-User-Role", "admin")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("morph bearer status %d, want 200", rec.Code)
	}

	allow = false
	req = httptest.NewRequest(http.MethodGet, "/emails", nil)
	req.Header.Set("Authorization", "Bearer good")
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-User-Permissions", "compose_email")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("spoofed admin status %d, want 403", rec.Code)
	}
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
