package main

import (
	"context"
	"database/sql"
	"errors"
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

func (r *PublishedPageRepository) ResolveUniqueSlug(ctx context.Context, name string) (string, error) {
	_ = ctx
	return slugifyPublishName(name), nil
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed") ||
		strings.Contains(msg, "duplicate entry") ||
		strings.Contains(msg, "error 1062")
}

type publishedPageRow struct {
	ID        int64
	Name      string
	Slug      string
	Theme     string
	MongoID   string
	CreatedBy int64
}

func (r *PublishedPageRepository) lookupBySlug(ctx context.Context, slug string) (*publishedPageRow, error) {
	const q = `
SELECT id, name, slug, theme, content_mongo_id, created_by
FROM published_pages
WHERE slug = ?
LIMIT 1`
	var row publishedPageRow
	err := r.db.QueryRowContext(ctx, q, strings.TrimSpace(slug)).Scan(
		&row.ID, &row.Name, &row.Slug, &row.Theme, &row.MongoID, &row.CreatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PublishedPageRepository) updateExisting(ctx context.Context, row *publishedPageRow, name, theme, html string) (*PublishedPage, error) {
	if err := r.content.UpdateHTML(ctx, row.MongoID, html, ""); err != nil {
		return nil, err
	}
	_, err := r.db.ExecContext(ctx, `
UPDATE published_pages
SET name = ?, theme = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?`, name, theme, row.ID)
	if err != nil {
		return nil, err
	}
	return &PublishedPage{
		ID:        row.ID,
		Name:      name,
		Slug:      row.Slug,
		Theme:     theme,
		CreatedBy: row.CreatedBy,
	}, nil
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

	slug := slugifyPublishName(name)
	existing, err := r.lookupBySlug(ctx, slug)
	if err == nil {
		return r.updateExisting(ctx, existing, name, theme, html)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	mongoID, err := r.content.InsertHTML(ctx, html, "")
	if err != nil {
		return nil, err
	}

	const ins = `
INSERT INTO published_pages (name, slug, theme, content_mongo_id, created_by, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	res, err := r.db.ExecContext(ctx, ins, name, slug, theme, mongoID, createdBy)
	if err != nil {
		_ = r.content.DeleteByHexID(ctx, mongoID)
		if isUniqueConstraint(err) {
			row, lookupErr := r.lookupBySlug(ctx, slug)
			if lookupErr != nil {
				return nil, err
			}
			return r.updateExisting(ctx, row, name, theme, html)
		}
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		_ = r.content.DeleteByHexID(ctx, mongoID)
		return nil, err
	}
	return &PublishedPage{
		ID:        id,
		Name:      name,
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
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&row.ID, &row.Name, &row.Slug, &row.Theme, &row.CreatedBy, scanDestTime{&createdAt}, scanDestTime{&updatedAt}); err != nil {
			return nil, 0, err
		}
		if createdAt.Valid {
			row.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			row.UpdatedAt = updatedAt.Time
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

func (r *PublishedPageRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return sql.ErrNoRows
	}
	const q = `SELECT content_mongo_id FROM published_pages WHERE id = ?`
	var mongoID string
	if err := r.db.QueryRowContext(ctx, q, id).Scan(&mongoID); err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM published_pages WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	_ = r.content.DeleteByHexID(ctx, mongoID)
	return nil
}
