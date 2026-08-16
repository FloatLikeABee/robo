package morphgraph

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	pdf "github.com/ledongthuc/pdf"
)

// ExtractPDFFile returns plain text from a PDF on disk.
func ExtractPDFFile(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	reader, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, reader); err != nil {
		return "", err
	}
	text := normalizeWhitespace(buf.String())
	if text == "" {
		return "", fmt.Errorf("no text extracted from PDF")
	}
	return text, nil
}

// ExtractPDFBytes writes bytes to a temp file then extracts text.
func ExtractPDFBytes(raw []byte) (string, error) {
	tmp, err := os.CreateTemp("", "morph-knowledge-*.pdf")
	if err != nil {
		return "", err
	}
	path := tmp.Name()
	defer os.Remove(path)
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return "", err
	}
	tmp.Close()
	return ExtractPDFFile(path)
}

func normalizeWhitespace(s string) string {
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
