package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
)

func testPublishedPagesRepo(t *testing.T) (*PublishedPageRepository, func() int) {
	t.Helper()
	db, err := openComposerXSQLite(filepath.Join(t.TempDir(), "pages.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	opts := badger.DefaultOptions("").WithInMemory(true)
	opts.Logger = nil
	bdb, err := badger.Open(opts)
	if err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = bdb.Close()
		_ = db.Close()
	})
	count := func() int {
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM published_pages`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	return NewPublishedPageRepository(db, NewEmailContentStore(bdb)), count
}

func TestCreateSameNameUpdatesExistingRow(t *testing.T) {
	repo, count := testPublishedPagesRepo(t)
	ctx := context.Background()

	first, err := repo.Create(ctx, "Summer Launch", "dark", "<p>one</p>", 1)
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if first.Slug != "summer-launch" {
		t.Fatalf("slug=%q want summer-launch", first.Slug)
	}

	second, err := repo.Create(ctx, "Summer Launch", "light", "<p>two</p>", 1)
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("id=%d want %d (same row)", second.ID, first.ID)
	}
	if second.Slug != first.Slug {
		t.Fatalf("slug=%q want %q", second.Slug, first.Slug)
	}
	if count() != 1 {
		t.Fatalf("row count=%d want 1", count())
	}

	got, err := repo.GetBySlug(ctx, first.Slug)
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if got.HTMLContent != "<p>two</p>" {
		t.Fatalf("html=%q want <p>two</p>", got.HTMLContent)
	}
	if got.Theme != "light" {
		t.Fatalf("theme=%q want light", got.Theme)
	}
}

func TestCreateDistinctNamesStayDistinct(t *testing.T) {
	repo, count := testPublishedPagesRepo(t)
	ctx := context.Background()

	a, err := repo.Create(ctx, "Alpha Page", "default", "<p>a</p>", 1)
	if err != nil {
		t.Fatal(err)
	}
	b, err := repo.Create(ctx, "Beta Page", "default", "<p>b</p>", 1)
	if err != nil {
		t.Fatal(err)
	}
	if a.Slug == b.Slug {
		t.Fatalf("expected distinct slugs, both %q", a.Slug)
	}
	if count() != 2 {
		t.Fatalf("row count=%d want 2", count())
	}
}

func TestResolvePublishSlugReturnsExistingNotSuffix(t *testing.T) {
	repo, _ := testPublishedPagesRepo(t)
	ctx := context.Background()
	first, err := repo.Create(ctx, "Summer Launch", "default", "<p>one</p>", 1)
	if err != nil {
		t.Fatal(err)
	}
	slug, err := repo.ResolveUniqueSlug(ctx, "Summer Launch")
	if err != nil {
		t.Fatal(err)
	}
	if slug != first.Slug {
		t.Fatalf("resolved slug=%q want %q", slug, first.Slug)
	}
}

func TestPublishedPagesSlugUniqueConstraint(t *testing.T) {
	db, err := openComposerXSQLite(filepath.Join(t.TempDir(), "unique-slug.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`
INSERT INTO published_pages (name, slug, theme, content_mongo_id, created_by, created_at, updated_at)
VALUES ('A', 'same-slug', 'default', 'mongo-a', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
INSERT INTO published_pages (name, slug, theme, content_mongo_id, created_by, created_at, updated_at)
VALUES ('B', 'same-slug', 'default', 'mongo-b', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	if err == nil {
		t.Fatal("expected unique slug constraint to reject the second insert")
	}
	if !isUniqueConstraint(err) {
		t.Fatalf("expected unique constraint error, got %v", err)
	}
}

func TestCreatePublishedPageHandlerRepublishSamePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo, count := testPublishedPagesRepo(t)
	app := &App{publishedPages: repo}
	r := gin.New()
	r.POST("/publishes", app.createPublishedPage)

	post := func(html string) map[string]any {
		t.Helper()
		body, err := json.Marshal(map[string]string{
			"name":         "Summer Launch",
			"theme":        "default",
			"html_content": html,
		})
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/publishes", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated && w.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		var payload map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		return payload
	}

	first := post("<p>one</p>")
	second := post("<p>two</p>")
	if first["slug"] != second["slug"] {
		t.Fatalf("slug %v vs %v", first["slug"], second["slug"])
	}
	if first["public_path"] != second["public_path"] {
		t.Fatalf("path %v vs %v", first["public_path"], second["public_path"])
	}
	if second["public_path"] != "/public/p/summer-launch" {
		t.Fatalf("public_path=%v", second["public_path"])
	}
	if count() != 1 {
		t.Fatalf("row count=%d want 1", count())
	}
}

func TestDeletePublishedPageThenPublic404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo, count := testPublishedPagesRepo(t)
	app := &App{publishedPages: repo}
	r := gin.New()
	r.POST("/publishes", app.createPublishedPage)
	r.DELETE("/publishes/:id", app.deletePublishedPage)
	r.GET("/public/p/:slug", app.servePublishedPage)

	body, err := json.Marshal(map[string]string{
		"name":         "Summer Launch",
		"theme":        "default",
		"html_content": "<p>old-html</p>",
	})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/publishes", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("publish status=%d body=%s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	id, ok := created["id"].(float64)
	if !ok || id <= 0 {
		t.Fatalf("id=%v", created["id"])
	}
	slug, _ := created["slug"].(string)
	if slug == "" {
		t.Fatalf("missing slug in %v", created)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/public/p/"+slug, nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("public GET before delete status=%d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "old-html") {
		t.Fatalf("public GET missing html: %s", w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/publishes/%d", int64(id)), nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", w.Code, w.Body.String())
	}
	if count() != 0 {
		t.Fatalf("row count=%d want 0", count())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/public/p/"+slug, nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("public GET after delete status=%d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "old-html") {
		t.Fatal("old HTML still served after delete")
	}
}
