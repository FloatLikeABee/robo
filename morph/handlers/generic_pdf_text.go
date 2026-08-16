package handlers

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	pdfLiteralStringRe = regexp.MustCompile(`\((?:\\.|[^\\)])*\)`)
	pdfHexStringRe     = regexp.MustCompile(`<[0-9A-Fa-f]+>`)
)

// pdfBytesToMarkdown extracts readable text from PDF bytes (local, no external services)
// or reads markdown files directly. Output is stored as content_markdown.
func pdfBytesToMarkdown(filename string, raw []byte) (string, error) {
	ext := strings.ToLower(path.Ext(filename))
	if ext == ".md" || ext == ".markdown" {
		md := strings.TrimSpace(strings.ToValidUTF8(string(raw), ""))
		if md == "" {
			return "", fmt.Errorf("markdown file is empty")
		}
		return md, nil
	}

	text, err := extractPDFTextLocal(raw)
	if err != nil {
		return "", err
	}
	return plainTextToMarkdown(text), nil
}

func extractPDFTextLocal(raw []byte) (string, error) {
	if len(raw) < 5 || !bytes.HasPrefix(raw, []byte("%PDF")) {
		return "", fmt.Errorf("not a valid PDF file")
	}

	seen := make(map[string]struct{})
	var lines []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || !isLikelyPDFText(s) {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		lines = append(lines, s)
	}

	for _, m := range pdfLiteralStringRe.FindAll(raw, -1) {
		if len(m) < 2 {
			continue
		}
		add(decodePDFLiteralString(string(m[1 : len(m)-1])))
	}
	for _, m := range pdfHexStringRe.FindAll(raw, -1) {
		if len(m) < 2 {
			continue
		}
		add(decodePDFHexString(string(m[1 : len(m)-1])))
	}

	out := strings.Join(lines, "\n")
	out = collapsePDFTextWhitespace(out)
	if strings.TrimSpace(out) == "" {
		return "", fmt.Errorf("no extractable text in PDF (image-only PDFs are not supported)")
	}
	return out, nil
}

func decodePDFLiteralString(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		if i+1 >= len(s) {
			break
		}
		i++
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case '\\', '(', ')':
			b.WriteByte(s[i])
		case '\n', '\r':
			// line continuation in PDF strings
		default:
			if s[i] >= '0' && s[i] <= '7' {
				end := i
				for end < len(s) && end < i+3 && s[end] >= '0' && s[end] <= '7' {
					end++
				}
				var val byte
				fmt.Sscanf(s[i:end], "%o", &val)
				b.WriteByte(val)
				i = end - 1
			} else {
				b.WriteByte(s[i])
			}
		}
	}
	return b.String()
}

func decodePDFHexString(hex string) string {
	hex = strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f') {
			return r
		}
		return -1
	}, hex)
	if len(hex) < 2 {
		return ""
	}
	if len(hex)%2 == 1 {
		hex = hex + "0"
	}
	var b strings.Builder
	for i := 0; i+1 < len(hex); i += 2 {
		var v byte
		fmt.Sscanf(hex[i:i+2], "%02x", &v)
		if v != 0 {
			b.WriteByte(v)
		}
	}
	return strings.ToValidUTF8(b.String(), "")
}

func isLikelyPDFText(s string) bool {
	if len(s) == 0 || len(s) > 8000 {
		return false
	}
	if strings.HasPrefix(s, "/") || strings.Contains(s, "Arial") && len(s) < 12 {
		return false
	}
	printable := 0
	for _, r := range s {
		if r == '\n' || r == '\t' || unicode.IsPrint(r) {
			printable++
		}
	}
	if printable*10 < len(s)*7 {
		return false
	}
	if !utf8.ValidString(s) {
		return false
	}
	letters := 0
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			letters++
		}
	}
	return letters >= 2
}

func collapsePDFTextWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func plainTextToMarkdown(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	lines := strings.Split(text, "\n")
	var b strings.Builder
	titleSet := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			b.WriteString("\n")
			continue
		}
		if !titleSet {
			b.WriteString("# ")
			b.WriteString(line)
			b.WriteString("\n\n")
			titleSet = true
			continue
		}
		if looksLikeMarkdownHeading(line) {
			if strings.HasPrefix(line, "#") {
				b.WriteString(line)
			} else {
				b.WriteString("## ")
				b.WriteString(line)
			}
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

func looksLikeMarkdownHeading(line string) bool {
	if strings.HasPrefix(line, "#") {
		return true
	}
	if len(line) < 80 && !strings.HasSuffix(line, ".") && strings.ToUpper(line) == line && strings.Contains(line, " ") {
		return true
	}
	return false
}
