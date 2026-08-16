package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type TemplateRepository struct {
	db *sql.DB
}

func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func scanEmailTemplate(rows *sql.Rows) (EmailTemplate, error) {
	var t EmailTemplate
	var builtinKey sql.NullString
	err := rows.Scan(
		&t.ID, &t.Name, &t.Tag, &t.Description, &builtinKey, &t.HTMLContent,
		&t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return t, err
	}
	t.IsBuiltin = builtinKey.Valid && builtinKey.String != ""
	return t, nil
}

func (r *TemplateRepository) List(ctx context.Context, limit, offset int) ([]EmailTemplate, int, error) {
	const listQuery = `
SELECT id, name, tag, description, builtin_key, html_content, created_by, created_at, updated_at
FROM email_templates
ORDER BY (builtin_key IS NULL), name ASC, id ASC
LIMIT ? OFFSET ?`

	const countQuery = `SELECT COUNT(*) FROM email_templates`

	rows, err := r.db.QueryContext(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var templates []EmailTemplate
	for rows.Next() {
		t, err := scanEmailTemplate(rows)
		if err != nil {
			return nil, 0, err
		}
		templates = append(templates, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *TemplateRepository) Get(ctx context.Context, id int64) (*EmailTemplate, error) {
	const q = `
SELECT id, name, tag, description, builtin_key, html_content, created_by, created_at, updated_at
FROM email_templates
WHERE id = ?`
	rows, err := r.db.QueryContext(ctx, q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	t, err := scanEmailTemplate(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepository) Create(ctx context.Context, name, tag, description, htmlContent string, createdBy int64) (int64, error) {
	const insertQuery = `
INSERT INTO email_templates (name, tag, description, html_content, created_by, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	res, err := r.db.ExecContext(ctx, insertQuery, name, tag, description, htmlContent, createdBy)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	enqueueGraphSync(ctx, r.db, "composerx", "email_template", fmt.Sprintf("%d", id), "upsert")
	return id, nil
}

func (r *TemplateRepository) Update(ctx context.Context, id int64, name, tag, description, htmlContent string) error {
	const updateQuery = `
UPDATE email_templates
SET name = ?, tag = ?, description = ?, html_content = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?`

	res, err := r.db.ExecContext(ctx, updateQuery, name, tag, description, htmlContent, id)
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
	enqueueGraphSync(ctx, r.db, "composerx", "email_template", fmt.Sprintf("%d", id), "upsert")
	return nil
}

func enqueueGraphSync(ctx context.Context, db *sql.DB, source, entityType, entityID, op string) {
	if db == nil {
		return
	}
	_, _ = db.ExecContext(ctx, `
		INSERT INTO graph_sync_outbox (source, entity_type, entity_id, op)
		VALUES (?, ?, ?, ?)`, source, entityType, entityID, op)
}

func (r *TemplateRepository) Delete(ctx context.Context, id int64) error {
	const deleteQuery = `DELETE FROM email_templates WHERE id = ?`

	res, err := r.db.ExecContext(ctx, deleteQuery, id)
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

// Simple helper to normalise "not found" semantics
func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

// MergeDataRepository handles merge_data_sources table.
type MergeDataRepository struct {
	db *sql.DB
}

func NewMergeDataRepository(db *sql.DB) *MergeDataRepository {
	return &MergeDataRepository{db: db}
}

func (r *MergeDataRepository) List(ctx context.Context, limit, offset int) ([]MergeDataSource, int, error) {
	const listQuery = `
SELECT id, name, file_type, file_path, uploaded_by, created_at
FROM merge_data_sources
ORDER BY created_at DESC
LIMIT ? OFFSET ?`

	const countQuery = `SELECT COUNT(*) FROM merge_data_sources`

	rows, err := r.db.QueryContext(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []MergeDataSource
	for rows.Next() {
		var m MergeDataSource
		if err := rows.Scan(&m.ID, &m.Name, &m.FileType, &m.FilePath, &m.UploadedBy, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *MergeDataRepository) Create(ctx context.Context, name, fileType, filePath string, uploadedBy int64) (int64, error) {
	const insertQuery = `
INSERT INTO merge_data_sources (name, file_type, file_path, uploaded_by, created_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`

	res, err := r.db.ExecContext(ctx, insertQuery, name, fileType, filePath, uploadedBy)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *MergeDataRepository) GetByID(ctx context.Context, id int64) (*MergeDataSource, error) {
	const selectQuery = `
SELECT id, name, file_type, file_path, uploaded_by, created_at
FROM merge_data_sources
WHERE id = ?`

	var m MergeDataSource
	if err := r.db.QueryRowContext(ctx, selectQuery, id).Scan(
		&m.ID, &m.Name, &m.FileType, &m.FilePath, &m.UploadedBy, &m.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MergeDataRepository) Delete(ctx context.Context, id int64) error {
	const deleteQuery = `DELETE FROM merge_data_sources WHERE id = ?`

	res, err := r.db.ExecContext(ctx, deleteQuery, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
