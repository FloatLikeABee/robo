package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func skillImproveRouter(h *Handlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/skills/improve", h.ImproveSkill)
	return r
}

func postSkillImprove(h *Handlers, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/skills/improve", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	skillImproveRouter(h).ServeHTTP(w, req)
	return w
}

func TestImproveSkillRejectsMissingName(t *testing.T) {
	w := postSkillImprove(&Handlers{}, `{"name":"","instructions":"Do the thing"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "name") {
		t.Fatalf("expected name error, got %s", w.Body.String())
	}
}

func TestImproveSkillRejectsMissingInstructions(t *testing.T) {
	w := postSkillImprove(&Handlers{}, `{"name":"Lookup","instructions":""}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "instruction") {
		t.Fatalf("expected instructions error, got %s", w.Body.String())
	}
}

func TestImproveSkillRequiresAI(t *testing.T) {
	w := postSkillImprove(&Handlers{}, `{"name":"Lookup","instructions":"Use /full routes."}`)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestParseSkillImproveJSON(t *testing.T) {
	name, desc, instr, err := parseSkillImproveJSON("```json\n{\"name\":\"Concise\",\"description\":\"Short replies\",\"instructions\":\"Lead with the answer.\"}\n```")
	if err != nil {
		t.Fatal(err)
	}
	if name != "Concise" || desc != "Short replies" || instr != "Lead with the answer." {
		t.Fatalf("got name=%q desc=%q instr=%q", name, desc, instr)
	}
}

func TestParseSkillImproveJSONRequiresFields(t *testing.T) {
	if _, _, _, err := parseSkillImproveJSON(`{"name":"X"}`); err == nil {
		t.Fatal("expected error for missing instructions")
	}
}
