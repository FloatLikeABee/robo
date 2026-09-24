package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// mountFrontend serves the CRA build. Webpack chunks live under /static.
// Files copied from public/ (icons, manifest, robots) sit at the build root.
// A path with no file falls through to index.html so client routes keep working.
func mountFrontend(r *gin.Engine, buildDir string) {
	r.Static("/static", filepath.Join(buildDir, "static"))
	r.StaticFile("/", filepath.Join(buildDir, "index.html"))
	r.NoRoute(func(c *gin.Context) {
		if serveBuildFile(c, buildDir) {
			return
		}
		c.File(filepath.Join(buildDir, "index.html"))
	})
}

// serveBuildFile writes a real file from the CRA build root.
// False means the SPA shell should answer (client route, API miss, or directory).
func serveBuildFile(c *gin.Context, buildDir string) bool {
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		return false
	}
	clean := path.Clean(c.Request.URL.Path)
	if clean == "/" || strings.HasPrefix(clean, "/api/") || strings.HasPrefix(clean, "/swagger") {
		return false
	}
	rel := strings.TrimPrefix(clean, "/")
	if rel == "" || rel == "." || strings.Contains(rel, "..") {
		return false
	}
	full := filepath.Join(buildDir, filepath.FromSlash(rel))
	rooted, err := filepath.Rel(buildDir, full)
	if err != nil || rooted == ".." || strings.HasPrefix(rooted, ".."+string(filepath.Separator)) {
		return false
	}
	info, err := os.Stat(full)
	if err != nil || info.IsDir() {
		return false
	}
	c.File(full)
	return true
}
