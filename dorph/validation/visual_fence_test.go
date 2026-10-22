package validation

import "testing"

func TestIsValidPromptAllowsMermaidAndPixelFences(t *testing.T) {
	msg := "Reply with a diagram.\n\n```mermaid\nflowchart LR\n  A[Start] --> B[Done]\n```\n\n```pixel\n#ff0000 .\n. #00ff00\n```"
	if !HasVisualFence(msg) {
		t.Fatal("expected visual fence")
	}
	if !IsValidPrompt(msg) {
		t.Fatal("mermaid/pixel prompt must not be rejected as gibberish")
	}
}
