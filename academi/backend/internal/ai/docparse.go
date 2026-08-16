package ai

import "strings"

const (
	academiDocStart = "---ACADEMI_DOC---"
	academiDocEnd   = "---END_ACADEMI_DOC---"
)

// parseAcademiDoc extracts a titled document from model output delimited in the system prompt.
// Returns display text for the chat (markers removed), title and body for persistence.
func parseAcademiDoc(reply string) (display, title, body string, ok bool) {
	start := strings.Index(reply, academiDocStart)
	end := strings.Index(reply, academiDocEnd)
	if start < 0 || end < 0 || end <= start {
		return reply, "", "", false
	}
	inner := strings.TrimSpace(reply[start+len(academiDocStart):end])
	title = ""
	body = inner
	lines := strings.Split(inner, "\n")
	if len(lines) == 0 {
		return reply, "", "", false
	}
	first := strings.TrimSpace(lines[0])
	if len(first) > 6 && strings.EqualFold(first[:6], "title:") {
		title = strings.TrimSpace(first[6:])
		body = strings.TrimSpace(strings.Join(lines[1:], "\n"))
	} else {
		body = inner
	}
	// Optional separator line "---" after title block
	body = strings.TrimSpace(body)
	if strings.HasPrefix(body, "---") {
		body = strings.TrimSpace(strings.TrimPrefix(body, "---"))
	}
	if strings.TrimSpace(body) == "" {
		return reply, "", "", false
	}
	if strings.TrimSpace(title) == "" {
		title = "Untitled document"
	}

	before := strings.TrimSpace(reply[:start])
	after := strings.TrimSpace(reply[end+len(academiDocEnd):])
	display = before
	if display != "" {
		display += "\n\n"
	}
	display += strings.TrimSpace(body)
	if after != "" {
		display += "\n\n" + after
	}
	display = strings.TrimSpace(display)

	return display, strings.TrimSpace(title), strings.TrimSpace(body), true
}
