package spa

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMountServesEventsInfoAndNotAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	index := "<!doctype html><html><body>event-logs</body></html>"
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("console.log(1)"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.GET("/api/v1/events-info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"rows": []any{}})
	})
	if !Mount(r, dir) {
		t.Fatal("expected UI mount")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/events-info", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "event-logs") {
		t.Fatalf("events-info: %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/events-info", nil))
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), "event-logs") {
		t.Fatalf("api must not be the UI document: %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/missing", nil))
	if w.Code != http.StatusNotFound || strings.Contains(w.Body.String(), "event-logs") {
		t.Fatalf("api miss must not be the UI document: %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), "event-logs") {
		t.Fatalf("asset: %d %s", w.Code, w.Body.String())
	}
}

func TestMountSkippedWhenUIMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if Mount(r, t.TempDir()) {
		t.Fatal("missing index.html must not mount")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/events-info", nil))
	if w.Code == http.StatusOK && strings.Contains(w.Body.String(), "event-logs") {
		t.Fatalf("mounted an absent UI: %s", w.Body.String())
	}
}
