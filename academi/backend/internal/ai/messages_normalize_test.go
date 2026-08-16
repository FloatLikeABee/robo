package ai

import "testing"

func TestNormalizeChatMessages_mergesLeadingSystem(t *testing.T) {
	in := []ChatMessage{
		{Role: "system", Content: "You are Academi."},
		{Role: "system", Content: "Attached doc:\nhello"},
		{Role: "assistant", Content: "Hi"},
		{Role: "user", Content: "research the doc"},
	}
	out := normalizeChatMessages(in)
	if len(out) != 3 {
		t.Fatalf("want 3 messages, got %d", len(out))
	}
	if out[0].Role != "system" || !contains(out[0].Content, "You are Academi.") || !contains(out[0].Content, "Attached doc") {
		t.Fatalf("unexpected merged system: %+v", out[0])
	}
	if out[1].Role != "assistant" || out[2].Role != "user" {
		t.Fatalf("unexpected order: %+v", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
