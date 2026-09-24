package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestChatHandlerEmptyUserIs401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handlers{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/chat", strings.NewReader(`{"message":"hi"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.ChatHandler(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401, body %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), `"admin"`) {
		t.Fatalf("empty user id was treated as admin: %s", w.Body.String())
	}
}
