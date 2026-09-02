package handlers

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/robo/docextract"
)

// pdfBytesToMarkdown extracts the PDF text layer (all pages) or reads markdown files.
func pdfBytesToMarkdown(filename string, raw []byte) (string, error) {
	ext := strings.ToLower(path.Ext(filename))
	if ext == ".md" || ext == ".markdown" {
		md := strings.TrimSpace(strings.ToValidUTF8(string(raw), ""))
		if md == "" {
			return "", fmt.Errorf("markdown file is empty")
		}
		return md, nil
	}

	text, err := docextract.ExtractPDFBytes(raw)
	if err != nil {
		if errors.Is(err, docextract.ErrNoText) {
			return "", fmt.Errorf("no extractable text in PDF (image-only PDFs are not supported)")
		}
		return "", err
	}
	return plainTextToMarkdown(text), nil
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
