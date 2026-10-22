package handlers

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestScanBigNoteAcceptsSQLiteTextTimestamps(t *testing.T) {
	db, err := sql.Open("sqlite", "file:bignote-scan?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`CREATE TABLE big_note (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		owner_key TEXT NOT NULL,
		title TEXT NOT NULL,
		idea TEXT NOT NULL,
		note_kind TEXT NOT NULL,
		markdown_content TEXT NOT NULL,
		html_content TEXT NOT NULL,
		questions_json TEXT NULL,
		theme TEXT NOT NULL,
		published_slug TEXT NULL,
		published_path TEXT NULL,
		created_on TEXT,
		last_updated TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO big_note (
		id, user_id, owner_key, title, idea, note_kind, markdown_content, html_content,
		questions_json, theme, published_slug, published_path, created_on, last_updated
	) VALUES (1, 9, 'owner', 'Note', 'idea', 'note', '# md', '<p>html</p>', NULL, 'default', NULL, NULL, '2026-08-23 10:00:00', '2026-08-23 11:30:00')`)
	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow(`SELECT ` + bigNoteSelectCols + ` FROM big_note WHERE id = 1`)
	got, err := scanBigNote(row)
	if err != nil {
		t.Fatalf("scanBigNote: %v", err)
	}
	if got.Title != "Note" || got.UserID != 9 {
		t.Fatalf("unexpected identity %#v", got)
	}
	if got.CreatedOn.Year() != 2026 || got.CreatedOn.Month() != 8 || got.CreatedOn.Day() != 23 {
		t.Fatalf("created_on not parsed: %v", got.CreatedOn)
	}
	if got.LastUpdated.Hour() != 11 || got.LastUpdated.Minute() != 30 {
		t.Fatalf("last_updated not parsed: %v", got.LastUpdated)
	}
}

func TestPersistBigNoteMarkdownRegeneratesHTML(t *testing.T) {
	db, err := sql.Open("sqlite", "file:bignote-md-patch?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`CREATE TABLE big_note (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		owner_key TEXT NOT NULL,
		title TEXT NOT NULL,
		idea TEXT NOT NULL,
		note_kind TEXT NOT NULL,
		markdown_content TEXT NOT NULL,
		html_content TEXT NOT NULL,
		questions_json TEXT NULL,
		theme TEXT NOT NULL,
		published_slug TEXT NULL,
		published_path TEXT NULL,
		created_on TEXT,
		last_updated TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO big_note (
		id, user_id, owner_key, title, idea, note_kind, markdown_content, html_content,
		questions_json, theme, created_on, last_updated
	) VALUES (1, 1, 'o', 'Note', 'idea', 'note', 'old', 'old-html', NULL, 'dark', '2026-01-01 00:00:00', '2026-01-01 00:00:00')`)
	if err != nil {
		t.Fatal(err)
	}
	n := bigNote{ID: 1, Title: "Note", NoteKind: "note", Theme: "dark", MarkdownContent: "old"}
	if err := persistBigNoteMarkdown(db, &n, "## Hello **world**"); err != nil {
		t.Fatal(err)
	}
	if n.MarkdownContent != "## Hello **world**" {
		t.Fatalf("note markdown: %q", n.MarkdownContent)
	}
	if n.HTMLContent == "old-html" || n.HTMLContent == "" {
		t.Fatal("expected regenerated html")
	}
	var md, html string
	if err := db.QueryRow(`SELECT markdown_content, html_content FROM big_note WHERE id = 1`).Scan(&md, &html); err != nil {
		t.Fatal(err)
	}
	if md != "## Hello **world**" || html != n.HTMLContent {
		t.Fatalf("db mismatch md=%q", md)
	}
}

func TestBigNoteHTMLStaysDarkSansEvenWhenThemeSaysLight(t *testing.T) {
	html := buildBigNoteHTML("Agent", "A short note.", "story", nil, "light")
	for _, needle := range []string{"Arial", "#38bdf8", "#a78bfa", "#34d399", "color-scheme:dark"} {
		if !strings.Contains(html, needle) {
			t.Fatalf("missing %q", needle)
		}
	}
	if strings.Contains(html, "Georgia") || strings.Contains(html, "#fafaf9") {
		t.Fatal(html)
	}
}

func TestBigNoteGeneratePromptGroundsIdea(t *testing.T) {
	idea := "Field guide: how agents use Dorph notes, kinds, sharing pool and safety"
	got := bigNoteGeneratePrompt(idea, "dark", "note", "", "")
	if !strings.Contains(got, idea) {
		t.Fatal("prompt dropped the idea")
	}
	low := strings.ToLower(got)
	if !strings.Contains(low, "do not invent") && !strings.Contains(low, "forbid invent") {
		t.Fatalf("prompt must forbid invented features:\n%s", got)
	}
	if !strings.Contains(got, "##") || (!strings.Contains(low, "at least 3") && !strings.Contains(low, "at least three") && !strings.Contains(low, "three ##")) {
		t.Fatalf("prompt must require at least three ## sections:\n%s", got)
	}
}

func TestBigNoteTimeoutIsRetryable(t *testing.T) {
	if bigNoteGenerateTimeout <= 90*time.Second {
		t.Fatalf("timeout = %s, want longer than 90s for Ollama", bigNoteGenerateTimeout)
	}
	err := wrapBigNoteAIError(context.DeadlineExceeded)
	if err == nil || strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "retry") {
		t.Fatalf("err = %v, want retryable", err)
	}
}

func TestResolveBigNoteExpandsTwoH2WithH3(t *testing.T) {
	idea := "PR154 run 3: how to plan a 2-day team offsite"
	raw := `{"title":"How to Plan a 2-Day Team Offsite","markdown":"## Introduction\n\nPlan the offsite.\n\n## Key Milestones\n\n### Define objectives\n\nA.\n\n### Create a schedule\n\nB.\n\n### Gather supplies\n\nC."}`
	out, err := resolveBigNoteFromReplies(idea, "dark", "note", []string{raw})
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, line := range strings.Split(out.Markdown, "\n") {
		tline := strings.TrimSpace(line)
		if strings.HasPrefix(tline, "## ") && !strings.HasPrefix(tline, "###") {
			n++
		}
	}
	if n < 3 {
		t.Fatalf("want ≥3 ## headings, got %d in:\n%s", n, out.Markdown)
	}
	if !strings.Contains(out.Markdown, "offsite") {
		t.Fatalf("expand dropped the idea: %q", out.Markdown)
	}
}

func TestResolveBigNoteExpandsOneParagraph(t *testing.T) {
	idea := "PR152 run 1: field guide to Dorph notes for agents: note kinds, the consent-based sharing pool, and the secret/PII block"
	thin := `{"title":"Creating DorphNotes: a Field Guide","markdown":"DorphNotes is designed to support secure, consent-based sharing pools. This guide will help agents understand the key features, including note kinds, the consent-based sharing pool, and the secret/PII block."}`
	out, err := resolveBigNoteFromReplies(idea, "dark", "note", []string{thin})
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(out.Markdown, "## "); n < 3 {
		t.Fatalf("want ≥3 ## sections, got %d in:\n%s", n, out.Markdown)
	}
	if !strings.Contains(out.Markdown, "sharing pool") && !strings.Contains(out.Markdown, "note kinds") {
		t.Fatalf("expand dropped the idea: %q", out.Markdown)
	}
}
