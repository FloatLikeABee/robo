package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func httpListenPort() string {
	if p := strings.TrimSpace(os.Getenv("COMPOSERX_PORT")); p != "" {
		return p
	}
	if p := strings.TrimSpace(os.Getenv("PORT")); p != "" {
		return p
	}
	return "8043"
}

func uiPathOpen(uiPaths map[string]struct{}, path string) bool {
	if _, ok := uiPaths[path]; ok {
		return true
	}
	if strings.HasPrefix(path, "/assets/") {
		_, ok := uiPaths["/assets/"]
		return ok
	}
	return false
}

// mountContentMakerUI serves the Vite build at the same origin as the API.
// It registers only files that exist. It does not catch unknown paths.
func mountContentMakerUI(r *gin.Engine, dir string, uiPaths map[string]struct{}) {
	dir = strings.TrimSpace(dir)
	if dir == "" || uiPaths == nil {
		return
	}
	index := filepath.Join(dir, "index.html")
	st, err := os.Stat(index)
	if err != nil || st.IsDir() {
		return
	}
	r.StaticFile("/", index)
	r.StaticFile("/index.html", index)
	uiPaths["/"] = struct{}{}
	uiPaths["/index.html"] = struct{}{}

	assets := filepath.Join(dir, "assets")
	if st, err := os.Stat(assets); err == nil && st.IsDir() {
		r.Static("/assets", assets)
		uiPaths["/assets/"] = struct{}{}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || name == "index.html" || strings.Contains(name, "/") {
			continue
		}
		p := "/" + name
		r.StaticFile(p, filepath.Join(dir, name))
		uiPaths[p] = struct{}{}
	}
}
