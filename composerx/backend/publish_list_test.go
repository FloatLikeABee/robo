package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublishedPagesListScansSQLiteTextTimestamps(t *testing.T) {
	db, err := openComposerXSQLite(filepath.Join(t.TempDir(), "pages.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`
INSERT INTO published_pages (name, slug, theme, content_mongo_id, created_by, created_at, updated_at)
VALUES ('Launch page', 'launch-page', 'default', 'mongo-1', 1, '2026-08-31 04:00:00', '2026-08-31 04:10:00')`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewPublishedPageRepository(db, nil)
	items, total, err := repo.List(context.Background(), 50, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 {
		t.Fatalf("total=%d want 1", total)
	}
	if len(items) != 1 {
		t.Fatalf("len(items)=%d want 1", len(items))
	}
	if items[0].Name != "Launch page" || items[0].Slug != "launch-page" {
		t.Fatalf("got name=%q slug=%q", items[0].Name, items[0].Slug)
	}
	if items[0].ID <= 0 {
		t.Fatalf("expected id, got %d", items[0].ID)
	}
}

func TestPublishedPagesListEmpty(t *testing.T) {
	db, err := openComposerXSQLite(filepath.Join(t.TempDir(), "empty.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPublishedPageRepository(db, nil)
	items, total, err := repo.List(context.Background(), 50, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 0 {
		t.Fatalf("total=%d want 0", total)
	}
	if len(items) != 0 {
		t.Fatalf("len(items)=%d want 0", len(items))
	}
}

func TestPublishDraftsListScansSQLiteTextTimestamps(t *testing.T) {
	db, err := openComposerXSQLite(filepath.Join(t.TempDir(), "drafts.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`
INSERT INTO publish_drafts (name, theme, html_content, created_by, created_at, updated_at)
VALUES ('WIP page', 'default', '<p>draft</p>', 1, '2026-08-31 04:00:00', '2026-08-31 04:10:00')`)
	if err != nil {
		t.Fatal(err)
	}

	repo := NewPublishDraftRepository(db)
	items, total, err := repo.List(context.Background(), 50, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("total=%d len=%d", total, len(items))
	}
	if items[0].Name != "WIP page" {
		t.Fatalf("name=%q", items[0].Name)
	}
	if items[0].ID <= 0 {
		t.Fatalf("expected id, got %d", items[0].ID)
	}

	got, err := repo.GetByID(context.Background(), items[0].ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !strings.Contains(got.HTMLContent, "draft") {
		t.Fatalf("html=%q", got.HTMLContent)
	}
}
