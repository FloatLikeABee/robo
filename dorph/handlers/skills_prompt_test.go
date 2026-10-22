package handlers

import (
	"strings"
	"testing"
)

func TestAlwaysOnSkillsDoNotOrderDiagrams(t *testing.T) {
	var b strings.Builder
	(&Handlers{}).appendAlwaysOnAgentSkills(&b)
	got := b.String()
	if !strings.Contains(got, "Research") || !strings.Contains(got, "Design") {
		t.Fatalf("missing default skills:\n%s", got)
	}
	if strings.Contains(got, "### Graphs") || strings.Contains(strings.ToLower(got), "diagram first") || strings.Contains(got, "short caption") {
		t.Fatalf("graphs or diagram-first leaked:\n%s", got)
	}
}

func TestPreferAlwaysOnInstructionsDropsStaleDiagramOrder(t *testing.T) {
	fallback := "Use a diagram only when it makes that structure clearer."
	got := preferAlwaysOnInstructions("Diagram first (mermaid) for structure.", fallback)
	if got != fallback {
		t.Fatal(got)
	}
	got = preferAlwaysOnInstructions("Visual-first: short caption, then a mermaid diagram.", fallback)
	if got != fallback {
		t.Fatal(got)
	}
	custom := "Cite the operator's notes and skip diagrams."
	if preferAlwaysOnInstructions(custom, fallback) != custom {
		t.Fatal("custom research text should stay")
	}
}
