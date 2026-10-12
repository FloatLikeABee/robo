package mcp_test

import (
	"context"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"idongivaflyinfa/auth"
	"idongivaflyinfa/mcp"
)

func TestGrillAppAndTool(t *testing.T) {
	secret := "stdio-test-secret"
	cfg := auth.TokenConfig{Secret: []byte(secret), ExpiryHours: 24}
	tok := signToken(t, cfg, "ada-id", "ada@example.com", "ada", nil)
	id, err := mcp.ResolveIdentity(tok, cfg)
	if err != nil {
		t.Fatal(err)
	}
	complete := func(_ context.Context, prompt string) (string, error) {
		if strings.Contains(prompt, "Write the body") {
			return "Great! I'm Morph AI, feel free to let me know.", nil
		}
		return "What should the agent notebook include?\nWho is it for?\n", nil
	}
	server, cleanup, err := mcp.NewServerWith(id, nil, nil, nil, "", complete)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	session := connectInMemory(t, server, "2025-06-18")
	ctx := context.Background()

	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	tool := toolByName(t, listed.Tools, "note_grill")
	if tool.Annotations == nil || tool.Annotations.ReadOnlyHint {
		t.Fatalf("annotations = %+v", tool.Annotations)
	}
	nested, flat, visibility := toolUI(tool)
	if nested != "ui://morph/grill" || flat != "ui://morph/grill" || !containsString(visibility, "model") || !containsString(visibility, "app") {
		t.Fatalf("ui nested %q flat %q vis %#v", nested, flat, visibility)
	}

	resources, err := session.ListResources(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !resourceURI(resources.Resources, "ui://morph/notes") || !resourceURI(resources.Resources, "ui://morph/grill") {
		t.Fatalf("resources = %+v", resources.Resources)
	}
	read, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: "ui://morph/grill"})
	if err != nil {
		t.Fatal(err)
	}
	if read == nil || len(read.Contents) != 1 {
		t.Fatalf("read = %+v", read)
	}
	doc := read.Contents[0]
	if doc.MIMEType != "text/html;profile=mcp-app" || doc.URI != "ui://morph/grill" {
		t.Fatalf("content = %+v", doc)
	}
	assertClosedGrillHTML(t, doc.Text, tok, secret)
	assertNoExternalCSP(t, doc.Meta)

	started, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name: "note_grill",
		Arguments: map[string]any{
			"action": "start",
			"kind":   "note",
			"title":  "make this morphAI a platform for agents, actually agent notebook",
		},
	})
	if err != nil || started.IsError {
		t.Fatalf("start err=%v res=%+v", err, started)
	}
	text := toolText(t, started)
	if strings.Contains(strings.ToLower(text), "<html") || strings.Contains(text, "I'm Morph") {
		t.Fatalf("start = %s", text)
	}
	body := structuredMap(t, started.StructuredContent)
	gid, _ := body["id"].(string)
	if gid == "" || body["phase"] != "asking" {
		t.Fatalf("start = %#v", body)
	}

	beaTok := signToken(t, cfg, "bea-id", "bea@example.com", "bea", nil)
	beaID, err := mcp.ResolveIdentity(beaTok, cfg)
	if err != nil {
		t.Fatal(err)
	}
	beaServer, beaCleanup, err := mcp.NewServerWith(beaID, nil, nil, nil, "", complete)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(beaCleanup)
	bea := connectInMemory(t, beaServer, "2025-06-18")
	foreign, err := bea.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "note_grill",
		Arguments: map[string]any{"action": "finish", "id": gid},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !foreign.IsError {
		t.Fatalf("bea finish succeeded: %+v", foreign)
	}
	foreignText := toolText(t, foreign)
	if strings.Contains(foreignText, "notebook") || strings.Contains(strings.ToLower(foreignText), "forbidden") {
		t.Fatalf("bea = %q", foreignText)
	}

	done, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "note_grill",
		Arguments: map[string]any{"action": "finish", "id": gid},
	})
	if err != nil || done.IsError {
		t.Fatalf("finish err=%v res=%+v", err, done)
	}
	draft := toolText(t, done)
	if strings.Contains(draft, "I'm Morph") || strings.Contains(strings.ToLower(draft), "feel free") {
		t.Fatalf("draft = %s", draft)
	}
	if strings.Contains(strings.ToLower(draft), "<html") {
		t.Fatal("draft is html")
	}
}

func resourceURI(resources []*sdkmcp.Resource, uri string) bool {
	for _, resource := range resources {
		if resource != nil && resource.URI == uri {
			return true
		}
	}
	return false
}

func assertClosedGrillHTML(t *testing.T, html, token, secret string) {
	t.Helper()
	lower := strings.ToLower(html)
	for _, banned := range []string{"http:", "https:", "url(", "innerhtml", "insertadjacenthtml", "outerhtml", "document.write"} {
		if strings.Contains(lower, banned) {
			t.Fatalf("html contains %q", banned)
		}
	}
	if !strings.Contains(html, "Arial") || !strings.Contains(lower, "#070b14") {
		t.Fatal("html is not the dark Arial page")
	}
	if !strings.Contains(html, "[hidden] { display: none !important; }") {
		t.Fatal("author display rules can unhide the start form")
	}
	if !strings.Contains(html, "textContent") || !strings.Contains(html, "Content-Security-Policy") {
		t.Fatal("html missing text rendering or CSP")
	}
	if !strings.Contains(lower, "connect-src 'none'") || !strings.Contains(lower, "frame-src 'none'") || !strings.Contains(lower, "object-src 'none'") {
		t.Fatal("CSP does not block network, frames, and objects")
	}
	for _, name := range []string{"note_grill", "create_note", "tools/call"} {
		if !strings.Contains(html, name) {
			t.Fatalf("html missing %s", name)
		}
	}
	if strings.Contains(html, token) || strings.Contains(html, secret) || strings.Contains(html, "MORPH_MCP_TOKEN") {
		t.Fatal("html contains a token or secret")
	}
}
