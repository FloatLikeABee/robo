package ai

import (
	"strings"
	"testing"
)

func TestFormHTMLPromptAsksForDarkSansAccents(t *testing.T) {
	got := BuildFormHTMLPrompt(`{"UDGridSections":[]}`, "Intake", "")
	for _, needle := range []string{"Arial", "dark", "#38bdf8", "#a78bfa", "#34d399", "html-anything"} {
		if !strings.Contains(got, needle) {
			t.Fatalf("missing %q", needle)
		}
	}
	if strings.Contains(got, "Dark Orange") || strings.Contains(got, "#FF8C00") {
		t.Fatal(got)
	}
	for _, banned := range []string{"cyan #38bdf8", "Cyan #38bdf8", "violet #a78bfa", "mint #34d399"} {
		if strings.Contains(got, banned) {
			t.Fatalf("colour name before hex %q in:\n%s", banned, got)
		}
	}
}
