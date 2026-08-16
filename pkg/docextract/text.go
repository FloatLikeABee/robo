package docextract

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ExtractText reads TXT, MD, or CSV bytes as UTF-8 text. Invalid byte sequences
// are replaced rather than failing, and line structure is preserved so CSV
// header and data rows survive.
func ExtractText(raw []byte) string {
	s := string(raw)
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "\uFFFD")
	}
	return normalizeLines(s)
}

// normalizeLines trims trailing space per line and collapses runs of blank
// lines, keeping newlines intact.
func normalizeLines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	blank := 0
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if strings.TrimSpace(trimmed) == "" {
			blank++
			if blank > 1 {
				continue
			}
			out = append(out, "")
			continue
		}
		blank = 0
		out = append(out, trimmed)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// CollapseWhitespace flattens all whitespace runs to single spaces. PDF text
// extraction produces ragged spacing, and line breaks there carry no meaning.
func CollapseWhitespace(s string) string {
	var b strings.Builder
	lastSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
			continue
		}
		lastSpace = false
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}
