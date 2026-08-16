package surveybot

import (
	"fmt"
	"regexp"
	"strings"
)

// Step is one question in a Survey Bot markdown template.
type Step struct {
	Index    int      `json:"index"`
	Field    string   `json:"field"`
	Collect  string   `json:"collect"` // text | mcp_html
	Widget   string   `json:"widget,omitempty"` // select | multiselect | confirm
	Options  []string `json:"options,omitempty"`
	Required bool     `json:"required"`
	Prompt   string   `json:"prompt"`
	Heading  string   `json:"heading,omitempty"`
}

// ParsedTemplate is the structured form of a Survey Bot MD/TXT file.
type ParsedTemplate struct {
	Slug         string   `json:"slug,omitempty"`
	Title        string   `json:"title"`
	Tags         []string `json:"tags,omitempty"`
	Instructions string   `json:"instructions,omitempty"`
	Steps        []Step   `json:"steps"`
	RawMarkdown  string   `json:"-"`
	NeedsCompile bool     `json:"needs_compile,omitempty"`
}

var (
	frontmatterRe = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n(.*)`)
	qnHeadingRe   = regexp.MustCompile(`(?m)^##\s+(Q\d+)\s*[—\-–:]?\s*(.*)$`)
	bulletKVRe    = regexp.MustCompile(`(?i)^\s*-\s*([a-z_]+)\s*:\s*(.+?)\s*$`)
)

// ParseMarkdown turns a Survey Bot MD/TXT template into structured steps.
// Description-only files (no ## Qn sections) are valid and set NeedsCompile=true.
func ParseMarkdown(md string) (*ParsedTemplate, error) {
	md = strings.TrimSpace(md)
	if md == "" {
		return nil, fmt.Errorf("empty markdown template")
	}
	out := &ParsedTemplate{RawMarkdown: md}
	body := md
	if m := frontmatterRe.FindStringSubmatch(md); len(m) == 3 {
		parseFrontmatter(m[1], out)
		body = m[2]
	}

	locs := qnHeadingRe.FindAllStringSubmatchIndex(body, -1)
	if len(locs) == 0 {
		// Description-only sheet: instructions = body, compile later.
		pre := strings.TrimSpace(body)
		out.Instructions = stripHeading(pre)
		if out.Title == "" {
			out.Title = firstH1(pre)
		}
		if out.Title == "" {
			out.Title = firstNonEmptyLine(pre)
		}
		if out.Title == "" {
			out.Title = "Survey"
		}
		if out.Instructions == "" {
			out.Instructions = pre
		}
		if strings.TrimSpace(out.Instructions) == "" {
			return nil, fmt.Errorf("empty description — add instructions or ## Qn sections")
		}
		out.NeedsCompile = true
		return out, nil
	}

	pre := strings.TrimSpace(body[:locs[0][0]])
	if pre != "" {
		out.Instructions = stripHeading(pre)
	}
	if out.Title == "" {
		out.Title = firstH1(pre)
	}
	if out.Title == "" {
		out.Title = "Survey"
	}

	for i, loc := range locs {
		start := loc[0]
		end := len(body)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		block := body[start:end]
		step, err := parseStepBlock(i, block)
		if err != nil {
			return nil, fmt.Errorf("step %d: %w", i+1, err)
		}
		out.Steps = append(out.Steps, step)
	}
	if len(out.Steps) == 0 {
		return nil, fmt.Errorf("template has no usable steps")
	}
	return out, nil
}

// RequireSteps returns an error if the template still needs question compilation.
func RequireSteps(p *ParsedTemplate) error {
	if p == nil {
		return fmt.Errorf("nil template")
	}
	if p.NeedsCompile || len(p.Steps) == 0 {
		return fmt.Errorf("template has no questions yet — compile from the description first")
	}
	return nil
}

// FormatMarkdown rebuilds a canonical MD template from a parsed structure.
func FormatMarkdown(p *ParsedTemplate) string {
	if p == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("---\n")
	if p.Slug != "" {
		b.WriteString("slug: " + p.Slug + "\n")
	}
	b.WriteString("title: " + p.Title + "\n")
	if len(p.Tags) > 0 {
		b.WriteString("tags: [" + strings.Join(p.Tags, ", ") + "]\n")
	}
	b.WriteString("---\n\n")
	b.WriteString("# Instructions\n")
	if strings.TrimSpace(p.Instructions) != "" {
		b.WriteString(strings.TrimSpace(p.Instructions))
		b.WriteString("\n")
	} else {
		b.WriteString("Ask one question at a time.\n")
	}
	b.WriteByte('\n')
	for i, s := range p.Steps {
		heading := s.Heading
		if heading == "" {
			heading = s.Prompt
		}
		b.WriteString(fmt.Sprintf("## Q%d — %s\n", i+1, heading))
		b.WriteString(fmt.Sprintf("- field: %s\n", s.Field))
		b.WriteString(fmt.Sprintf("- collect: %s\n", s.Collect))
		if s.Collect == "mcp_html" {
			widget := s.Widget
			if widget == "" {
				widget = "select"
			}
			b.WriteString(fmt.Sprintf("- widget: %s\n", widget))
			if len(s.Options) > 0 {
				b.WriteString(fmt.Sprintf("- options: [%s]\n", strings.Join(s.Options, ", ")))
			}
		}
		req := "false"
		if s.Required {
			req = "true"
		}
		b.WriteString(fmt.Sprintf("- required: %s\n", req))
		b.WriteString(fmt.Sprintf("- prompt: %s\n\n", s.Prompt))
	}
	return strings.TrimSpace(b.String()) + "\n"
}

func parseFrontmatter(raw string, out *ParsedTemplate) {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)
		switch key {
		case "id", "slug":
			out.Slug = val
		case "title":
			out.Title = val
		case "tags":
			out.Tags = parseList(val)
		}
	}
}

func parseList(val string) []string {
	val = strings.TrimSpace(val)
	val = strings.TrimPrefix(val, "[")
	val = strings.TrimSuffix(val, "]")
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"'`)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseStepBlock(index int, block string) (Step, error) {
	lines := strings.Split(block, "\n")
	heading := ""
	if len(lines) > 0 {
		if m := qnHeadingRe.FindStringSubmatch(lines[0]); len(m) == 3 {
			heading = strings.TrimSpace(m[2])
		}
	}
	step := Step{
		Index:    index,
		Heading:  heading,
		Collect:  "text",
		Required: true,
	}
	for _, line := range lines[1:] {
		m := bulletKVRe.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(m[1]))
		val := strings.TrimSpace(m[2])
		val = strings.Trim(val, `"'`)
		switch key {
		case "field":
			step.Field = val
		case "collect":
			step.Collect = strings.ToLower(val)
		case "widget":
			step.Widget = strings.ToLower(val)
		case "options":
			step.Options = parseList(val)
		case "required":
			step.Required = strings.EqualFold(val, "true") || val == "1" || strings.EqualFold(val, "yes")
		case "prompt":
			step.Prompt = val
		}
	}
	if step.Field == "" {
		step.Field = fmt.Sprintf("q%d", index+1)
	}
	if step.Prompt == "" {
		if heading != "" {
			step.Prompt = heading + "?"
		} else {
			step.Prompt = "Please answer:"
		}
	}
	if step.Collect == "mcp_html" && step.Widget == "" {
		step.Widget = "select"
	}
	if step.Collect == "mcp_html" && (step.Widget == "select" || step.Widget == "multiselect") && len(step.Options) == 0 {
		return step, fmt.Errorf("field %s: mcp_html %s requires options", step.Field, step.Widget)
	}
	return step, nil
}

func firstH1(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}
		if len(line) > 80 {
			return line[:80]
		}
		return line
	}
	return ""
}

func stripHeading(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "# ") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

// SearchScore ranks a template against a free-text query (simple keyword overlap).
func SearchScore(title, summary string, tags []string, markdown, query string) int {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return 0
	}
	hay := strings.ToLower(title + " " + summary + " " + strings.Join(tags, " ") + " " + markdown)
	score := 0
	for _, tok := range strings.Fields(q) {
		tok = strings.Trim(tok, ".,!?")
		if len(tok) < 2 {
			continue
		}
		if strings.Contains(hay, tok) {
			score++
		}
	}
	return score
}
