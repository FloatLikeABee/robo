package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveBigNoteNeverHardFailsOnQwenReplies(t *testing.T) {
	fenced, err := os.ReadFile(filepath.Join("testdata", "ai_json", "qwen_fenced.md"))
	if err != nil {
		t.Fatal(err)
	}
	prose, err := os.ReadFile(filepath.Join("testdata", "ai_json", "qwen_non_json.txt"))
	if err != nil {
		t.Fatal(err)
	}
	idea := "PR152 run 1: field guide to Morph notes for agents: note kinds, the consent-based sharing pool, and the secret/PII block"

	out, err := resolveBigNoteFromReplies(idea, "dark", "note", []string{string(fenced), string(prose)})
	if err != nil {
		t.Fatalf("generate must not return a parse error: %v", err)
	}
	if !strings.Contains(out.Markdown, "sharing pool") && !strings.Contains(out.Markdown, "MorphNotes") {
		t.Fatalf("fallback markdown dropped the reply: %q", out.Markdown)
	}
	if strings.HasPrefix(strings.TrimSpace(out.Markdown), "```") {
		t.Fatalf("fences were left on: %q", out.Markdown)
	}
}

func TestResolveBigNoteDoesNotRetryWhenJSONParses(t *testing.T) {
	raw := `{"title":"Hi","markdown":"## One\n\nA.\n\n## Two\n\nB.\n\n## Three\n\nC."}`
	out, err := resolveBigNoteFromReplies("idea text", "dark", "note", []string{raw, "SHOULD NOT USE"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Title != "Hi" || !strings.Contains(out.Markdown, "## One") {
		t.Fatalf("out = %+v", out)
	}
	if strings.Contains(out.Markdown, "SHOULD NOT USE") {
		t.Fatal("used the retry reply after a good parse")
	}
}
