package main

import "testing"

func TestNormalizeComposerDraftHTML(t *testing.T) {
	html := `<html><body><h1>Workshop</h1><p>Day one intro.</p><ul><li>Safety</li><li>Skills</li></ul></body></html>`
	got := normalizeComposerDraft(html)
	if got == "" {
		t.Fatal("expected markdown from html")
	}
	if !containsAll(got, "Workshop", "Day one intro", "Safety", "Skills") {
		t.Fatalf("unexpected markdown: %q", got)
	}
}

func TestParseComposerAIResponseLegacyHTML(t *testing.T) {
	raw := `{"assistant_message":"Draft ready.","proposed_email_html":"<h1>Title</h1><p>Body text</p>"}`
	msg, prop := parseComposerAIResponse(raw)
	if msg == "" {
		t.Fatal("expected assistant message")
	}
	if prop == nil || *prop == "" {
		t.Fatal("expected proposed markdown from legacy html field")
	}
	if !containsAll(*prop, "Title", "Body text") {
		t.Fatalf("unexpected draft: %q", *prop)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
