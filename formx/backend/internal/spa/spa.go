package spa

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// Mount serves a Vite dist. Client routes fall through to index.html.
// /api, /uploads, /swagger, and /health stay out of that document.
// A missing index.html leaves the router unchanged so a local API can run without the UI build.
func Mount(r *gin.Engine, dir string) bool {
	index := filepath.Join(dir, "index.html")
	info, err := os.Stat(index)
	if err != nil || info.IsDir() {
		return false
	}
	r.NoRoute(func(c *gin.Context) {
		if apiOrAssetRoot(c.Request.URL.Path) || (c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead) {
			c.Status(http.StatusNotFound)
			return
		}
		if serveFile(c, dir) {
			return
		}
		c.File(index)
	})
	return true
}

func apiOrAssetRoot(urlPath string) bool {
	clean := path.Clean("/" + strings.TrimPrefix(urlPath, "/"))
	switch {
	case clean == "/api" || strings.HasPrefix(clean, "/api/"):
		return true
	case clean == "/uploads" || strings.HasPrefix(clean, "/uploads/"):
		return true
	case clean == "/swagger" || strings.HasPrefix(clean, "/swagger/"):
		return true
	case clean == "/health":
		return true
	default:
		return false
	}
}

func serveFile(c *gin.Context, dir string) bool {
	clean := path.Clean(c.Request.URL.Path)
	rel := strings.TrimPrefix(clean, "/")
	if rel == "" || rel == "." || strings.Contains(rel, "..") {
		return false
	}
	full := filepath.Join(dir, filepath.FromSlash(rel))
	rooted, err := filepath.Rel(dir, full)
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
