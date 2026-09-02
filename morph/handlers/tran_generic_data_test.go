package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGenericDataExtToType(t *testing.T) {
	cases := map[string]string{
		".csv":  "csv",
		".xlsx": "csv",
		".json": "json",
		".pdf":  "pdf",
		".md":   "pdf",
	}
	for ext, want := range cases {
		got, ok := genericDataExtToType(ext)
		if !ok || got != want {
			t.Errorf("genericDataExtToType(%q)=%q,%v want %q,true", ext, got, ok, want)
		}
	}
	if _, ok := genericDataExtToType(".xls"); ok {
		t.Error("legacy .xls should be rejected")
	}
	if _, ok := genericDataExtToType(".docx"); ok {
		t.Error(".docx should be rejected")
	}
}

func TestPdfBytesToMarkdownMarkdownFile(t *testing.T) {
	md, err := pdfBytesToMarkdown("notes.md", []byte("# Hello\n\nBody"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, "# Hello") {
		t.Fatalf("unexpected markdown %q", md)
	}
	if _, err := pdfBytesToMarkdown("empty.md", []byte("  ")); err == nil {
		t.Fatal("expected empty markdown error")
	}
}

func TestPdfBytesToMarkdownInvalidPDF(t *testing.T) {
	if _, err := pdfBytesToMarkdown("scan.pdf", []byte("not a pdf")); err == nil {
		t.Fatal("expected invalid PDF error")
	}
}

func TestParseCSVBytes(t *testing.T) {
	rows, cols, err := parseCSVBytes([]byte("name,qty\nbus,2\nvan,1\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 2 || cols[0] != "name" {
		t.Fatalf("cols=%v", cols)
	}
	if len(rows) != 2 || rows[0]["name"] != "bus" {
		t.Fatalf("rows=%v", rows)
	}
}

func TestParseGenericDataImportInvalidXLSX(t *testing.T) {
	h := &Handlers{}
	_, _, err := h.parseGenericDataImport(nil, "csv", "sheet.xlsx", []byte("not zip"))
	if err == nil {
		t.Fatal("expected xlsx parse error")
	}
}

func TestExtractJSONPrompt(t *testing.T) {
	p := extractJSONPrompt("asset_detail", "Yellow bus 42")
	if !strings.Contains(p, "Yellow bus 42") || !strings.Contains(p, "asset") {
		t.Fatalf("prompt missing context: %s", p)
	}
}

func TestExtractJSONFromTextValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handlers{}
	r := gin.New()
	r.POST("/api/tran/extract-json", h.ExtractJSONFromText)

	post := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/tran/extract-json", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		return w
	}

	if w := post(`{"text":"hi","purpose":"nope"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("bad purpose: status %d body %s", w.Code, w.Body.String())
	}
	if w := post(`{"text":"","purpose":"generic_data"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("empty text: status %d body %s", w.Code, w.Body.String())
	}
	if w := post(`{"text":"a bus","purpose":"generic_data"}`); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("no AI: status %d body %s", w.Code, w.Body.String())
	}
}
