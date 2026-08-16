package surveybot

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed seeds/*.md
var seedFS embed.FS

// SeedFile is one embedded starter template.
type SeedFile struct {
	Slug     string
	Title    string
	Tags     []string
	Markdown string
	Summary  string
}

// LoadSeeds returns built-in Survey Bot markdown templates.
func LoadSeeds() ([]SeedFile, error) {
	entries, err := seedFS.ReadDir("seeds")
	if err != nil {
		return nil, err
	}
	out := make([]SeedFile, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := seedFS.ReadFile("seeds/" + e.Name())
		if err != nil {
			return nil, err
		}
		parsed, err := ParseMarkdown(string(raw))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		slug := parsed.Slug
		if slug == "" {
			slug = strings.TrimSuffix(e.Name(), ".md")
		}
		out = append(out, SeedFile{
			Slug:     slug,
			Title:    parsed.Title,
			Tags:     parsed.Tags,
			Markdown: string(raw),
			Summary:  truncateRunes(parsed.Instructions, 160),
		})
	}
	return out, nil
}

func truncateRunes(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= max {
		return string(r)
	}
	return string(r[:max]) + "…"
}
