package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSPAIconsAreSVGNotIndexHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "icons"), 0o755); err != nil {
		t.Fatal(err)
	}
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)
	for _, name := range []string{"bk-icon.svg", "morph-data-icon.svg", "morph-utils-icon.svg"} {
		if err := os.WriteFile(filepath.Join(dir, "icons", name), svg, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	index := []byte("<!doctype html><html><body>spa</body></html>")
	if err := os.WriteFile(filepath.Join(dir, "index.html"), index, 0o644); err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	mountFrontend(r, dir)

	for _, p := range []string{"/icons/bk-icon.svg", "/icons/morph-data-icon.svg", "/icons/morph-utils-icon.svg"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, p, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("%s status %d", p, w.Code)
		}
		ct := w.Header().Get("Content-Type")
		if !strings.HasPrefix(ct, "image/svg+xml") {
			t.Fatalf("%s content-type %q, want image/svg+xml", p, ct)
		}
		if strings.Contains(w.Body.String(), "<!doctype") {
			t.Fatalf("%s returned the SPA shell", p)
		}
		if w.Body.String() != string(svg) {
			t.Fatalf("%s body %q", p, w.Body.String())
		}
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/morphdata/research", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "<!doctype") {
		t.Fatalf("client route should be the SPA, got %d %q", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/etc/passwd", nil)
	req.URL.Path = "/etc/passwd"
	c.Request = req
	if serveBuildFile(c, dir) {
		t.Fatalf("escaped build dir: %s", w.Body.String())
	}
}

func TestSPAIconsFromRelativeBuildDir(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	build := filepath.Join(root, "frontend", "build")
	if err := os.MkdirAll(filepath.Join(build, "icons"), 0o755); err != nil {
		t.Fatal(err)
	}
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" id="rel"></svg>`)
	if err := os.WriteFile(filepath.Join(build, "icons", "bk-icon.svg"), svg, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(build, "index.html"), []byte("<!doctype html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	r := gin.New()
	mountFrontend(r, "./frontend/build")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/icons/bk-icon.svg", nil))
	if w.Code != http.StatusOK || w.Body.String() != string(svg) {
		t.Fatalf("relative build dir: %d %q type %q", w.Code, w.Body.String(), w.Header().Get("Content-Type"))
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/svg+xml") {
		t.Fatalf("content-type %q", ct)
	}
}
