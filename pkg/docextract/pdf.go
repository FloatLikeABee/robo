package docextract

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	pdf "github.com/ledongthuc/pdf"
)

// ErrNoText reports a PDF that parsed successfully but carries no text layer —
// typically a scan. Callers surface this as a per-file error rather than
// treating it as an empty document.
var ErrNoText = errors.New("no text could be extracted from the PDF (it may be a scan without a text layer)")

// ExtractPDF returns the text layer of a PDF on disk.
func ExtractPDF(path string) (text string, err error) {
	// ledongthuc/pdf panics on some malformed files.
	defer func() {
		if r := recover(); r != nil {
			text = ""
			err = fmt.Errorf("could not parse PDF: %v", r)
		}
	}()

	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("could not open PDF: %w", err)
	}
	defer f.Close()

	reader, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("could not read PDF text: %w", err)
	}
	var b strings.Builder
	if _, err := io.Copy(&b, reader); err != nil {
		return "", fmt.Errorf("could not read PDF text: %w", err)
	}

	out := CollapseWhitespace(b.String())
	if out == "" {
		return "", ErrNoText
	}
	return out, nil
}

// ExtractPDFBytes writes bytes to a temp file and extracts the text layer.
func ExtractPDFBytes(raw []byte) (string, error) {
	tmp, err := os.CreateTemp("", "docextract-*.pdf")
	if err != nil {
		return "", err
	}
	path := tmp.Name()
	defer os.Remove(path)

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	return ExtractPDF(path)
}
