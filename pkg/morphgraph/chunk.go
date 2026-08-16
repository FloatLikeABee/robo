package morphgraph

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	DefaultChunkRunes  = 900
	DefaultChunkOverlap = 120
)

// ChunkText splits text into overlapping rune windows.
func ChunkText(text string, size, overlap int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if size <= 0 {
		size = DefaultChunkRunes
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= size {
		overlap = size / 4
	}
	runes := []rune(text)
	if len(runes) <= size {
		return []string{string(runes)}
	}
	var out []string
	step := size - overlap
	for i := 0; i < len(runes); i += step {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		piece := strings.TrimSpace(string(runes[i:end]))
		if piece != "" {
			out = append(out, piece)
		}
		if end >= len(runes) {
			break
		}
	}
	return out
}

// ExtractPlainText turns common knowledge formats into searchable text.
func ExtractPlainText(filename, contentType, raw string) string {
	name := strings.ToLower(filename)
	ct := strings.ToLower(contentType)
	switch {
	case strings.HasSuffix(name, ".json") || strings.Contains(ct, "json"):
		return raw
	case strings.HasSuffix(name, ".csv") || strings.Contains(ct, "csv"):
		return raw
	case strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".txt") || strings.HasPrefix(ct, "text/"):
		return raw
	default:
		return raw
	}
}

// TruncateRunes shortens s to at most max runes.
func TruncateRunes(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max]) + "…"
}

// TokenOverlapScore is a simple lexical relevance score (0..1-ish).
func TokenOverlapScore(query, doc string) float64 {
	qTokens := tokenize(query)
	if len(qTokens) == 0 {
		return 0
	}
	d := " " + strings.ToLower(doc) + " "
	hit := 0
	for _, t := range qTokens {
		if strings.Contains(d, " "+t+" ") || strings.Contains(strings.ToLower(doc), t) {
			hit++
		}
	}
	return float64(hit) / float64(len(qTokens))
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	var b strings.Builder
	var out []string
	flush := func() {
		if b.Len() == 0 {
			return
		}
		tok := b.String()
		b.Reset()
		if len(tok) >= 2 {
			out = append(out, tok)
		}
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}

// CosineSim returns cosine similarity of two equal-length vectors.
func CosineSim(a, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		av := float64(a[i])
		bv := float64(b[i])
		dot += av * bv
		na += av * av
		nb += bv * bv
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (sqrt(na) * sqrt(nb))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 12; i++ {
		z = 0.5 * (z + x/z)
	}
	return z
}
