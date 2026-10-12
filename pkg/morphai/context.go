package morphai

import (
	"strings"
	"unicode/utf8"
)

// Default history limits for satellite AI assistants (smaller = faster).
const (
	DefaultHistoryMaxMessages = 8
	DefaultHistoryMaxRunes    = 800
	DefaultToolResultMaxRunes = 12_000
	DefaultToolMaxRounds      = 8
)

const FastToolFirstInstructions = `Use the fastest grounded source available. Prefer live MCP-style catalogs, product APIs, repositories, or read-only database/schema lookups before broad reasoning. List or search first, then fetch details by id. Use read-only SQL unless the user explicitly asks to create, update, or delete.`

// ReplyShapeInstructions is the short default for assistant replies.
// A diagram, chart, table, or code block is allowed only when it makes the answer clearer.
const ReplyShapeInstructions = `Answer the question in plain language. Be concise. Use a diagram, chart, table, or code block only when the question is about structure, quantities, or code and that form is clearer. Do not add a caption, mermaid diagram, or pixel grid for an ordinary question. Do not invent facts.`

// VisualFirstInstructions is the previous name of ReplyShapeInstructions.
// It no longer asks for a caption or a diagram.
const VisualFirstInstructions = ReplyShapeInstructions

// ModelIdentityInstructions tells the model the configured provider and model.
// Empty provider and model means the configuration is unknown.
func ModelIdentityInstructions(provider, model string) string {
	provider = trimSpace(provider)
	model = trimSpace(model)
	if provider == "" && model == "" {
		return "If asked which model you are, say you do not know. Do not invent a name, version, or date."
	}
	if provider == "" {
		provider = "unknown"
	}
	if model == "" {
		model = "unknown"
	}
	return "If asked which model you are, say provider " + provider + ", model " + model + ". Do not invent a version or date."
}

// TruncateRunes shortens s to at most max Unicode runes without splitting UTF-8.
func TruncateRunes(s string, max int) string {
	if max <= 0 {
		return s
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "…"
}

// TruncateHistoryContent trims a single chat turn for prompt injection.
func TruncateHistoryContent(content string, maxRunes int) string {
	content = trimSpace(content)
	if maxRunes <= 0 {
		return content
	}
	return TruncateRunes(content, maxRunes)
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// ToolFollowUpPrompt builds a compact tool-loop message (avoids echoing the prior JSON call).
func ToolFollowUpPrompt(toolResult string) string {
	return toolResult + "\n\n" + ReplyShapeInstructions + " Never call an image API from the tool loop. If you need another tool, reply with only one JSON object."
}

// ToolFollowUpPromptWithInstruction keeps product-specific answer requirements while
// preserving the compact tool-loop contract.
func ToolFollowUpPromptWithInstruction(toolResult, instruction string) string {
	instruction = trimSpace(instruction)
	if instruction == "" {
		return ToolFollowUpPrompt(toolResult)
	}
	return toolResult + "\n\n" + ReplyShapeInstructions + "\n" + instruction + " Never call an image API from the tool loop. If another tool is needed, reply with only one JSON object."
}

// ExtractJSONObject returns the first complete top-level JSON object from model output.
func ExtractJSONObject(s string) (string, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	if start < 0 {
		return "", false
	}
	depth := 0
	for i, ch := range s[start:] {
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : start+i+1], true
			}
		}
	}
	return "", false
}
