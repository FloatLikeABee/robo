package morphai

import (
	"strings"
	"testing"
)

func TestToolFollowUpPromptIncludesVisualFirst(t *testing.T) {
	got := ToolFollowUpPrompt("TOOL_RESULT ok")
	for _, p := range []string{"TOOL_RESULT ok", "Visual-first", "mermaid", "```pixel```", "never call an image API"} {
		if !strings.Contains(got, p) {
			t.Fatalf("missing %q in:\n%s", p, got)
		}
	}
}

func TestToolFollowUpPromptWithInstructionKeepsFieldsAndVisualFirst(t *testing.T) {
	got := ToolFollowUpPromptWithInstruction("BODY", "Reply in markdown with every meaningful field.")
	for _, p := range []string{"BODY", "Visual-first", "every meaningful field"} {
		if !strings.Contains(got, p) {
			t.Fatalf("missing %q in:\n%s", p, got)
		}
	}
}
