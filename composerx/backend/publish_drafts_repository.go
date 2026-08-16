package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type PublishDraft struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Theme       string    `json:"theme"`
	HTMLContent string    `json:"html_content"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PublishDraftListRow struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Theme     string    `json:"theme"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PublishDraftRepository struct {
	db *sql.DB
}

func NewPublishDraftRepository(db *sql.DB) *PublishDraftRepository {
	return &PublishDraftRepository{db: db}
}

func (r *PublishDraftRepository) Create(ctx context.Context, name, theme, html string, createdBy int64) (int64, error) {
	name = strings.TrimSpace(name)
	theme = strings.TrimSpace(theme)
	if name == "" || strings.TrimSpace(html) == "" {
		return 0, errors.New("name and html required")
	}
	if createdBy <= 0 {
		createdBy = 1
	}
	if theme == "" {
		theme = "default"
	}
	const q = `
INSERT INTO publish_drafts (name, theme, html_content, created_by, created_at, updated_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	res, err := r.db.ExecContext(ctx, q, name, theme, html, createdBy)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *PublishDraftRepository) List(ctx context.Context, limit, offset int) ([]PublishDraftListRow, int, error) {
	const listQ = `
SELECT id, name, theme, created_by, created_at, updated_at
FROM publish_drafts
ORDER BY updated_at DESC
LIMIT ? OFFSET ?`
	const countQ = `SELECT COUNT(*) FROM publish_drafts`

	rows, err := r.db.QueryContext(ctx, listQ, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []PublishDraftListRow
	for rows.Next() {
		var row PublishDraftListRow
		if err := rows.Scan(&row.ID, &row.Name, &row.Theme, &row.CreatedBy, &row.CreatedAt, &row.UpdatedAt); err != nil {
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

func (r *PublishDraftRepository) GetByID(ctx context.Context, id int64) (*PublishDraft, error) {
	const q = `
SELECT id, name, theme, html_content, created_by, created_at, updated_at
FROM publish_drafts
WHERE id = ?`
	var row PublishDraft
	if err := r.db.QueryRowContext(ctx, q, id).Scan(
		&row.ID,
		&row.Name,
		&row.Theme,
		&row.HTMLContent,
		&row.CreatedBy,
		&row.CreatedAt,
		&row.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PublishDraftRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM publish_drafts WHERE id = ?`, id)
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
	return nil
}
