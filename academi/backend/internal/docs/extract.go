package docs

import (
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

func extractPDFText(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return "", err
	}
	r, err := pdf.NewReader(f, st.Size())
	if err != nil {
		return "", err
	}
	pr, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(pr)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
