package handlers

import (
	"strings"
	"testing"
)

const phased = `A timeline of the work.

## Phase 1
- Built the facade
- Seeded the isle

## Phase 2
- REST API
`

func TestPhasedMarkdownBecomesCardsAndRailWithoutChart(t *testing.T) {
	html := markdownToHTMLFragment(phased)
	if strings.Count(html, `class="doc-card"`) != 2 {
		t.Fatalf("cards:\n%s", html)
	}
	if !strings.Contains(html, `class="rail"`) {
		t.Fatalf("rail missing:\n%s", html)
	}
	if strings.Contains(html, "<svg") {
		t.Fatalf("chart without stats:\n%s", html)
	}
}

func TestStatFenceBecomesChartWithMatchingBars(t *testing.T) {
	html := markdownToHTMLFragment(phased + "\n```stat\nPlaces | 5\nTalks | 20\n```\n")
	if !strings.Contains(html, ">5<") || !strings.Contains(html, ">20<") {
		t.Fatalf("stat chips:\n%s", html)
	}
	if !strings.Contains(html, `width="50"`) || !strings.Contains(html, `width="200"`) {
		t.Fatalf("bar widths:\n%s", html)
	}
	if strings.Count(html, `class="doc-card"`) != 2 {
		t.Fatalf("cards with stats:\n%s", html)
	}
}

func TestTimelinePromptAsksForShape(t *testing.T) {
	for _, text := range []string{documentShapePrompt} {
		if !strings.Contains(text, "## sections") || !strings.Contains(text, "stat fence") {
			t.Fatalf("shape prompt: %s", text)
		}
		if !strings.Contains(text, "Do not invent numbers") || !strings.Contains(text, "never raw HTML") {
			t.Fatalf("forbid lines: %s", text)
		}
	}
}
