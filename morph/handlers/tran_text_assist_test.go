package handlers

import (
	"strings"
	"testing"
)

var textAssistSchoolOpsLeftovers = []string{
	"transportation / school operations",
	"school transportation",
	"staff member",
	"route change, parent communication, vehicle check",
	"MorphData-style operations",
}

func assertNoSchoolOpsLeftover(t *testing.T, prompt string) {
	t.Helper()
	lower := strings.ToLower(prompt)
	for _, phrase := range textAssistSchoolOpsLeftovers {
		if strings.Contains(lower, strings.ToLower(phrase)) {
			t.Errorf("leftover school-ops phrase %q in prompt:\n%s", phrase, prompt)
		}
	}
}

func assertSeedLeadsPrompt(t *testing.T, prompt, seed string) {
	t.Helper()
	idx := strings.Index(prompt, seed)
	if idx < 0 {
		t.Fatalf("seed %q not in prompt:\n%s", seed, prompt)
	}
	if idx > 80 {
		t.Errorf("seed %q should appear near the start of the prompt (index %d):\n%s", seed, idx, prompt)
	}
}

func TestTextAssistGenerateTodoPromptIncludesSeedAndDropsSchoolOps(t *testing.T) {
	seed := "make money on AI stock"
	p := textAssistGenerateTodoPrompt(seed)
	assertSeedLeadsPrompt(t, p, seed)
	assertNoSchoolOpsLeftover(t, p)
}

func TestTextAssistGenerateNotePromptIncludesSeedAndDropsSchoolOps(t *testing.T) {
	seed := "make money on AI stock"
	p := textAssistGenerateNotePrompt(seed)
	assertSeedLeadsPrompt(t, p, seed)
	assertNoSchoolOpsLeftover(t, p)
}

func TestTextAssistGeneratePromptsEmptySeedHaveNoSchoolOpsExamples(t *testing.T) {
	assertNoSchoolOpsLeftover(t, textAssistGenerateTodoPrompt(""))
	assertNoSchoolOpsLeftover(t, textAssistGenerateNotePrompt(""))
}

func TestTextAssistTaskChainStepPromptIncludesTaskAndDropsSchoolOps(t *testing.T) {
	seed := "make money on AI stock"
	p := textAssistTaskChainStepPrompt(seed, "")
	if !strings.Contains(p, seed) {
		t.Fatalf("task text not in prompt:\n%s", p)
	}
	assertNoSchoolOpsLeftover(t, p)
}

func TestTextAssistImprovePromptIncludesTitle(t *testing.T) {
	seed := "make money on AI stock"
	p := textAssistImprovePrompt("todo", seed, "buy shares and watch the news")
	if !strings.Contains(p, seed) {
		t.Fatalf("title not in improve prompt:\n%s", p)
	}
	assertNoSchoolOpsLeftover(t, p)
}
