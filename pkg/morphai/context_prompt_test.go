package morphai

import (
	"strings"
	"testing"
)

func TestToolFollowUpPromptUsesReplyShape(t *testing.T) {
	got := ToolFollowUpPrompt("TOOL_RESULT ok")
	for _, p := range []string{"TOOL_RESULT ok", "ordinary question", "Never call an image API", "one JSON object"} {
		if !strings.Contains(got, p) {
			t.Fatalf("missing %q in:\n%s", p, got)
		}
	}
	if strings.Contains(got, "Visual-first") || strings.Contains(got, "short caption") {
		t.Fatalf("diagram-first instruction leaked:\n%s", got)
	}
}

func TestToolFollowUpPromptWithInstructionKeepsFields(t *testing.T) {
	got := ToolFollowUpPromptWithInstruction("BODY", "Reply in markdown with every meaningful field.")
	for _, p := range []string{"BODY", "ordinary question", "every meaningful field"} {
		if !strings.Contains(got, p) {
			t.Fatalf("missing %q in:\n%s", p, got)
		}
	}
}

func TestModelIdentityInstructionsNamesConfig(t *testing.T) {
	got := ModelIdentityInstructions("ollama", "llama3.2")
	if !strings.Contains(got, "provider ollama") || !strings.Contains(got, "model llama3.2") {
		t.Fatal(got)
	}
	if strings.Contains(got, "5.1") || strings.Contains(got, "2023-10-25") {
		t.Fatal(got)
	}
}

func TestModelIdentityInstructionsUnknown(t *testing.T) {
	got := ModelIdentityInstructions("  ", "")
	if !strings.Contains(got, "do not know") {
		t.Fatal(got)
	}
}

func TestModelIdentityPartial(t *testing.T) {
	got := ModelIdentityInstructions("ollama", "")
	if !strings.Contains(got, "provider ollama") || !strings.Contains(got, "model unknown") {
		t.Fatal(got)
	}
}

func TestClientModelIdentityWithoutKey(t *testing.T) {
	c := NewClient(Config{Provider: ProviderOllama, Model: "llama3.2"})
	provider, model := c.ModelIdentity()
	if provider != "ollama" || model != "llama3.2" {
		t.Fatalf("provider %q model %q", provider, model)
	}
}

func TestReplyShapeAliasMatches(t *testing.T) {
	if VisualFirstInstructions != ReplyShapeInstructions {
		t.Fatal("VisualFirstInstructions must be the reply-shape text")
	}
	if strings.Contains(ReplyShapeInstructions, "short caption") {
		t.Fatal(ReplyShapeInstructions)
	}
}
