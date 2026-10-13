package morphai

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestHTMLSkillStaysUnderOllamaBudget(t *testing.T) {
	got := HTMLSkillInstructions()
	if got == "" {
		t.Fatal("empty HTML skill")
	}
	if n := EstimateTokens(got); n > HTMLSkillTokenBudget {
		t.Fatalf("injected skill is %d tokens, budget is %d", n, HTMLSkillTokenBudget)
	}
	if EstimateTokens(got) != (utf8.RuneCountInString(got)+3)/4 {
		t.Fatal("EstimateTokens must be runes/4 rounding up")
	}
	low := strings.ToLower(got)
	for _, needle := range []string{"html-anything", "arial", "dark", "#38bdf8", "#a78bfa"} {
		if !strings.Contains(low, strings.ToLower(needle)) {
			t.Fatalf("missing %q in:\n%s", needle, got)
		}
	}
}

func TestWithHTMLSkillAttachesInstructions(t *testing.T) {
	got := WithHTMLSkill("Title: demo")
	if !strings.HasPrefix(got, "Title: demo") {
		t.Fatalf("caller text must stay first:\n%s", got)
	}
	if !strings.Contains(got, HTMLSkillInstructions()) {
		t.Fatal("missing compact skill")
	}
}

func TestHTMLSkillVendorTree(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, "skills", "vendor", "html-anything")
	license, err := os.ReadFile(filepath.Join(dir, "LICENSE"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(license), "Apache License") || !strings.Contains(string(license), "Version 2.0") {
		t.Fatal("LICENSE is not Apache-2.0")
	}
	notice, err := os.ReadFile(filepath.Join(dir, "NOTICE"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(notice), "nexu-io/html-anything") || !strings.Contains(string(notice), "https://github.com/nexu-io/html-anything") {
		t.Fatalf("NOTICE missing attribution:\n%s", notice)
	}
	if _, err := os.Stat(filepath.Join(dir, "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	condensed, err := os.ReadFile(filepath.Join(dir, "condensed.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(condensed) != htmlSkillCondensed {
		t.Fatal("skills/vendor/html-anything/condensed.md must match the morphai embed")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "docs", "agents", "16-morph-for-agents.md")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}
