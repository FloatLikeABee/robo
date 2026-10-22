package mcp

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed agent_guide.md
var AgentGuideMarkdown string

// AgentRuntimeToolCount is suggest_note_kind and dorph_guide.
const AgentRuntimeToolCount = 2

const guideURI = "dorph://guide"
const guideMIME = "text/markdown"

const (
	kindNote     = "note"
	kindTodo     = "todo"
	kindTask     = "task"
	kindBigNote  = "big_note"
	kindTimeline = "timeline"
	kindResearch = "research"
	kindStick    = "stick_note"
	kindEventLog = "event_log"
)

// KindSuggestion is the structured result of suggest_note_kind.
type KindSuggestion struct {
	Kind       string  `json:"kind"`
	Reason     string  `json:"reason"`
	UseGrill   bool    `json:"use_grill"`
	Confidence float64 `json:"confidence"`
}

type suggestKindInput struct {
	Intent      string `json:"intent" jsonschema:"The user's ask or title."`
	ContentHint string `json:"content_hint,omitempty" jsonschema:"Optional extra text the agent already has."`
}

type guideInput struct {
	Topic string `json:"topic,omitempty" jsonschema:"kinds, sharing, grill, limits, setup, checking, pool, errors, tools, writing, or privacy. Omit for the outline and golden rules."`
}

type guideOutput struct {
	Topic    string `json:"topic"`
	Markdown string `json:"markdown"`
}

var (
	clockRe = regexp.MustCompile(`(?i)\b\d{1,2}(?::\d{2})?\s*(?:am|pm)\b|\b\d{1,2}:\d{2}\b`)
	urlRe   = regexp.MustCompile(`(?i)https?://|www\.`)
	todoRe  = regexp.MustCompile(`(?i)^\s*todo\b`)
	wordRe  = map[string]*regexp.Regexp{
		"buy":    regexp.MustCompile(`(?i)\bbuy\b`),
		"call":   regexp.MustCompile(`(?i)\bcall\b`),
		"send":   regexp.MustCompile(`(?i)\bsend\b`),
		"submit": regexp.MustCompile(`(?i)\bsubmit\b`),
		"fix":    regexp.MustCompile(`(?i)\bfix\b`),
	}
)

// GuideTopics are the dorph_guide topic keys.
var GuideTopics = []string{
	"kinds", "writing", "grill", "checking", "sharing", "pool", "errors", "limits", "setup", "privacy", "tools",
}

var guideTopicHeading = map[string]string{
	"kinds":    "4. Choosing a kind",
	"writing":  "5. Writing good notes",
	"grill":    "6. The grill",
	"checking": "7. Checking notes",
	"sharing":  "8. Sharing",
	"pool":     "9. Using the pool",
	"errors":   "10. Errors and limits",
	"limits":   "10. Errors and limits",
	"setup":    "11. Per-client setup",
	"privacy":  "12. Privacy and what Dorph stores",
	"tools":    "13. Tool reference",
}

func registerAgentRuntime(server *sdkmcp.Server, recheck func() error) {
	sdkmcp.AddTool(server, readOnlyTool(
		"suggest_note_kind",
		"Suggest note kind",
		"Pick a Dorph note kind from the user's ask. Rules only; does not call a model. Returns kind, reason, use_grill, and confidence.",
	), suggestNoteKindTool(recheck))
	sdkmcp.AddTool(server, readOnlyTool(
		"dorph_guide",
		"Dorph guide",
		"Return a markdown excerpt of the Dorph agent playbook. Omit topic for the outline and golden rules. Topics: kinds, sharing, grill, limits, setup, checking, pool, errors, tools, writing, privacy.",
	), dorphGuideTool(recheck))
	server.AddResource(guideResource(), readGuide)
}

func readOnlyTool(name, title, description string) *sdkmcp.Tool {
	closedWorld := false
	return &sdkmcp.Tool{
		Name:        name,
		Description: description,
		Annotations: &sdkmcp.ToolAnnotations{
			Title:         title,
			ReadOnlyHint:  true,
			OpenWorldHint: &closedWorld,
		},
	}
}

func guideResource() *sdkmcp.Resource {
	return &sdkmcp.Resource{
		URI:         guideURI,
		Name:        "dorph-guide",
		Title:       "Dorph for agents",
		Description: "Playbook for any agent using Dorph notes.",
		MIMEType:    guideMIME,
	}
}

func readGuide(context.Context, *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
	return &sdkmcp.ReadResourceResult{
		Contents: []*sdkmcp.ResourceContents{{
			URI:      guideURI,
			MIMEType: guideMIME,
			Text:     AgentGuideMarkdown,
		}},
	}, nil
}

func suggestNoteKindTool(recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, suggestKindInput) (*sdkmcp.CallToolResult, KindSuggestion, error) {
	return func(_ context.Context, _ *sdkmcp.CallToolRequest, in suggestKindInput) (*sdkmcp.CallToolResult, KindSuggestion, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, KindSuggestion{}, err
		}
		out, err := SuggestNoteKind(in.Intent, in.ContentHint)
		if err != nil {
			return nil, KindSuggestion{}, err
		}
		return nil, out, nil
	}
}

func dorphGuideTool(recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, guideInput) (*sdkmcp.CallToolResult, guideOutput, error) {
	return func(_ context.Context, _ *sdkmcp.CallToolRequest, in guideInput) (*sdkmcp.CallToolResult, guideOutput, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, guideOutput{}, err
		}
		md, err := GuideExcerpt(in.Topic)
		if err != nil {
			return nil, guideOutput{}, err
		}
		topic := strings.TrimSpace(in.Topic)
		if topic == "" {
			topic = "outline"
		}
		return nil, guideOutput{Topic: topic, Markdown: md}, nil
	}
}

// SuggestNoteKind maps an ask to a kind using ordered rules. It does not call a model.
func SuggestNoteKind(intent, contentHint string) (KindSuggestion, error) {
	intent = strings.TrimSpace(intent)
	if intent == "" {
		return KindSuggestion{}, fmt.Errorf("invalid_input: intent is required")
	}
	blob := strings.ToLower(strings.TrimSpace(intent + " " + contentHint))
	kind, reason, conf := matchKind(blob, contentHint)
	return KindSuggestion{
		Kind:       kind,
		Reason:     reason,
		UseGrill:   shouldGrill(intent, contentHint, blob),
		Confidence: conf,
	}, nil
}

func matchKind(blob, contentHint string) (kind, reason string, conf float64) {
	switch {
	case isStick(blob):
		return kindStick, "Timed reminder or calendar stick.", 0.9
	case isTask(blob):
		return kindTask, "Board, assignee, or staged work.", 0.9
	case isTimeline(blob):
		return kindTimeline, "Dated steps, itinerary, or chronology.", 0.9
	case isEventLog(blob):
		return kindEventLog, "One happening to log.", 0.9
	case isTodo(blob):
		return kindTodo, "A single actionable item.", 0.9
	case isResearch(blob):
		return kindResearch, "Inquiry with sources or findings.", 0.9
	case isBigNote(blob, contentHint):
		return kindBigNote, "Long-form or chaptered document.", 0.9
	default:
		if hasSpecifics(blob, contentHint) {
			return kindNote, "A fact, reference, or short summary.", 0.75
		}
		return kindNote, "Default kind when the ask is not a task, todo, or dated record.", 0.55
	}
}

func isStick(blob string) bool {
	if strings.Contains(blob, "calendar reminder") || strings.Contains(blob, "stick a reminder") || strings.Contains(blob, "stick reminder") {
		return true
	}
	if strings.Contains(blob, "don't forget at") || strings.Contains(blob, "do not forget at") {
		return true
	}
	return strings.Contains(blob, "reminder") && clockRe.MatchString(blob)
}

func isTask(blob string) bool {
	return strings.Contains(blob, "assign") ||
		strings.Contains(blob, "assignee") ||
		strings.Contains(blob, "kanban") ||
		strings.Contains(blob, "dorphnotes board") ||
		strings.Contains(blob, " board") ||
		strings.HasPrefix(blob, "board") ||
		strings.Contains(blob, "in review") ||
		strings.Contains(blob, "in doing") ||
		strings.Contains(blob, "put it in") ||
		strings.Contains(blob, "column")
}

func isTimeline(blob string) bool {
	return strings.Contains(blob, "itinerary") ||
		strings.Contains(blob, "chronology") ||
		strings.Contains(blob, "day 1") ||
		strings.Contains(blob, "weekly plan") ||
		(strings.Contains(blob, "history") && strings.Contains(blob, "date")) ||
		strings.Contains(blob, "roadmap")
}

func isEventLog(blob string) bool {
	return strings.Contains(blob, "log the incident") ||
		strings.Contains(blob, "decision log") ||
		strings.Contains(blob, "incident:") ||
		strings.Contains(blob, "event log") ||
		strings.Contains(blob, "log the ")
}

func isTodo(blob string) bool {
	if todoRe.MatchString(blob) || strings.TrimSpace(blob) == "todo" {
		return true
	}
	if strings.Contains(blob, "need to") || strings.Contains(blob, "pick up") {
		return true
	}
	for _, re := range wordRe {
		if re.MatchString(blob) {
			return true
		}
	}
	return false
}

func isResearch(blob string) bool {
	if strings.Contains(blob, "research") || strings.Contains(blob, "cite source") || strings.Contains(blob, "cite sources") {
		return true
	}
	if strings.Contains(blob, "investigate") && strings.Contains(blob, "finding") {
		return true
	}
	return strings.Contains(blob, "compare") && strings.Contains(blob, "source")
}

func isBigNote(blob, contentHint string) bool {
	if wordCount(contentHint) > 2000 {
		return true
	}
	return strings.Contains(blob, "chapter") ||
		strings.Contains(blob, "long-form") ||
		strings.Contains(blob, "long form") ||
		strings.Contains(blob, "2000 word")
}

func shouldGrill(intent, contentHint, blob string) bool {
	if wordCount(contentHint) > 40 {
		return false
	}
	if strings.Contains(blob, "help me write") || strings.Contains(blob, "help me draft") {
		return true
	}
	if strings.Contains(blob, "share") && strings.Contains(blob, "pool") {
		return true
	}
	// ponytail: two-word ceiling for "vague ask"; longer asks are treated as specified enough to skip the grill. Upgrade: fold in a specificity scorer if the table grows past short labels.
	return wordCount(intent) <= 2
}

func hasSpecifics(blob, contentHint string) bool {
	if urlRe.MatchString(blob) || clockRe.MatchString(blob) {
		return true
	}
	if wordCount(contentHint) >= 8 {
		return true
	}
	if strings.Contains(blob, "summary") || strings.Contains(blob, "handoff") || strings.Contains(blob, " is ") {
		return true
	}
	return wordCount(blob) >= 6
}

func wordCount(s string) int {
	n := 0
	inWord := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			inWord = false
			continue
		}
		if !inWord {
			inWord = true
			n++
		}
	}
	return n
}

// GuideExcerpt returns a markdown slice of the playbook. Empty topic is outline plus golden rules.
func GuideExcerpt(topic string) (string, error) {
	topic = strings.ToLower(strings.TrimSpace(topic))
	if topic == "" {
		return guideOutline() + "\n\n" + sectionByHeading("3. Golden rules"), nil
	}
	heading, ok := guideTopicHeading[topic]
	if !ok {
		return "", fmt.Errorf("invalid_input: unknown topic; try %s", strings.Join(GuideTopics, ", "))
	}
	sec := sectionByHeading(heading)
	if strings.TrimSpace(sec) == "" {
		return "", fmt.Errorf("invalid_input: topic %s has no section", topic)
	}
	return sec, nil
}

func guideOutline() string {
	var b strings.Builder
	b.WriteString("# Dorph for agents\n\nRead the full playbook at dorph://guide.\n\n")
	b.WriteString("HTML: Dorph's base skill is html-anything at skills/vendor/html-anything/. Load it when writing HTML.\n\n")
	for _, line := range strings.Split(AgentGuideMarkdown, "\n") {
		if strings.HasPrefix(line, "## ") {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func sectionByHeading(heading string) string {
	needle := "## " + heading
	idx := strings.Index(AgentGuideMarkdown, needle)
	if idx < 0 {
		return ""
	}
	rest := AgentGuideMarkdown[idx:]
	next := strings.Index(rest[len(needle):], "\n## ")
	if next < 0 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:len(needle)+next])
}
