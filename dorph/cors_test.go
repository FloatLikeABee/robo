package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSAllowHeadersAreExplicit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(corsMiddleware())
	r.POST("/api/chat", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/chat", nil)
	req.Header.Set("Origin", "http://localhost:3031")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	got := w.Header().Get("Access-Control-Allow-Headers")
	if got == "*" {
		t.Fatalf("Access-Control-Allow-Headers is *, want an explicit list")
	}
	if strings.Contains(got, "X-User-ID") || strings.Contains(got, "X-User-Role") {
		t.Fatalf("Access-Control-Allow-Headers %q still allows identity headers", got)
	}
	for _, name := range []string{"Authorization", "Content-Type", "Accept"} {
		if !strings.Contains(got, name) {
			t.Fatalf("Access-Control-Allow-Headers %q missing %s", got, name)
		}
	}
}
