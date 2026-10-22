package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestManagementPathAllowlist(t *testing.T) {
	allowed := []string{
		"/api/tran/members",
		"/api/forms/templates",
		"/api/formsx/events-info",
		"/api/sheetx/events-info",
		"/api/composerx/ai/mcp-tools",
	}
	for _, path := range allowed {
		if !allowedManagementPath(path) {
			t.Fatalf("expected %s to be allowed", path)
		}
	}
	if allowedManagementPath("/api/auth/me") {
		t.Fatal("expected auth path to be blocked")
	}
}

func TestParseManagementCall(t *testing.T) {
	call, err := parseManagementCall(`{"method":"GET","path":"/api/tran/members","query":"limit=5","body":null}`)
	if err != nil {
		t.Fatalf("parseManagementCall: %v", err)
	}
	if call.Method != "GET" || call.Path != "/api/tran/members" || call.Query != "limit=5" {
		t.Fatalf("unexpected call: %+v", call)
	}
	if call.Body != nil {
		t.Fatalf("expected nil body, got %s", string(call.Body))
	}
}

func TestSanitizeQueryForURL(t *testing.T) {
	got := sanitizeQueryForURL("search=hello world&limit=5\nignored=true")
	if !strings.Contains(got, "search=hello+world") || !strings.Contains(got, "limit=5") {
		t.Fatalf("unexpected sanitized query: %q", got)
	}
	if strings.Contains(got, "ignored") {
		t.Fatalf("expected newline noise to be removed: %q", got)
	}
}

func TestExecManagementAPIPathQueryNormalization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/tran/things", func(c *gin.Context) {
		if c.Query("search") != "hello world" {
			t.Fatalf("unexpected search query: %q", c.Query("search"))
		}
		if c.Query("limit") != "5" {
			t.Fatalf("unexpected limit query: %q", c.Query("limit"))
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	h := &Handlers{ginEngine: router}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/chat", nil)

	code, body := h.execManagementAPI(c, http.MethodGet, "/api/tran/things?search=hello world", "limit=5", nil)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", code, string(body))
	}
}
