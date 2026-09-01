package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	morphdb "idongivaflyinfa/db"

	"github.com/gin-gonic/gin"
)

func TestBuildCaseTaskMarkdownTitleDescriptionAndJSON(t *testing.T) {
	md := buildCaseTaskMarkdown(caseTaskViewDoc{
		Title:       "Route delay",
		Description: "Check the AM window.",
		DetailJSON:  `{"priority":"high"}`,
	})
	if !strings.Contains(md, "# Route delay") {
		t.Fatalf("expected title heading, got:\n%s", md)
	}
	if !strings.Contains(md, "Check the AM window.") {
		t.Fatalf("expected description body, got:\n%s", md)
	}
	if !strings.Contains(md, "## Detail") || !strings.Contains(md, "**Priority:** high") {
		t.Fatalf("expected structured detail labels, got:\n%s", md)
	}
	if strings.Contains(md, "```json") {
		t.Fatalf("must not dump detail as a json fence, got:\n%s", md)
	}
	if strings.Contains(md, "**Assignees:**") {
		t.Fatalf("empty assignees must be omitted, got:\n%s", md)
	}
	if strings.Contains(md, "**Start:**") || strings.Contains(md, "**End:**") {
		t.Fatalf("empty times must be omitted, got:\n%s", md)
	}
	if strings.Contains(md, "**Location:**") {
		t.Fatalf("empty location must be omitted, got:\n%s", md)
	}
}

func TestBuildCaseTaskMarkdownIncludesOptionalFields(t *testing.T) {
	start := time.Date(2026, 8, 31, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 11, 30, 0, 0, time.UTC)
	md := buildCaseTaskMarkdown(caseTaskViewDoc{
		Title:      "Route delay",
		Assignees:  "Ada (employee)",
		StartAt:    &start,
		EndAt:      &end,
		Location:   `{"label":"Depot A","area":[]}`,
		DetailJSON: `{}`,
	})
	if !strings.Contains(md, "**Assignees:** Ada (employee)") {
		t.Fatalf("expected assignees line, got:\n%s", md)
	}
	if !strings.Contains(md, "**Start:** 2026-08-31 09:00") {
		t.Fatalf("expected start line, got:\n%s", md)
	}
	if !strings.Contains(md, "**End:** 2026-08-31 11:30") {
		t.Fatalf("expected end line, got:\n%s", md)
	}
	if !strings.Contains(md, "**Location:** Depot A") {
		t.Fatalf("expected location label, got:\n%s", md)
	}
}

func TestBuildCaseTaskMarkdownEmptyTitleFallback(t *testing.T) {
	md := strings.TrimSpace(buildCaseTaskMarkdown(caseTaskViewDoc{}))
	if !strings.HasPrefix(md, "# Case/task") {
		t.Fatalf("expected fallback heading, got:\n%s", md)
	}
}

func TestBuildCaseTaskMarkdownNestedDetailAndList(t *testing.T) {
	md := buildCaseTaskMarkdown(caseTaskViewDoc{
		Title:      "Route delay",
		DetailJSON: `{"case_summary":{"priority":"high"},"tags":["fleet"]}`,
	})
	if !strings.Contains(md, "### Case Summary") {
		t.Fatalf("expected nested heading for case_summary, got:\n%s", md)
	}
	if !strings.Contains(md, "**Priority:** high") {
		t.Fatalf("expected nested priority label, got:\n%s", md)
	}
	if !strings.Contains(md, "### Tags") || !strings.Contains(md, "- fleet") {
		t.Fatalf("expected tags list, got:\n%s", md)
	}
	if strings.Contains(md, "```json") {
		t.Fatalf("must not dump nested detail as a json fence, got:\n%s", md)
	}
}

func TestBuildCaseTaskMarkdownEmptyDetailOmitsSection(t *testing.T) {
	for _, raw := range []string{"", "{}", "  "} {
		md := buildCaseTaskMarkdown(caseTaskViewDoc{Title: "Route delay", DetailJSON: raw})
		if strings.Contains(md, "## Detail") {
			t.Fatalf("empty detail %q must omit Detail section, got:\n%s", raw, md)
		}
		if strings.Contains(md, "```json") || strings.Contains(md, "{}") {
			t.Fatalf("empty detail %q must not dump {}, got:\n%s", raw, md)
		}
	}
}

func TestBuildCaseTaskMarkdownInvalidJSONShowsRaw(t *testing.T) {
	md := buildCaseTaskMarkdown(caseTaskViewDoc{
		Title:       "Route delay",
		Description: "Check the AM window.",
		DetailJSON:  `{not-json`,
	})
	if !strings.Contains(md, "# Route delay") || !strings.Contains(md, "Check the AM window.") {
		t.Fatalf("expected title and description, got:\n%s", md)
	}
	if !strings.Contains(md, "{not-json") {
		t.Fatalf("expected raw invalid detail, got:\n%s", md)
	}
	if strings.Contains(md, "```json") {
		t.Fatalf("invalid detail must not be a json fence, got:\n%s", md)
	}
}

func TestBuildCaseTaskHTMLDarkPreview(t *testing.T) {
	html := buildCaseTaskHTML(caseTaskViewDoc{
		Title:       "Route delay",
		Description: "Check the AM window.",
		DetailJSON:  `{"priority":"high"}`,
	})
	if !strings.Contains(html, "#0b1220") {
		t.Fatalf("expected dark background, got snippet:\n%s", html[:min(len(html), 400)])
	}
	if !strings.Contains(html, "color-scheme:dark") && !strings.Contains(html, "color-scheme: dark") {
		t.Fatalf("expected dark color-scheme for iframe scrollbars, got snippet:\n%s", html[:min(len(html), 500)])
	}
	if !strings.Contains(html, "scrollbar-color") {
		t.Fatalf("expected themed scrollbar-color, got snippet:\n%s", html[:min(len(html), 500)])
	}
	if !strings.Contains(html, "Route delay") {
		t.Fatalf("expected title in HTML, got snippet:\n%s", html[:min(len(html), 800)])
	}
	if !strings.Contains(html, "Case/task") {
		t.Fatalf("expected Case/task label, got snippet:\n%s", html[:min(len(html), 800)])
	}
	if !strings.Contains(html, "Check the AM window.") {
		t.Fatalf("expected description in HTML")
	}
	if !strings.Contains(html, `class="detail-ui"`) {
		t.Fatalf("expected structured detail UI, got snippet:\n%s", html[:min(len(html), 1200)])
	}
	if !strings.Contains(html, "Priority") || !strings.Contains(html, "high") {
		t.Fatalf("expected Priority / high in HTML detail UI, got snippet:\n%s", html[:min(len(html), 1200)])
	}
	if strings.Contains(html, `<pre>`) && strings.Contains(html, `"priority"`) {
		t.Fatalf("must not dump pretty JSON in pre/code, got snippet:\n%s", html[:min(len(html), 1200)])
	}
}

func TestBuildCaseTaskHTMLEmptyDetailOmitsSection(t *testing.T) {
	html := buildCaseTaskHTML(caseTaskViewDoc{Title: "Route delay", DetailJSON: `{}`})
	if strings.Contains(html, `class="detail-ui"`) {
		t.Fatalf("empty detail must omit detail-ui, got snippet:\n%s", html[:min(len(html), 800)])
	}
	if strings.Contains(html, `<pre>`) {
		t.Fatalf("empty detail must not dump JSON in pre/code, got snippet:\n%s", html[:min(len(html), 800)])
	}
}

func TestBuildCaseTaskHTMLInvalidJSONShowsRaw(t *testing.T) {
	html := buildCaseTaskHTML(caseTaskViewDoc{
		Title:       "Route delay",
		Description: "Check the AM window.",
		DetailJSON:  `{not-json`,
	})
	if !strings.Contains(html, "Route delay") || !strings.Contains(html, "Check the AM window.") {
		t.Fatalf("expected title and description, got snippet:\n%s", html[:min(len(html), 800)])
	}
	if !strings.Contains(html, "{not-json") {
		t.Fatalf("expected raw invalid detail in HTML, got snippet:\n%s", html[:min(len(html), 1200)])
	}
	if !strings.Contains(html, `class="detail-fallback"`) {
		t.Fatalf("expected detail-fallback for invalid JSON, got snippet:\n%s", html[:min(len(html), 1200)])
	}
}

func TestGetCaseTaskFullIncludesComputedMarkdownAndHTML(t *testing.T) {
	sqlDB := openLegacyCaseTaskDB(t)
	gin.SetMode(gin.TestMode)
	h := &Handlers{TranMySQL: &morphdb.TranSQL{DB: sqlDB}, EntityDetails: &memEntityDetails{}}
	r := gin.New()
	r.POST("/api/tran/case-tasks", h.CreateCaseTask)
	r.GET("/api/tran/case-tasks/:id/full", h.GetCaseTaskFull)

	created := postCreateCaseTask(r, `{"title":"Route delay","description":"Check the AM window.","start_at":"2026-09-01T00:00","end_at":"2026-09-01T23:59","detail":{"priority":"high"}}`)
	if created.Code != http.StatusOK {
		t.Fatalf("create status %d body %s", created.Code, created.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(created.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	id, _ := out["id"].(float64)
	if id <= 0 {
		t.Fatalf("expected id, got %s", created.Body.String())
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/tran/case-tasks/%d/full", int(id)), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("full status %d body %s", w.Code, w.Body.String())
	}
	var full map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &full); err != nil {
		t.Fatal(err)
	}
	md, _ := full["markdown_content"].(string)
	html, _ := full["html_content"].(string)
	if !strings.Contains(md, "# Route delay") || !strings.Contains(md, "Check the AM window.") || !strings.Contains(md, "**Priority:** high") || strings.Contains(md, "```json") {
		t.Fatalf("expected computed structured markdown, got %q", md)
	}
	if !strings.Contains(html, `class="detail-ui"`) || !strings.Contains(html, "Priority") {
		t.Fatalf("expected computed html detail UI, got snippet %q", html[:min(len(html), 400)])
	}
	if !strings.Contains(html, "#0b1220") || !strings.Contains(html, "Route delay") {
		t.Fatalf("expected computed html, got snippet %q", html[:min(len(html), 400)])
	}
}
