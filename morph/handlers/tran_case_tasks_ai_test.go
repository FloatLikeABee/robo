package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func caseTaskAIDraftRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &Handlers{}
	r := gin.New()
	r.POST("/api/tran/case-tasks/ai-draft", h.CreateCaseTaskAIDraft)
	return r
}

func postCaseTaskAIDraftJSON(r http.Handler, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/case-tasks/ai-draft", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func postCaseTaskAIDraftFile(r http.Handler, prompt, filename string, fileBytes []byte) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("prompt", prompt)
	if filename != "" {
		part, err := mw.CreateFormFile("file", filename)
		if err != nil {
			panic(err)
		}
		_, _ = part.Write(fileBytes)
	}
	_ = mw.Close()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/tran/case-tasks/ai-draft", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)
	return w
}

func TestCreateCaseTaskAIDraftRejectsNeitherSource(t *testing.T) {
	r := caseTaskAIDraftRouter()
	w := postCaseTaskAIDraftJSON(r, `{"prompt":""}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "prompt") {
		t.Fatalf("expected prompt/file error, got %s", w.Body.String())
	}
}

func TestCreateCaseTaskAIDraftRejectsUnsupportedFile(t *testing.T) {
	r := caseTaskAIDraftRouter()
	w := postCaseTaskAIDraftFile(r, "", "photo.png", []byte("not-an-image"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "unsupported") {
		t.Fatalf("expected unsupported-type error, got %s", w.Body.String())
	}
}

func TestCreateCaseTaskAIDraftRejectsEmptyPDFExtract(t *testing.T) {
	r := caseTaskAIDraftRouter()
	w := postCaseTaskAIDraftFile(r, "", "scan.pdf", []byte("not a pdf"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	body := strings.ToLower(w.Body.String())
	if !strings.Contains(body, "pdf") && !strings.Contains(body, "extract") && !strings.Contains(body, "text") {
		t.Fatalf("expected PDF extract error, got %s", w.Body.String())
	}
}

func TestCreateCaseTaskAIDraftUnavailableWithoutAI(t *testing.T) {
	r := caseTaskAIDraftRouter()
	w := postCaseTaskAIDraftJSON(r, `{"prompt":"Inspect bus 12 at the north depot"}`)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestParseCaseTaskAIDraftOutput(t *testing.T) {
	raw := `{
		"title": "Depot inspection",
		"description": "Walk the north lot.",
		"start_at": "2026-08-27T09:00:00",
		"end_at": "2026-08-27T11:00:00",
		"location": {"label": "North depot", "area": [[39.1, -94.5], ["bad", 1], [40.2, -95.1]]},
		"detail": {"case_summary": {"priority": "high"}, "tags": ["fleet"]}
	}`
	got, err := parseCaseTaskAIDraftOutput(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Depot inspection" || got.Description != "Walk the north lot." {
		t.Fatalf("identity %#v", got)
	}
	if got.StartAt != "2026-08-27T09:00:00" || got.EndAt != "2026-08-27T11:00:00" {
		t.Fatalf("dates %#v", got)
	}
	if got.Location == nil || got.Location.Label != "North depot" {
		t.Fatalf("location %#v", got.Location)
	}
	if len(got.Location.Area) != 2 {
		t.Fatalf("expected invalid coordinates dropped, area=%v", got.Location.Area)
	}
	if got.Detail == nil {
		t.Fatal("detail missing")
	}
	if _, ok := got.Detail["case_summary"]; !ok {
		t.Fatalf("detail keys %v", got.Detail)
	}
}

func TestParseCaseTaskAIDraftOutputRequiresTitleAndDetailObject(t *testing.T) {
	if _, err := parseCaseTaskAIDraftOutput(`{"title":"","detail":{}}`); err == nil {
		t.Fatal("expected empty title error")
	}
	if _, err := parseCaseTaskAIDraftOutput(`{"title":"X","detail":[]}`); err == nil {
		t.Fatal("expected detail-must-be-object error")
	}
	if _, err := parseCaseTaskAIDraftOutput(`not json`); err == nil {
		t.Fatal("expected parse error")
	}
	got, err := parseCaseTaskAIDraftOutput(`{"title":"X","detail":{"a":1},"location":{"label":"","area":[["x","y"]]}}`)
	if err != nil {
		t.Fatal(err)
	}
	if got.Location != nil {
		t.Fatalf("invented location from invalid coords: %#v", got.Location)
	}
}

func TestExtractCaseTaskAIFileText(t *testing.T) {
	text, err := extractCaseTaskAIFileText("brief.txt", []byte("  Check route 7  "))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Check route 7") {
		t.Fatalf("text=%q", text)
	}
	csvText, err := extractCaseTaskAIFileText("jobs.csv", []byte("title,when\nInspect lot,morning\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(csvText, "Inspect lot") {
		t.Fatalf("csv text=%q", csvText)
	}
	if _, err := extractCaseTaskAIFileText("x.png", []byte("nope")); err == nil {
		t.Fatal("expected unsupported type")
	}
}
