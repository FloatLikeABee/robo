package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEventsInfoCollectionInfoRouteNotCapturedByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	protected := r.Group("")
	protected.GET("/events-info/collection-info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	protected.GET("/events-info/:id", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
	})

	req := httptest.NewRequest(http.MethodGet, "/events-info/collection-info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /events-info/collection-info: got %d body=%s", w.Code, w.Body.String())
	}
}
