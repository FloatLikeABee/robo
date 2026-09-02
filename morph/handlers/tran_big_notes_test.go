package handlers

import (
	"database/sql"
	"testing"

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
