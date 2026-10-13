package mcp_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"idongivaflyinfa/mcp"
)

func TestSuggestNoteKind(t *testing.T) {
	longHint := strings.Repeat("We shipped the grill, kept notes private, and confirmed PAT setup. ", 8)
	cases := []struct {
		intent, hint, kind string
		grill              bool
	}{
		{"Calendar reminder Friday 9am: dentist", "", "stick_note", false},
		{"Stick a reminder on Thursday 15:00: standup", "", "stick_note", false},
		{"Don't forget at 17:00 to take the cake out", "", "stick_note", false},
		{"Assign the login bug to Bea and put it in Doing", "", "task", false},
		{"Move the onboarding card to In Review and assign it to Sam", "", "task", false},
		{"Add a high-priority bug to the MorphNotes board", "", "task", false},
		{"Kanban: design → build → qa for the launch", "", "task", false},
		{"Trip itinerary for Japan: Day 1 Tokyo, Day 2 Kyoto", "", "timeline", false},
		{"Project history from kickoff to launch with dates", "", "timeline", false},
		{"Weekly plan: Mon draft, Wed review, Fri ship", "", "timeline", false},
		{"Chronology of the outage from 09:00 to 14:00", "", "timeline", false},
		{"Log the incident: API 500 at 14:02", "", "event_log", false},
		{"Decision log: we picked alias attribution", "", "event_log", false},
		{"Call the dentist to reschedule", "", "todo", false},
		{"Buy milk tomorrow", "", "todo", false},
		{"Send the invoice to Acme by Friday", "", "todo", false},
		{"TODO: fix the leaky tap", "", "todo", false},
		{"I need to pick up the kids after school", "", "todo", false},
		{"Research local LLM options and cite sources", "", "research", false},
		{"Investigate why the login fails; include findings and links", "", "research", false},
		{"Compare SQLite FTS vs embeddings with sources", "", "research", false},
		{"Draft the full story of the migration, chapter by chapter", "", "big_note", false},
		{"Write a chaptered guide to our onboarding (over 2000 words)", "", "big_note", false},
		{"Long-form narrative of the company founding", "", "big_note", false},
		{"API base URL is http://127.0.0.1:9090", "", "note", false},
		{"Meeting summary: we chose CC BY 4.0", "", "note", false},
		{"Password reset goes through the users panel", "", "note", false},
		{"Handoff: grill sessions live in memory", "", "note", false},
		{"write something", "", "note", true},
		{"help me write a note about the launch", "", "note", true},
		{"todo", "", "todo", true},
		{"help me write something to share with the pool", "", "note", true},
		{"Save a session summary", longHint, "note", false},
	}
	if len(cases) < 30 {
		t.Fatalf("table has %d rows, want at least 30", len(cases))
	}
	for _, c := range cases {
		got, err := mcp.SuggestNoteKind(c.intent, c.hint)
		if err != nil {
			t.Fatalf("%q: %v", c.intent, err)
		}
		if got.Kind != c.kind || got.UseGrill != c.grill {
			t.Errorf("%q: kind=%s grill=%v want %s %v (%s)", c.intent, got.Kind, got.UseGrill, c.kind, c.grill, got.Reason)
		}
		if strings.TrimSpace(got.Reason) == "" {
			t.Errorf("%q: empty reason", c.intent)
		}
		if got.Confidence <= 0 || got.Confidence > 1 {
			t.Errorf("%q: confidence %v", c.intent, got.Confidence)
		}
	}
}

func TestSuggestNoteKindEmptyIntent(t *testing.T) {
	_, err := mcp.SuggestNoteKind("  ", "")
	if err == nil || !strings.Contains(err.Error(), "invalid_input") {
		t.Fatalf("err = %v", err)
	}
}

func TestSuggestNoteKindHasNoAIImport(t *testing.T) {
	b, err := os.ReadFile("agent_runtime.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(b)
	for _, banned := range []string{"github.com/robo/morphai", "ollama", "ChatCompletion"} {
		if strings.Contains(src, banned) {
			t.Fatalf("agent_runtime.go imports or calls %s", banned)
		}
	}
}

func TestAgentRuntimeToolCount(t *testing.T) {
	if mcp.AgentRuntimeToolCount != 2 {
		t.Fatalf("AgentRuntimeToolCount = %d", mcp.AgentRuntimeToolCount)
	}
	b, err := os.ReadFile("records_tools.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "const ToolCount = 26") {
		t.Fatal("records_tools.go ToolCount changed")
	}
	if mcp.PoolToolCount != 9 {
		t.Fatalf("PoolToolCount = %d", mcp.PoolToolCount)
	}
}

func TestGuideDocMatchesEmbed(t *testing.T) {
	doc, err := os.ReadFile(filepath.Join(repoRoot(t), "docs", "agents", "16-morph-for-agents.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(doc) != mcp.AgentGuideMarkdown {
		t.Fatal("docs/agents/16-morph-for-agents.md must match morph/mcp/agent_guide.md")
	}
	if !strings.Contains(mcp.AgentGuideMarkdown, "Never share without explicit human") {
		t.Fatal("guide missing consent rule")
	}
}

func TestMorphGuideTopics(t *testing.T) {
	outline, err := mcp.GuideExcerpt("")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(outline, "Golden rules") || !strings.Contains(outline, "morph://guide") {
		t.Fatalf("outline = %s", outline)
	}
	sharing, err := mcp.GuideExcerpt("sharing")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(sharing), "never share without explicit human") {
		t.Fatalf("sharing = %s", sharing)
	}
	_, err = mcp.GuideExcerpt("nope")
	if err == nil || !strings.Contains(err.Error(), "invalid_input") {
		t.Fatalf("unknown topic err = %v", err)
	}
	for _, topic := range []string{"kinds", "writing", "grill", "checking", "sharing", "pool", "errors", "limits", "setup", "privacy", "tools"} {
		got, err := mcp.GuideExcerpt(topic)
		if err != nil {
			t.Fatalf("%s: %v", topic, err)
		}
		if strings.TrimSpace(got) == "" {
			t.Fatalf("%s empty", topic)
		}
	}
}

func TestAgentRuntimeToolsAndGuideResource(t *testing.T) {
	id, err := mcp.ResolveIdentity(signToken(t, testTokenConfig(), "user-1", "ada@example.com", "ada", nil), testTokenConfig())
	if err != nil {
		t.Fatal(err)
	}
	server, cleanup, err := mcp.NewServer(id, nil, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	session := connectInMemory(t, server, "2025-06-18")
	init := session.InitializeResult()
	if init == nil {
		t.Fatal("missing initialize")
	}
	if len(init.Instructions) >= 2000 {
		t.Fatalf("instructions length %d", len(init.Instructions))
	}
	low := strings.ToLower(init.Instructions)
	if !strings.Contains(low, "read-only") || !strings.Contains(low, "morph://guide") || !strings.Contains(low, "never share") {
		t.Fatalf("instructions = %q", init.Instructions)
	}

	listed, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tools) != mcp.AdvertisedToolCount() {
		t.Fatalf("tools = %d", len(listed.Tools))
	}
	for _, name := range []string{"suggest_note_kind", "morph_guide"} {
		tool := toolByName(t, listed.Tools, name)
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Fatalf("%s annotations = %+v", name, tool.Annotations)
		}
	}

	kindRes, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "suggest_note_kind",
		Arguments: map[string]any{"intent": "Buy milk tomorrow"},
	})
	if err != nil || kindRes.IsError {
		t.Fatalf("suggest err=%v res=%+v", err, kindRes)
	}
	got := structuredMap(t, kindRes.StructuredContent)
	if got["kind"] != "todo" {
		t.Fatalf("suggest = %#v", got)
	}

	empty, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "suggest_note_kind",
		Arguments: map[string]any{"intent": ""},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !empty.IsError || !strings.Contains(toolText(t, empty), "invalid_input") {
		t.Fatalf("empty suggest = %+v", empty)
	}

	resources, err := session.ListResources(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !resourceURI(resources.Resources, "morph://guide") {
		t.Fatalf("resources = %+v", resources.Resources)
	}
	read, err := session.ReadResource(context.Background(), &sdkmcp.ReadResourceParams{URI: "morph://guide"})
	if err != nil || read == nil || len(read.Contents) != 1 {
		t.Fatalf("read = %+v err=%v", read, err)
	}
	doc := read.Contents[0]
	if doc.MIMEType != "text/markdown" || doc.URI != "morph://guide" {
		t.Fatalf("content = %+v", doc)
	}
	if !strings.Contains(doc.Text, "Golden rules") {
		t.Fatal("guide resource missing golden rules")
	}

	guideRes, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "morph_guide",
		Arguments: map[string]any{"topic": "sharing"},
	})
	if err != nil || guideRes.IsError {
		t.Fatalf("guide err=%v res=%+v", err, guideRes)
	}
	g := structuredMap(t, guideRes.StructuredContent)
	md, _ := g["markdown"].(string)
	if !strings.Contains(strings.ToLower(md), "never share without explicit human") {
		t.Fatalf("morph_guide sharing = %#v", g)
	}
}
