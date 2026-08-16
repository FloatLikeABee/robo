// Package docextract turns uploaded files into AI-readable text.
//
// It defines which file types count as readable documents or images, extracts
// plain text from PDFs and text files, and caps extracted content so callers
// can keep AI prompts inside their budget.
package docextract

import (
	"path/filepath"
	"strings"
)

// Kind classifies an upload for extraction purposes.
type Kind string

const (
	KindDocument    Kind = "document"
	KindImage       Kind = "image"
	KindUnsupported Kind = "unsupported"
)

// DocumentExtensions are the file extensions extractable as text.
var DocumentExtensions = []string{".pdf", ".txt", ".md", ".markdown", ".csv"}

// ImageExtensions are the file extensions readable by a vision model.
var ImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

var documentMimes = map[string]bool{
	"application/pdf": true,
	"text/plain":      true,
	"text/markdown":   true,
	"text/csv":        true,
	"application/csv": true,
}

var imageMimes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// genericMimes carry no type information, so classification falls back to the extension.
var genericMimes = map[string]bool{
	"":                         true,
	"application/octet-stream": true,
	"binary/octet-stream":      true,
	"application/x-download":   true,
}

// Classify reports whether a file is an extractable document, a vision-readable
// image, or unsupported. The declared MIME type wins when it is meaningful;
// otherwise the filename extension decides.
func Classify(name, mime string) Kind {
	m := normalizeMime(mime)
	if !genericMimes[m] {
		if documentMimes[m] {
			return KindDocument
		}
		if imageMimes[m] {
			return KindImage
		}
		// A text/* type we do not name explicitly is still readable as text.
		if strings.HasPrefix(m, "text/") {
			return KindDocument
		}
		// An image/* subtype we cannot decode is not usable.
		if strings.HasPrefix(m, "image/") {
			return KindUnsupported
		}
	}
	return classifyByExtension(name)
}

func classifyByExtension(name string) Kind {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(name)))
	if ext == "" {
		return KindUnsupported
	}
	for _, e := range DocumentExtensions {
		if ext == e {
			return KindDocument
		}
	}
	for _, e := range ImageExtensions {
		if ext == e {
			return KindImage
		}
	}
	return KindUnsupported
}

// IsPDF reports whether a file should be parsed as a PDF.
func IsPDF(name, mime string) bool {
	if normalizeMime(mime) == "application/pdf" {
		return true
	}
	return strings.EqualFold(filepath.Ext(strings.TrimSpace(name)), ".pdf")
}

// AcceptedTypesMessage renders the accepted-type list for an error response, so
// a rejection always tells the user what would have worked.
func AcceptedTypesMessage(kinds ...Kind) string {
	var parts []string
	for _, k := range kinds {
		switch k {
		case KindDocument:
			parts = append(parts, "PDF, TXT, MD, CSV")
		case KindImage:
			parts = append(parts, "JPEG, PNG, GIF, WebP")
		}
	}
	if len(parts) == 0 {
		return "no file types are accepted"
	}
	return "accepted file types: " + strings.Join(parts, ", ")
}

func normalizeMime(mime string) string {
	m := strings.ToLower(strings.TrimSpace(mime))
	if i := strings.Index(m, ";"); i >= 0 {
		m = strings.TrimSpace(m[:i])
	}
	return m
}
