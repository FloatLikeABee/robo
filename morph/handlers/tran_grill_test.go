package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGrillRequiresSessionAndSkipsTheModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := false
	h := &Handlers{grillComplete: func(context.Context, string) (string, error) {
		called = true
		return "", nil
	}}
	w := postGrill(h, "", `{"action":"start","kind":"note","title":"buy milk"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if called {
		t.Fatal("model was called")
	}
	ui := getGrillUI(h, "")
	if ui.Code != http.StatusUnauthorized {
		t.Fatalf("ui status %d", ui.Code)
	}
}

func TestGrillChattyDraftAndOtherSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handlers{grillComplete: func(_ context.Context, prompt string) (string, error) {
		if strings.Contains(prompt, "Write the body") {
			return "Great! I'm Morph AI, feel free to let me know.", nil
		}
		return "What should the agent notebook include?\nWho is it for?\n", nil
	}}
	start := postGrill(h, "ada", `{"action":"start","kind":"note","title":"make this morphAI a platform for agents, actually agent notebook"}`)
	if start.Code != http.StatusOK {
		t.Fatalf("start %d %s", start.Code, start.Body.String())
	}
	body := start.Body.String()
	if strings.Contains(body, "I'm Morph") || !strings.Contains(body, `"phase":"asking"`) {
		t.Fatalf("start body %s", body)
	}
	id := grillID(t, body)
	other := postGrill(h, "bea", `{"action":"finish","id":"`+id+`"}`)
	if other.Code != http.StatusNotFound || strings.Contains(other.Body.String(), "agent notebook") {
		t.Fatalf("other %d %s", other.Code, other.Body.String())
	}
	done := postGrill(h, "ada", `{"action":"finish","id":"`+id+`"}`)
	if done.Code != http.StatusOK {
		t.Fatalf("finish %d %s", done.Code, done.Body.String())
	}
	out := done.Body.String()
	if strings.Contains(out, "I'm Morph") || strings.Contains(strings.ToLower(out), "feel free") || !strings.Contains(out, exampleSnippet()) {
		t.Fatalf("draft %s", out)
	}
	ui := getGrillUI(h, "ada")
	if ui.Code != http.StatusOK || !strings.Contains(ui.Body.String(), "Arial") || !strings.Contains(ui.Body.String(), "note_grill") {
		t.Fatalf("ui %d len %d", ui.Code, ui.Body.Len())
	}
	if ui.Header().Get("Content-Type") == "" || strings.Contains(ui.Body.String(), "MORPH_MCP_TOKEN") {
		t.Fatal("ui content type or token")
	}
}

func exampleSnippet() string {
	return "agent notebook"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func grillID(t *testing.T, body string) string {
	t.Helper()
	const key = `"id":"`
	i := strings.Index(body, key)
	if i < 0 {
		t.Fatalf("no id in %s", body)
	}
	rest := body[i+len(key):]
	j := strings.Index(rest, `"`)
	if j <= 0 {
		t.Fatal(body)
	}
	return rest[:j]
}

func postGrill(h *Handlers, user, jsonBody string) *httptest.ResponseRecorder {
	r := gin.New()
	if user != "" {
		r.Use(func(c *gin.Context) {
			c.Set("auth_user_id", user)
			c.Next()
		})
	}
	r.POST("/api/tran/grill", h.TranGrill)
	req := httptest.NewRequest(http.MethodPost, "/api/tran/grill", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func getGrillUI(h *Handlers, user string) *httptest.ResponseRecorder {
	r := gin.New()
	if user != "" {
		r.Use(func(c *gin.Context) {
			c.Set("auth_user_id", user)
			c.Next()
		})
	}
	r.GET("/api/tran/grill/ui", h.TranGrillUI)
	req := httptest.NewRequest(http.MethodGet, "/api/tran/grill/ui", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
