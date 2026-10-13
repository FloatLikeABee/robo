package handlers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseModelJSONRawNewlinesInString(t *testing.T) {
	raw := "{\n  \"title\": \"Hi\",\n  \"markdown\": \"Line one\nLine two\"\n}"
	if json.Valid([]byte(raw)) {
		t.Fatal("fixture must be invalid JSON so the repair is what we test")
	}
	var out struct {
		Title    string `json:"title"`
		Markdown string `json:"markdown"`
	}
	if err := parseModelJSON(raw, &out); err != nil {
		t.Fatalf("parseModelJSON: %v", err)
	}
	if out.Title != "Hi" || out.Markdown != "Line one\nLine two" {
		t.Fatalf("out = %+v", out)
	}
}

func TestParseModelJSONFixtureWithFencesAndProse(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "ai_json", "raw_newline.json"))
	if err != nil {
		t.Fatal(err)
	}
	if json.Valid(raw) {
		t.Fatal("fixture must stay invalid until repair")
	}
	var out struct {
		Title     string          `json:"title"`
		Markdown  string          `json:"markdown"`
		Questions json.RawMessage `json:"questions"`
	}
	if err := parseModelJSON(string(raw), &out); err != nil {
		t.Fatalf("parseModelJSON fixture: %v", err)
	}
	if out.Title != "Morph Notes field guide" {
		t.Fatalf("title = %q", out.Title)
	}
	if !strings.Contains(out.Markdown, "signed-in human") || !strings.Contains(out.Markdown, "sharing pool") {
		t.Fatalf("markdown = %q", out.Markdown)
	}
	if !strings.Contains(out.Markdown, "\n") {
		t.Fatal("raw newline in the string was lost")
	}
}

func TestParseModelJSONMissingCommaAfterValue(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "ai_json", "qwen_missing_comma.json"))
	if err != nil {
		t.Fatal(err)
	}
	if json.Valid(raw) {
		t.Fatal("fixture must stay invalid until missing-comma repair")
	}
	var out struct {
		Title    string `json:"title"`
		Markdown string `json:"markdown"`
	}
	if err := parseModelJSON(string(raw), &out); err != nil {
		t.Fatalf("parseModelJSON missing comma: %v", err)
	}
	if out.Title != "Creating MorphNotes: a Field Guide" {
		t.Fatalf("title = %q", out.Title)
	}
	if !strings.Contains(out.Markdown, "sharing pools") {
		t.Fatalf("markdown = %q", out.Markdown)
	}
}

func TestParseModelJSONValidUnchanged(t *testing.T) {
	raw := `{"title":"Hi","markdown":"Line one\nLine two"}`
	var out struct {
		Title    string `json:"title"`
		Markdown string `json:"markdown"`
	}
	if err := parseModelJSON(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.Title != "Hi" || out.Markdown != "Line one\nLine two" {
		t.Fatalf("out = %+v", out)
	}
}
