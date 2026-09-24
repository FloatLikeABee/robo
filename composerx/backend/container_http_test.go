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
	gin.SetMode(gin.TestMode)
	uiPaths := map[string]struct{}{}
	r := gin.New()
	r.Use(requireTranmailAccess("http://127.0.0.1:9", uiPaths))
	r.GET("/health", (&App{}).handleHealth)
	r.GET("/templates", func(c *gin.Context) {
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

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
