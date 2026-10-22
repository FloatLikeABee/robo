package htmldoc

import (
	"strings"
	"testing"
)

func TestDarkDocumentCSSUsesSansMultiAccentDarkTheme(t *testing.T) {
	css := DarkDocumentCSS()
	if strings.Contains(css, "Georgia") || strings.Contains(css, "#fafaf9") || strings.Contains(css, "#134e4a") {
		t.Fatal("dark document CSS must not use a serif face, a light page, or the old teal palette")
	}
	for _, needle := range []string{"Arial", "color-scheme:dark", "#38bdf8", "#a78bfa", "#34d399", "#070b14"} {
		if !strings.Contains(css, needle) {
			t.Fatalf("missing %q", needle)
		}
	}
}

func TestGeneratedHeaderUsesArialNotSystemUI(t *testing.T) {
	for _, css := range []string{DarkDocumentCSS(), DarkBigNoteDocumentCSS(), DarkCaseTaskDocumentCSS("")} {
		if !strings.Contains(css, `.meta{font:12px/1.4 `+sansStack) && !strings.Contains(css, `.meta{font:12px/1.4 Arial`) {
			t.Fatalf("header meta is not Arial:\n%s", css)
		}
		if strings.Contains(css, `.meta{font:12px/1.4 system-ui`) {
			t.Fatal("header meta still uses system-ui")
		}
	}
}
