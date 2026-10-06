package handlers

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

const documentShapePrompt = `
Document shape (markdown only, never raw HTML):
- Open with a one-paragraph lede.
- Use ## sections. Put milestone bullets under a section.
- Use a blockquote only for a constraint or warning that is in the source.
- If the source states quantities, add one stat fence and no others. Example:
` + "```stat\nPlaces | 5\nTalks | 20\n```" + `
- Do not invent numbers. If the source has no quantities, do not add a stat fence.
`

var (
	statFenceRE = regexp.MustCompile(`(?s)<pre><code class="language-stat">(.*?)</code></pre>`)
	quoteRE     = regexp.MustCompile(`(?s)<blockquote>(.*?)</blockquote>`)
	h2RE        = regexp.MustCompile(`(?s)<h2[^>]*>.*?</h2>`)
)

func layoutDocumentFragment(fragment string) string {
	fragment = strings.TrimSpace(fragment)
	if fragment == "" {
		return ""
	}
	fragment = statFenceRE.ReplaceAllStringFunc(fragment, renderStatFence)
	fragment = quoteRE.ReplaceAllString(fragment, `<aside class="callout">$1</aside>`)
	return wrapDocCards(fragment)
}

func renderStatFence(block string) string {
	m := statFenceRE.FindStringSubmatch(block)
	if len(m) < 2 {
		return block
	}
	type row struct {
		label string
		value float64
	}
	var rows []row
	var max float64
	for _, line := range strings.Split(html.UnescapeString(m[1]), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		label := strings.TrimSpace(parts[0])
		raw := strings.TrimSpace(parts[1])
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil || label == "" || n < 0 {
			continue
		}
		if n > max {
			max = n
		}
		rows = append(rows, row{label: label, value: n})
	}
	if len(rows) == 0 || max <= 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="stat-row">`)
	for _, row := range rows {
		b.WriteString(`<div class="stat"><b>`)
		b.WriteString(html.EscapeString(trimFloat(row.value)))
		b.WriteString(`</b><span>`)
		b.WriteString(html.EscapeString(row.label))
		b.WriteString(`</span></div>`)
	}
	b.WriteString(`</div><div class="chart"><svg viewBox="0 0 320 `)
	b.WriteString(strconv.Itoa(len(rows)*28 + 8))
	b.WriteString(`" role="img">`)
	for i, row := range rows {
		w := int(row.value / max * 200)
		if w < 1 {
			w = 1
		}
		y := 8 + i*28
		fmt.Fprintf(&b, `<rect x="108" y="%d" height="16" width="%d" rx="4" fill="#38bdf8"/>`, y, w)
		fmt.Fprintf(&b, `<text x="0" y="%d" fill="#94a3b8" font-size="12" font-family="system-ui,sans-serif">%s</text>`, y+13, html.EscapeString(row.label))
	}
	b.WriteString(`</svg></div>`)
	return b.String()
}

func trimFloat(n float64) string {
	if n == float64(int(n)) {
		return strconv.Itoa(int(n))
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}

func wrapDocCards(fragment string) string {
	idxs := h2RE.FindAllStringIndex(fragment, -1)
	if len(idxs) == 0 {
		return fragment
	}
	var b strings.Builder
	b.WriteString(fragment[:idxs[0][0]])
	for i, span := range idxs {
		end := len(fragment)
		if i+1 < len(idxs) {
			end = idxs[i+1][0]
		}
		card := fragment[span[0]:end]
		card = strings.ReplaceAll(card, "<ul>", `<ul class="rail">`)
		card = strings.ReplaceAll(card, "<ol>", `<ol class="rail">`)
		b.WriteString(`<section class="doc-card">`)
		b.WriteString(card)
		b.WriteString(`</section>`)
	}
	return b.String()
}
