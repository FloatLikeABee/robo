package handlers

import (
	"strings"
	"testing"
)

func TestSelectedSkillIncludedWhenAbsentFromEnabledCatalog(t *testing.T) {
	enabledOnly := []skillPromptRow{{
		ID:           "builtin-research",
		Name:         "Research",
		Instructions: "search first",
	}}
	_ = enabledOnly
	out := selectedSkillInstructions([]string{"custom-1"}, []skillPromptRow{{
		ID:           "custom-1",
		Name:         "Custom",
		Instructions: "Do the custom thing.",
	}})
	if !strings.Contains(out, "Follow these selected skills for this reply.") || !strings.Contains(out, "Do the custom thing.") {
		t.Fatalf("selected skill missing from prompt:\n%s", out)
	}
	if strings.Contains(out, "search first") {
		t.Fatalf("enabled-catalog body should not be required in the selected block:\n%s", out)
	}
}

func TestChatRAGWithoutAssistantIncludesSelected(t *testing.T) {
	got := collectionsForChatRAG(nil, []string{"docs"})
	if len(got) != 1 || got[0] != "docs" {
		t.Fatalf("got %v", got)
	}
}

func TestChatRAGKeepsSelectedBesideAssistantCollections(t *testing.T) {
	got := collectionsForChatRAG([]string{"a", "b", "c", "d"}, []string{"docs"})
	joined := strings.Join(got, ",")
	if !strings.Contains(joined, "docs") || !strings.Contains(joined, "a") {
		t.Fatalf("got %v", got)
	}
}
