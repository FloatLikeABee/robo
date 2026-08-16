package surveybot

import "strings"

// UIBlock is an MCP-app style interactive widget for platform-chat.
type UIBlock struct {
	Type     string          `json:"type"` // mcp_app
	Widget   string          `json:"widget"`
	ID       string          `json:"id"`
	Label    string          `json:"label"`
	Options  []UIBlockOption `json:"options,omitempty"`
	SubmitAs map[string]string `json:"submit_as,omitempty"`
}

// UIBlockOption is one choice in a select/multiselect/confirm widget.
type UIBlockOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// BlockForStep builds a ui_block for an mcp_html survey step.
func BlockForStep(step Step) *UIBlock {
	if step.Collect != "mcp_html" {
		return nil
	}
	widget := step.Widget
	if widget == "" {
		widget = "select"
	}
	opts := make([]UIBlockOption, 0, len(step.Options))
	for _, o := range step.Options {
		val := slugifyOption(o)
		opts = append(opts, UIBlockOption{Value: val, Label: o})
	}
	if widget == "confirm" && len(opts) == 0 {
		opts = []UIBlockOption{
			{Value: "yes", Label: "Yes"},
			{Value: "no", Label: "No"},
		}
	}
	return &UIBlock{
		Type:     "mcp_app",
		Widget:   widget,
		ID:       step.Field,
		Label:    step.Prompt,
		Options:  opts,
		SubmitAs: map[string]string{"field": step.Field},
	}
}

// ConfirmBlock builds a yes/no confirm widget.
func ConfirmBlock(id, label string) UIBlock {
	return UIBlock{
		Type:   "mcp_app",
		Widget: "confirm",
		ID:     id,
		Label:  label,
		Options: []UIBlockOption{
			{Value: "yes", Label: "Yes"},
			{Value: "no", Label: "No"},
		},
		SubmitAs: map[string]string{"field": id},
	}
}

func slugifyOption(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "_")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "option"
	}
	return out
}

// ParseSurveyAnswerMessage extracts field=value from structured user replies.
// Accepts:
//   survey_bot_answer:gender=female
//   survey_bot_answer: field=gender value=female
func ParseSurveyAnswerMessage(msg string) (field, value string, ok bool) {
	msg = strings.TrimSpace(msg)
	low := strings.ToLower(msg)
	if !strings.HasPrefix(low, "survey_bot_answer:") {
		return "", "", false
	}
	rest := strings.TrimSpace(msg[len("survey_bot_answer:"):])
	if strings.Contains(rest, "=") && !strings.Contains(strings.ToLower(rest), "field=") {
		parts := strings.SplitN(rest, "=", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
	}
	// field=x value=y
	var f, v string
	for _, tok := range strings.Fields(rest) {
		if kv := strings.SplitN(tok, "=", 2); len(kv) == 2 {
			switch strings.ToLower(kv[0]) {
			case "field":
				f = kv[1]
			case "value":
				v = kv[1]
			}
		}
	}
	if f != "" {
		return f, v, true
	}
	return "", "", false
}
