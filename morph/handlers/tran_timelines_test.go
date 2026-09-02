package handlers

import (
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestTimelineFileExtOK(t *testing.T) {
	cases := map[string]bool{
		"a.txt":       true,
		"a.PDF":       true,
		"notes.md":    true,
		"x.markdown":  true,
		"photo.png":   false,
		"doc.docx":    false,
		"noext":       false,
		"archive.tar": false,
	}
	for name, want := range cases {
		if got := timelineFileExtOK(name); got != want {
			t.Errorf("timelineFileExtOK(%q)=%v want %v", name, got, want)
		}
	}
}

func TestIsBlockedTimelineHost(t *testing.T) {
	blocked := []string{"localhost", "127.0.0.1", "0.0.0.0", "10.0.0.1", "192.168.1.1", "169.254.169.254"}
	for _, h := range blocked {
		if !isBlockedTimelineHost(h) {
			t.Errorf("expected blocked host %q", h)
		}
	}
	if isBlockedTimelineHost("example.com") {
		t.Errorf("example.com should not be blocked by IP private rules alone")
	}
}

func TestExtractTimelineFileTextTxt(t *testing.T) {
	text, err := extractTimelineFileText("sample.txt", []byte("  Hello timeline  "))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "Hello timeline") {
		t.Fatalf("unexpected text %q", text)
	}
	if _, err := extractTimelineFileText("x.png", []byte("nope")); err == nil {
		t.Fatal("expected unsupported type error")
	}
	if _, err := extractTimelineFileText("empty.txt", []byte("   ")); err == nil {
		t.Fatal("expected empty text error")
	}
}

func TestBuildTimelineHTML(t *testing.T) {
	html := buildTimelineHTML("Demo", "## 2020\n\n- Event one")
	if !strings.Contains(html, "Demo") || !strings.Contains(html, "Event one") {
		t.Fatalf("html missing content: %s", truncateRunes(html, 200))
	}
	if !strings.Contains(html, "<!doctype html>") {
		t.Fatal("expected full html document")
	}
}

func TestScanTimelineAcceptsSQLiteTextTimestamps(t *testing.T) {
	db, err := sql.Open("sqlite", "file:timeline-scan?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`CREATE TABLE timeline (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		owner_key TEXT NOT NULL,
		title TEXT NOT NULL,
		source_summary TEXT NOT NULL,
		source_file_name TEXT NULL,
		source_url TEXT NULL,
		has_paste INTEGER NOT NULL,
		markdown_content TEXT NOT NULL,
		html_content TEXT NOT NULL,
		published_slug TEXT NULL,
		published_path TEXT NULL,
		created_on TEXT,
		last_updated TEXT
	)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO timeline (
		id, user_id, owner_key, title, source_summary, source_file_name, source_url, has_paste,
		markdown_content, html_content, published_slug, published_path, created_on, last_updated
	) VALUES (1, 9, 'owner', 'Apollo', 'file:notes.txt', NULL, NULL, 0, '# md', '<p>html</p>', NULL, NULL, '2026-08-23 10:00:00', '2026-08-23 11:30:00')`)
	if err != nil {
		t.Fatal(err)
	}

	row := db.QueryRow(`SELECT ` + timelineSelectCols + ` FROM timeline WHERE id = 1`)
	got, err := scanTimeline(row)
	if err != nil {
		t.Fatalf("scanTimeline: %v", err)
	}
	if got.Title != "Apollo" || got.UserID != 9 || got.OwnerKey != "owner" {
		t.Fatalf("unexpected identity %#v", got)
	}
	if got.CreatedOn.Year() != 2026 || got.CreatedOn.Month() != 8 || got.CreatedOn.Day() != 23 {
		t.Fatalf("created_on not parsed: %v", got.CreatedOn)
	}
	if got.LastUpdated.Hour() != 11 || got.LastUpdated.Minute() != 30 {
		t.Fatalf("last_updated not parsed: %v", got.LastUpdated)
	}
}
