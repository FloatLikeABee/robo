package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"idongivaflyinfa/internal/htmldoc"
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
	css := htmldoc.DarkCaseTaskDocumentCSS(caseTaskDetailUICSS())
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
