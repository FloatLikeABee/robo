package ai

import (
	"strings"
	"testing"
)

func TestFormHTMLPromptAsksForDarkSansAccents(t *testing.T) {
	got := BuildFormHTMLPrompt(`{"UDGridSections":[]}`, "Intake", "")
	for _, needle := range []string{"Arial", "dark", "#8b5cf6", "#a78bfa", "#f97316", "html-anything"} {
		if !strings.Contains(got, needle) {
			t.Fatalf("missing %q", needle)
		}
	}
	if strings.Contains(got, "Dark Orange") || strings.Contains(got, "#FF8C00") {
		t.Fatal(got)
	}
	for _, banned := range []string{"cyan #38bdf8", "Cyan #38bdf8", "violet #a78bfa", "mint #34d399", "#38bdf8", "#34d399", "#070b14"} {
		if strings.Contains(got, banned) {
			t.Fatalf("colour name before hex %q in:\n%s", banned, got)
		}
	}
}
