package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type caseTaskViewDoc struct {
	Title       string
	Description string
	Assignees   string
	StartAt     *time.Time
	EndAt       *time.Time
	Location    string
	DetailJSON  string
}

func formatCaseTaskViewTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04")
}

func caseTaskLocationSummary(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		return s
	}
	label, _ := parsed["label"].(string)
	label = strings.TrimSpace(label)
	if label == "" {
		if loc, ok := parsed["location"].(string); ok {
			label = strings.TrimSpace(loc)
		}
	}
	areaLen := 0
	if arr, ok := parsed["area"].([]interface{}); ok {
		areaLen = len(arr)
	}
	switch {
	case label != "" && areaLen > 0:
		return fmt.Sprintf("%s (%d points)", label, areaLen)
	case label != "":
		return label
	case areaLen > 0:
		return fmt.Sprintf("Area (%d points)", areaLen)
	default:
		return ""
	}
}

func buildCaseTaskMarkdownCore(in caseTaskViewDoc) string {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = "Case/task"
	}
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(title)
	b.WriteString("\n\n")
	if a := strings.TrimSpace(in.Assignees); a != "" {
		b.WriteString("**Assignees:** ")
		b.WriteString(a)
		b.WriteString("\n\n")
	}
	if t := formatCaseTaskViewTime(in.StartAt); t != "" {
		b.WriteString("**Start:** ")
		b.WriteString(t)
		b.WriteString("\n\n")
	}
	if t := formatCaseTaskViewTime(in.EndAt); t != "" {
		b.WriteString("**End:** ")
		b.WriteString(t)
		b.WriteString("\n\n")
	}
	if loc := caseTaskLocationSummary(in.Location); loc != "" {
		b.WriteString("**Location:** ")
		b.WriteString(loc)
		b.WriteString("\n\n")
	}
	if d := strings.TrimSpace(in.Description); d != "" {
		b.WriteString(d)
		b.WriteString("\n\n")
	}
	return b.String()
}

func buildCaseTaskMarkdown(in caseTaskViewDoc) string {
	core := buildCaseTaskMarkdownCore(in)
	detail := detailJSONToMarkdown(in.DetailJSON)
	if detail == "" {
		return core
	}
	if strings.TrimSpace(core) == "" {
		return detail
	}
	return strings.TrimRight(core, "\n") + "\n\n" + detail
}

func buildCaseTaskHTML(in caseTaskViewDoc) string {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = "Case/task"
	}
	markdown := buildCaseTaskMarkdownCore(in)
	proseCSS := `.prose{font-size:1.05rem;color:var(--ink)}
.prose>:first-child{margin-top:0}
.prose h1,.prose h2,.prose h3,.prose h4{line-height:1.25;margin:1.35em 0 .55em;font-weight:700;color:#f8fafc}
.prose h1{font-size:1.55rem}.prose h2{font-size:1.3rem}.prose h3{font-size:1.12rem}
.prose p,.prose ul,.prose ol,.prose blockquote,.prose pre,.prose table{margin:.85em 0}
.prose ul,.prose ol{padding-left:1.4em}
.prose li{margin:.35em 0}
.prose blockquote{padding:.35em 0 .35em 1em;border-left:3px solid var(--accent);color:var(--muted)}
.prose code{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:.9em;background:var(--card);padding:.12em .35em;border-radius:4px;border:1px solid var(--line)}
.prose pre{overflow:auto;padding:1em;border-radius:10px;background:var(--card);border:1px solid var(--line)}
.prose a{color:var(--accent)}
.prose hr{border:0;border-top:1px solid var(--line);margin:1.5em 0}
.prose strong{font-weight:700}
.prose table{border-collapse:collapse;width:100%;font-size:.95em}
.prose th,.prose td{border:1px solid var(--line);padding:.45em .6em;text-align:left}
.prose th{background:var(--card)}`
	css := `:root{color-scheme:dark;--ink:#e8eef7;--muted:#94a3b8;--line:#1e293b;--bg:#0b1220;--card:#111827;--accent:#38bdf8;--scrollbar-thumb:#64748b;--scrollbar-track:#0b1220;scrollbar-width:thin;scrollbar-color:var(--scrollbar-thumb) var(--scrollbar-track)}
html,body{height:100%;margin:0}
*{box-sizing:border-box}body{font-family:Georgia,"Times New Roman",serif;color:var(--ink);background:radial-gradient(1200px 600px at 10% -10%,#1e293b 0%,var(--bg) 55%);line-height:1.55}
.wrap{max-width:760px;margin:0 auto;padding:2rem 1.25rem 3rem}
.meta{font:12px/1.4 system-ui,sans-serif;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin-bottom:.75rem}
h1.page-title{font-size:clamp(1.6rem,3vw,2.2rem);margin:0 0 1rem;line-height:1.2;color:#f8fafc}
` + proseCSS + `
` + caseTaskDetailUICSS() + `
.foot{font:11px/1.4 system-ui,sans-serif;color:var(--muted);margin-top:2rem}`
	mdBody := stripLeadingTitleHeading(markdown, title)
	rendered := markdownToHTMLFragment(mdBody)
	var body strings.Builder
	body.WriteString(`<!doctype html><html lang="en"><head><meta charset="utf-8"/>`)
	body.WriteString(`<meta name="color-scheme" content="dark"/>`)
	body.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1"/>`)
	body.WriteString(`<title>` + htmlEscapeMinimal(title) + `</title>`)
	body.WriteString(`<style>` + css + `</style></head><body><div class="wrap">`)
	body.WriteString(`<div class="meta">Case/task</div>`)
	body.WriteString(`<h1 class="page-title">` + htmlEscapeMinimal(title) + `</h1>`)
	body.WriteString(`<div class="prose">` + rendered + `</div>`)
	body.WriteString(detailJSONToHTMLFragment(in.DetailJSON))
	body.WriteString(`<p class="foot">Generated Case/task</p></div></body></html>`)
	return body.String()
}

func parseCaseTaskViewTime(v interface{}) *time.Time {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case time.Time:
		tt := t
		return &tt
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil
		}
		layouts := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02T15:04",
			"2006-01-02 15:04",
			"2006-01-02",
		}
		for _, layout := range layouts {
			if parsed, err := time.Parse(layout, s); err == nil {
				return &parsed
			}
		}
	}
	return nil
}

func mapStringVal(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func caseTaskViewDocFromFullRow(m map[string]interface{}) caseTaskViewDoc {
	return caseTaskViewDoc{
		Title:       mapStringVal(m, "title"),
		Description: mapStringVal(m, "description"),
		Assignees:   mapStringVal(m, "assignees_label"),
		StartAt:     parseCaseTaskViewTime(m["start_at"]),
		EndAt:       parseCaseTaskViewTime(m["end_at"]),
		Location:    mapStringVal(m, "location"),
		DetailJSON:  existingDetailFromRow(m),
	}
}

func attachCaseTaskViewDocuments(m map[string]interface{}) {
	if m == nil {
		return
	}
	doc := caseTaskViewDocFromFullRow(m)
	m["markdown_content"] = buildCaseTaskMarkdown(doc)
	m["html_content"] = buildCaseTaskHTML(doc)
}
