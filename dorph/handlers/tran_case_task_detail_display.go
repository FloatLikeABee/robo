package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const caseTaskDetailMaxNesting = 5

var caseTaskCamelSplit = regexp.MustCompile(`([a-z0-9])([A-Z])`)

type caseTaskDetailKind int

const (
	caseTaskDetailEmpty caseTaskDetailKind = iota
	caseTaskDetailInvalid
	caseTaskDetailValue
)

type caseTaskDetailParse struct {
	kind  caseTaskDetailKind
	value any
	raw   string
}

func parseCaseTaskDetailJSON(raw string) caseTaskDetailParse {
	s := strings.TrimSpace(raw)
	if s == "" || s == "{}" || s == "[]" {
		return caseTaskDetailParse{kind: caseTaskDetailEmpty, raw: s}
	}
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return caseTaskDetailParse{kind: caseTaskDetailInvalid, raw: s}
	}
	if caseTaskDetailIsEmptyValue(v) {
		return caseTaskDetailParse{kind: caseTaskDetailEmpty, raw: s, value: v}
	}
	return caseTaskDetailParse{kind: caseTaskDetailValue, raw: s, value: v}
}

func caseTaskDetailIsEmptyValue(v any) bool {
	switch t := v.(type) {
	case nil:
		return true
	case map[string]any:
		return len(t) == 0
	case []any:
		return len(t) == 0
	default:
		return false
	}
}

func prettyCaseTaskDetailLabel(key string) string {
	s := strings.TrimSpace(key)
	s = strings.ReplaceAll(s, "_", " ")
	s = caseTaskCamelSplit.ReplaceAllString(s, "$1 $2")
	parts := strings.Fields(s)
	for i, w := range parts {
		if strings.EqualFold(w, "id") {
			parts[i] = "ID"
			continue
		}
		runes := []rune(strings.ToLower(w))
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}

func formatCaseTaskDetailScalar(v any) string {
	switch t := v.(type) {
	case nil:
		return "—"
	case bool:
		if t {
			return "Yes"
		}
		return "No"
	case float64:
		if !math.IsInf(t, 0) && !math.IsNaN(t) && t == math.Trunc(t) && math.Abs(t) < 1e15 {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	case string:
		return t
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprint(t)
		}
		return string(b)
	}
}

func sortedCaseTaskDetailKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func asCaseTaskDetailMap(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	return m, ok
}

func asCaseTaskDetailSlice(v any) ([]any, bool) {
	s, ok := v.([]any)
	return s, ok
}

func caseTaskDetailHeading(depth int, label string) string {
	n := depth + 3
	if n < 3 {
		n = 3
	}
	if n > 6 {
		n = 6
	}
	return strings.Repeat("#", n) + " " + label
}

func writeCaseTaskDetailMarkdownValue(b *strings.Builder, v any, depth int) {
	if depth > caseTaskDetailMaxNesting {
		b.WriteString(formatCaseTaskDetailScalar(v))
		b.WriteString("\n\n")
		return
	}
	if m, ok := asCaseTaskDetailMap(v); ok {
		if len(m) == 0 {
			b.WriteString("Empty\n\n")
			return
		}
		for _, k := range sortedCaseTaskDetailKeys(m) {
			label := prettyCaseTaskDetailLabel(k)
			child := m[k]
			if _, isObj := asCaseTaskDetailMap(child); isObj {
				b.WriteString(caseTaskDetailHeading(depth, label))
				b.WriteString("\n\n")
				writeCaseTaskDetailMarkdownValue(b, child, depth+1)
				continue
			}
			if arr, isArr := asCaseTaskDetailSlice(child); isArr {
				b.WriteString(caseTaskDetailHeading(depth, label))
				b.WriteString("\n\n")
				writeCaseTaskDetailMarkdownList(b, arr, depth+1)
				continue
			}
			b.WriteString("**")
			b.WriteString(label)
			b.WriteString(":** ")
			b.WriteString(formatCaseTaskDetailScalar(child))
			b.WriteString("\n\n")
		}
		return
	}
	if arr, ok := asCaseTaskDetailSlice(v); ok {
		writeCaseTaskDetailMarkdownList(b, arr, depth)
		return
	}
	b.WriteString(formatCaseTaskDetailScalar(v))
	b.WriteString("\n\n")
}

func writeCaseTaskDetailMarkdownList(b *strings.Builder, arr []any, depth int) {
	if len(arr) == 0 {
		b.WriteString("Empty list\n\n")
		return
	}
	for _, item := range arr {
		if _, isObj := asCaseTaskDetailMap(item); isObj {
			b.WriteString("- ")
			var inner strings.Builder
			writeCaseTaskDetailMarkdownValue(&inner, item, depth+1)
			line := strings.TrimSpace(inner.String())
			line = strings.ReplaceAll(line, "\n\n", "; ")
			line = strings.ReplaceAll(line, "\n", " ")
			b.WriteString(line)
			b.WriteString("\n")
			continue
		}
		if nested, isArr := asCaseTaskDetailSlice(item); isArr {
			b.WriteString("- \n")
			writeCaseTaskDetailMarkdownList(b, nested, depth+1)
			continue
		}
		b.WriteString("- ")
		b.WriteString(formatCaseTaskDetailScalar(item))
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func detailJSONToMarkdown(raw string) string {
	parsed := parseCaseTaskDetailJSON(raw)
	switch parsed.kind {
	case caseTaskDetailEmpty:
		return ""
	case caseTaskDetailInvalid:
		return "## Detail\n\n" + parsed.raw + "\n"
	}
	var b strings.Builder
	b.WriteString("## Detail\n\n")
	writeCaseTaskDetailMarkdownValue(&b, parsed.value, 0)
	return b.String()
}

func writeCaseTaskDetailHTMLValue(b *strings.Builder, v any, depth int) {
	if depth > caseTaskDetailMaxNesting {
		b.WriteString(`<p class="detail-value">`)
		b.WriteString(htmlEscapeMinimal(formatCaseTaskDetailScalar(v)))
		b.WriteString(`</p>`)
		return
	}
	if m, ok := asCaseTaskDetailMap(v); ok {
		if len(m) == 0 {
			b.WriteString(`<p class="detail-muted">Empty</p>`)
			return
		}
		if depth == 0 {
			for _, k := range sortedCaseTaskDetailKeys(m) {
				label := prettyCaseTaskDetailLabel(k)
				b.WriteString(`<article class="detail-card"><h3>`)
				b.WriteString(htmlEscapeMinimal(label))
				b.WriteString(`</h3>`)
				writeCaseTaskDetailHTMLValue(b, m[k], depth+1)
				b.WriteString(`</article>`)
			}
			return
		}
		b.WriteString(`<div class="detail-nested">`)
		for _, k := range sortedCaseTaskDetailKeys(m) {
			label := prettyCaseTaskDetailLabel(k)
			b.WriteString(`<div class="detail-field"><div class="detail-label">`)
			b.WriteString(htmlEscapeMinimal(label))
			b.WriteString(`</div>`)
			writeCaseTaskDetailHTMLValue(b, m[k], depth+1)
			b.WriteString(`</div>`)
		}
		b.WriteString(`</div>`)
		return
	}
	if arr, ok := asCaseTaskDetailSlice(v); ok {
		if len(arr) == 0 {
			b.WriteString(`<p class="detail-muted">Empty list</p>`)
			return
		}
		b.WriteString(`<ul class="detail-list">`)
		for i, item := range arr {
			b.WriteString(`<li>`)
			if _, isObj := asCaseTaskDetailMap(item); isObj {
				b.WriteString(`<div class="detail-item-label">Item `)
				b.WriteString(strconv.Itoa(i + 1))
				b.WriteString(`</div>`)
			}
			writeCaseTaskDetailHTMLValue(b, item, depth+1)
			b.WriteString(`</li>`)
		}
		b.WriteString(`</ul>`)
		return
	}
	b.WriteString(`<p class="detail-value">`)
	b.WriteString(htmlEscapeMinimal(formatCaseTaskDetailScalar(v)))
	b.WriteString(`</p>`)
}

func detailJSONToHTMLFragment(raw string) string {
	parsed := parseCaseTaskDetailJSON(raw)
	switch parsed.kind {
	case caseTaskDetailEmpty:
		return ""
	case caseTaskDetailInvalid:
		return `<section class="detail-ui"><h2>Detail</h2><pre class="detail-fallback">` + htmlEscapeMinimal(parsed.raw) + `</pre></section>`
	}
	var b strings.Builder
	b.WriteString(`<section class="detail-ui"><h2>Detail</h2>`)
	writeCaseTaskDetailHTMLValue(&b, parsed.value, 0)
	b.WriteString(`</section>`)
	return b.String()
}

func caseTaskDetailUICSS() string {
	return `.detail-ui{margin-top:1.5rem}
.detail-ui h2{font:700 1.3rem/1.25 system-ui,sans-serif;color:#f8fafc;margin:0 0 .85rem}
.detail-card{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:1rem 1.1rem;margin:0 0 .85rem}
.detail-card h3{font:700 .95rem/1.3 system-ui,sans-serif;color:var(--accent);margin:0 0 .5rem;letter-spacing:.02em}
.detail-label{font:700 11px/1.3 system-ui,sans-serif;color:var(--muted);margin:0 0 .2rem}
.detail-value{font:15px/1.45 system-ui,sans-serif;color:var(--ink);margin:0;word-break:break-word;white-space:pre-wrap}
.detail-muted{font:13px/1.4 system-ui,sans-serif;color:var(--muted);margin:0}
.detail-nested{border-left:2px solid var(--line);padding-left:.85rem;margin:.35rem 0}
.detail-field{margin:.55rem 0}
.detail-list{margin:.35rem 0 .2rem;padding-left:1.2em}
.detail-item-label{font:700 11px/1.3 system-ui,sans-serif;color:var(--accent);margin:.15rem 0 .25rem}
.detail-fallback{overflow:auto;padding:1em;border-radius:10px;background:var(--card);border:1px solid var(--line);font:13px/1.45 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;white-space:pre-wrap;word-break:break-word;color:var(--ink)}`
}
