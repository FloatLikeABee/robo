package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type PublishedPage struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Theme     string `json:"theme"`
	CreatedBy int64  `json:"created_by"`
}

type PublishedPageDetail struct {
	PublishedPage
	HTMLContent string `json:"html_content"`
}

type PublishedPageListRow struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Theme     string    `json:"theme"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PublishedPageRepository struct {
	db      *sql.DB
	content *EmailContentStore
}

func NewPublishedPageRepository(db *sql.DB, content *EmailContentStore) *PublishedPageRepository {
	return &PublishedPageRepository{db: db, content: content}
}

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

func slugifyPublishName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = nonSlugChars.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "page"
	}
	return s
}

func (r *PublishedPageRepository) nextUniqueSlug(ctx context.Context, base string) (string, error) {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "page"
	}

	candidate := base
	for i := 2; i <= 5000; i++ {
		var exists int
		err := r.db.QueryRowContext(ctx, `SELECT 1 FROM published_pages WHERE slug = ? LIMIT 1`, candidate).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return "", errors.New("could not find available publish slug")
}

func (r *PublishedPageRepository) ResolveUniqueSlug(ctx context.Context, name string) (string, error) {
	return r.nextUniqueSlug(ctx, slugifyPublishName(name))
}

func (r *PublishedPageRepository) nextUniqueName(ctx context.Context, base string) (string, error) {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "Untitled page"
	}
	candidate := base
	for i := 2; i <= 5000; i++ {
		var exists int
		err := r.db.QueryRowContext(ctx, `SELECT 1 FROM published_pages WHERE name = ? LIMIT 1`, candidate).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
		candidate = fmt.Sprintf("%s (%d)", base, i)
	}
	return "", errors.New("could not find available publish name")
}

func (r *PublishedPageRepository) Create(ctx context.Context, name, theme, html string, createdBy int64) (*PublishedPage, error) {
	name = strings.TrimSpace(name)
	theme = strings.TrimSpace(theme)
	if name == "" || strings.TrimSpace(html) == "" {
		return nil, errors.New("name and html required")
	}
	if createdBy <= 0 {
		createdBy = 1
	}
	if theme == "" {
		theme = "default"
	}

	uniqueName, err := r.nextUniqueName(ctx, name)
	if err != nil {
		return nil, err
	}
	slug, err := r.ResolveUniqueSlug(ctx, uniqueName)
	if err != nil {
		return nil, err
	}

	mongoID, err := r.content.InsertHTML(ctx, html, "")
	if err != nil {
		return nil, err
	}

	const ins = `
INSERT INTO published_pages (name, slug, theme, content_mongo_id, created_by, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	res, err := r.db.ExecContext(ctx, ins, uniqueName, slug, theme, mongoID, createdBy)
	if err != nil {
		_ = r.content.DeleteByHexID(ctx, mongoID)
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		_ = r.content.DeleteByHexID(ctx, mongoID)
		return nil, err
	}
	return &PublishedPage{
		ID:        id,
		Name:      uniqueName,
		Slug:      slug,
		Theme:     theme,
		CreatedBy: createdBy,
	}, nil
}

func (r *PublishedPageRepository) GetBySlug(ctx context.Context, slug string) (*PublishedPageDetail, error) {
	const q = `
SELECT id, name, slug, theme, content_mongo_id, created_by
FROM published_pages
WHERE slug = ?
LIMIT 1`

	var out PublishedPageDetail
	var mongoID string
	if err := r.db.QueryRowContext(ctx, q, strings.TrimSpace(slug)).Scan(
		&out.ID,
		&out.Name,
		&out.Slug,
		&out.Theme,
		&mongoID,
		&out.CreatedBy,
	); err != nil {
		return nil, err
	}
	html, _, err := r.content.GetHTML(ctx, mongoID)
	if err != nil {
		return nil, err
	}
	out.HTMLContent = html
	return &out, nil
}

func (r *PublishedPageRepository) List(ctx context.Context, limit, offset int) ([]PublishedPageListRow, int, error) {
	const listQ = `
SELECT id, name, slug, theme, created_by, created_at, updated_at
FROM published_pages
ORDER BY updated_at DESC
LIMIT ? OFFSET ?`
	const countQ = `SELECT COUNT(*) FROM published_pages`

	rows, err := r.db.QueryContext(ctx, listQ, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []PublishedPageListRow
	for rows.Next() {
		var row PublishedPageListRow
		if err := rows.Scan(&row.ID, &row.Name, &row.Slug, &row.Theme, &row.CreatedBy, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
