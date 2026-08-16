package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/formsx/backend/internal/surveybot"
	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

// newFromFileRouter mounts the handler alone, bypassing the workspace-access
// middleware so these tests exercise upload validation rather than auth.
func newFromFileRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/survey-bot/templates/from-file", h.DraftSurveyBotTemplateFromFile)
	return r
}

// handlerWithAI returns a handler whose AI client is configured but never
// reached, so tests stop at validation instead of making network calls.
func handlerWithAI() *Handler {
	return &Handler{AI: morphai.NewClient(morphai.Config{
		APIKey:  "test-key",
		Model:   morphai.DefaultModel,
		BaseURL: morphai.DefaultBaseURL,
	})}
}

type uploadPart struct {
	field    string
	filename string
	mime     string
	content  []byte
}

func multipartBody(t *testing.T, fields map[string]string, parts ...uploadPart) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("write field %s: %v", k, err)
		}
	}
	for _, p := range parts {
		hdr := make(map[string][]string)
		hdr["Content-Disposition"] = []string{
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`, p.field, p.filename),
		}
		if p.mime != "" {
			hdr["Content-Type"] = []string{p.mime}
		}
		fw, err := w.CreatePart(hdr)
		if err != nil {
			t.Fatalf("create part: %v", err)
		}
		if _, err := fw.Write(p.content); err != nil {
			t.Fatalf("write part: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return body, w.FormDataContentType()
}

func postUpload(t *testing.T, h *Handler, fields map[string]string, parts ...uploadPart) *httptest.ResponseRecorder {
	t.Helper()
	body, contentType := multipartBody(t, fields, parts...)
	req := httptest.NewRequest(http.MethodPost, "/survey-bot/templates/from-file", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	newFromFileRouter(h).ServeHTTP(w, req)
	return w
}

func errorFrom(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	s, _ := out["error"].(string)
	return s
}

func TestFromFileRejectsTwoFiles(t *testing.T) {
	w := postUpload(t, handlerWithAI(), nil,
		uploadPart{field: "file", filename: "a.pdf", mime: "application/pdf", content: []byte("%PDF-1.4")},
		uploadPart{field: "file", filename: "b.pdf", mime: "application/pdf", content: []byte("%PDF-1.4")},
	)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400: %s", w.Code, w.Body.String())
	}
	if msg := errorFrom(t, w); !strings.Contains(msg, "one file") {
		t.Errorf("error should mention the one-file limit, got %q", msg)
	}
}

func TestFromFileRequiresAFile(t *testing.T) {
	w := postUpload(t, handlerWithAI(), map[string]string{"title_hint": "Nothing attached"})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400: %s", w.Code, w.Body.String())
	}
	if msg := errorFrom(t, w); !strings.Contains(msg, "required") {
		t.Errorf("error should say a file is required, got %q", msg)
	}
}

func TestFromFileRejectsUnsupportedType(t *testing.T) {
	w := postUpload(t, handlerWithAI(), nil,
		uploadPart{field: "file", filename: "plan.dwg", mime: "application/acad", content: []byte("dwg")},
	)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400: %s", w.Code, w.Body.String())
	}
	msg := errorFrom(t, w)
	if !strings.Contains(msg, "unsupported file type") {
		t.Errorf("error should name the problem, got %q", msg)
	}
	if !strings.Contains(msg, "PDF") || !strings.Contains(msg, "PNG") {
		t.Errorf("error should list accepted types, got %q", msg)
	}
}

func TestFromFileRejectsTextFileWithGuidance(t *testing.T) {
	// Markdown loads client-side; the endpoint should say so rather than accept it.
	w := postUpload(t, handlerWithAI(), nil,
		uploadPart{field: "file", filename: "survey.md", mime: "text/markdown", content: []byte("# hi")},
	)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400: %s", w.Code, w.Body.String())
	}
	if msg := errorFrom(t, w); !strings.Contains(strings.ToLower(msg), "editor") {
		t.Errorf("error should point at the editor path, got %q", msg)
	}
}

func TestFromFileReportsUnconfiguredAI(t *testing.T) {
	h := &Handler{AI: morphai.NewClient(morphai.Config{})}
	w := postUpload(t, h, nil,
		uploadPart{field: "file", filename: "a.pdf", mime: "application/pdf", content: []byte("%PDF-1.4")},
	)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503: %s", w.Code, w.Body.String())
	}
	if msg := errorFrom(t, w); !strings.Contains(msg, "MORPH_AI_API_KEY") {
		t.Errorf("error should name the missing setting, got %q", msg)
	}
}

func TestFromFileReportsPDFWithoutText(t *testing.T) {
	w := postUpload(t, handlerWithAI(), nil,
		uploadPart{field: "file", filename: "scan.pdf", mime: "application/pdf", content: []byte("not a real pdf at all")},
	)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400: %s", w.Code, w.Body.String())
	}
	if msg := errorFrom(t, w); msg == "" {
		t.Error("expected an explanatory error")
	}
}

func TestFromFileRejectsImageOnNativeEndpoint(t *testing.T) {
	h := &Handler{AI: morphai.NewClient(morphai.Config{
		APIKey:       "test-key",
		Model:        morphai.DefaultModel,
		APIURL:       morphai.DefaultAPIURL,
		UseNativeAPI: true,
	})}
	w := postUpload(t, h, nil,
		uploadPart{field: "file", filename: "form.png", mime: "image/png", content: []byte{0x89, 'P', 'N', 'G'}},
	)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503: %s", w.Code, w.Body.String())
	}
	if msg := errorFrom(t, w); !strings.Contains(msg, "MORPH_AI_API_URL") {
		t.Errorf("error should name the setting to change, got %q", msg)
	}
}

func TestVisionPromptAsksForANoContentSentinel(t *testing.T) {
	// The sentinel is how an unreadable image is turned into an error instead of
	// an invented survey, so the prompt must keep asking for it.
	if visionNoContentSentinel == "" {
		t.Fatal("the sentinel must not be empty")
	}
	if strings.Contains(surveyFromFileSystemPrompt, visionNoContentSentinel) {
		t.Error("the sentinel belongs in the vision prompt, not the generation prompt")
	}
}

func TestNormalizeSurveyMarkdownAddsFrontMatter(t *testing.T) {
	raw := "# Instructions\nAsk politely.\n\n## Q1 — Name\n- field: name\n- collect: text\n- required: true\n- prompt: Your name?"

	md := normalizeSurveyMarkdown(raw, "Customer Onboarding!")
	parsed, err := surveybot.ParseMarkdown(md)
	if err != nil {
		t.Fatalf("normalized markdown does not parse: %v\n%s", err, md)
	}
	if parsed.Title != "Customer Onboarding!" {
		t.Errorf("title hint ignored: %q", parsed.Title)
	}
	if parsed.Slug != "customer-onboarding" {
		t.Errorf("slug not normalized: %q", parsed.Slug)
	}
	if len(parsed.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(parsed.Steps))
	}
}

func TestNormalizeSurveyMarkdownStripsCodeFence(t *testing.T) {
	raw := "```markdown\n---\nslug: Feedback Survey\ntitle: Feedback\ntags: [a]\n---\n\n# Instructions\nGo.\n\n## Q1 — Name\n- field: name\n- collect: text\n- required: true\n- prompt: Name?\n```"

	md := normalizeSurveyMarkdown(raw, "")
	if strings.Contains(md, "```") {
		t.Fatalf("code fence survived:\n%s", md)
	}
	parsed, err := surveybot.ParseMarkdown(md)
	if err != nil {
		t.Fatalf("parse: %v\n%s", err, md)
	}
	if parsed.Slug != "feedback-survey" {
		t.Errorf("slug not normalized to kebab-case: %q", parsed.Slug)
	}
	if parsed.Tags == nil || parsed.Tags[0] != "a" {
		t.Errorf("existing tags should be preserved, got %v", parsed.Tags)
	}
}

func TestNormalizeSurveyMarkdownSlugFallbacks(t *testing.T) {
	// Non-ASCII title cannot produce a slug, so the generic fallback applies.
	md := normalizeSurveyMarkdown("## Q1 — A\n- field: a\n- collect: text\n- prompt: A?", "日本語")
	parsed, err := surveybot.ParseMarkdown(md)
	if err != nil {
		t.Fatalf("parse: %v\n%s", err, md)
	}
	if parsed.Slug != "survey" {
		t.Errorf("expected the fallback slug, got %q", parsed.Slug)
	}
	if parsed.Title != "日本語" {
		t.Errorf("title should be preserved verbatim, got %q", parsed.Title)
	}
}

func TestNormalizeSurveyMarkdownIgnoresInstructionsHeadingAsTitle(t *testing.T) {
	md := normalizeSurveyMarkdown("# Instructions\nGo.\n\n## Q1 — A\n- field: a\n- collect: text\n- prompt: A?", "")
	parsed, err := surveybot.ParseMarkdown(md)
	if err != nil {
		t.Fatalf("parse: %v\n%s", err, md)
	}
	if strings.EqualFold(parsed.Title, "instructions") {
		t.Errorf("section heading leaked into the title: %q", parsed.Title)
	}
}

func TestFromFileRouteIsRegistered(t *testing.T) {
	h := &Handler{AI: morphai.NewClient(morphai.Config{})}
	r := gin.New()
	h.Register(r.Group(""))

	found := false
	for _, route := range r.Routes() {
		if route.Method == http.MethodPost && route.Path == "/api/v1/survey-bot/templates/from-file" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("POST /api/v1/survey-bot/templates/from-file is not registered")
	}
}
