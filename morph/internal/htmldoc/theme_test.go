package htmldoc

import (
	"strings"
	"testing"
)

func TestDarkDocumentCSSUsesPurpleGradient(t *testing.T) {
	css := DarkDocumentCSS()
	if strings.Contains(css, "#134e4a") || strings.Contains(css, "#2dd4bf") {
		t.Fatal("dark document CSS must not use teal/green palette")
	}
	if !strings.Contains(css, "#1e1b4b") || !strings.Contains(css, "#38bdf8") {
		t.Fatal("dark document CSS must use blue/purple palette")
	}
}
