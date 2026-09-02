package handlers

import (
	"strings"
	"testing"

	"idongivaflyinfa/models"
)

func TestContextFingerprintChangesWithPins(t *testing.T) {
	a := ContextFingerprint(true, true, true, []string{"aaa", "bbb"})
	b := ContextFingerprint(true, true, true, []string{"aaa", "bbb"})
	if a != b {
		t.Fatalf("same pins should match: %s vs %s", a, b)
	}
	c := ContextFingerprint(true, true, true, []string{"aaa"})
	if a == c {
		t.Fatal("pin change must not reuse the same fingerprint")
	}
	d := ContextFingerprint(true, true, false, []string{"aaa", "bbb"})
	if a == d {
		t.Fatal("include flag change must not reuse the same fingerprint")
	}
}

func TestCacheHitVsPinChangeMiss(t *testing.T) {
	h := &Handlers{}
	files := []models.AgentPinnedFile{{Path: "a.md", Hash: "h1", Content: "hello"}}
	fp := ContextFingerprint(true, false, false, pinHashesFrom(files))
	blob := AssemblePinnedBlob(files)
	h.putAgentCache("user1", "sess1", fp, blob)

	got, ok := h.getAgentCache("user1", "sess1", fp)
	if !ok || got != blob {
		t.Fatal("expected cache hit for unchanged pins")
	}

	files2 := []models.AgentPinnedFile{{Path: "a.md", Hash: "h1", Content: "hello"}, {Path: "b.md", Hash: "h2", Content: "world"}}
	fp2 := ContextFingerprint(true, false, false, pinHashesFrom(files2))
	if _, hit := h.getAgentCache("user1", "sess1", fp2); hit {
		t.Fatal("expected cache miss after pin change")
	}
	if _, leak := h.getAgentCache("other", "sess1", fp); leak {
		t.Fatal("cache must not leak another user's session")
	}
}

func TestRouteSubAgentsFanOutAndSimpleSkip(t *testing.T) {
	workers := RouteSubAgents("Summarize this folder and list my members", true, false, false)
	if len(workers) < 2 {
		t.Fatalf("expected fan-out, got %v", workers)
	}
	if RouteSubAgents("hi", true, true, true) != nil {
		t.Fatal("simple chat should skip sub-agents")
	}
}

func TestSubAgentsNeverWriteUserFolder(t *testing.T) {
	workers := RouteSubAgents("rewrite every file in the workspace folder", true, true, true)
	for _, w := range workers {
		if strings.Contains(strings.ToLower(w), "write") || w == "disk" {
			t.Fatalf("sub-agent must not write the user folder, got %q", w)
		}
	}
	if len(allowedWriteWorkers()) != 0 {
		t.Fatal("no disk-write workers allowed")
	}
}

func TestApplyAgentContextHonorsIncludeFlags(t *testing.T) {
	h := &Handlers{}
	off := false
	req := &models.ChatRequest{
		Message:          "Summarize this folder",
		IncludeFiles:     &off,
		IncludeNotes:     &off,
		IncludeKnowledge: &off,
		PinnedFiles:      []models.AgentPinnedFile{{Path: "a.md", Hash: "h1", Content: "SECRET_PIN"}},
	}
	got := h.applyAgentContext("u", "s", req)
	if strings.Contains(got.message, "SECRET_PIN") {
		t.Fatal("excluded files must not be in the prompt")
	}
	if got.includeKnowledge {
		t.Fatal("knowledge include should be off")
	}
}

func allowedWriteWorkers() []string {
	return nil
}
