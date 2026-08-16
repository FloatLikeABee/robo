package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
)

type SavedEmailRepository struct {
	db      *sql.DB
	content *EmailContentStore
}

func NewSavedEmailRepository(db *sql.DB, content *EmailContentStore) *SavedEmailRepository {
	return &SavedEmailRepository{db: db, content: content}
}

// SavedEmailListRow is SQL index metadata only (no body).
type SavedEmailListRow struct {
	ID        int64
	Name      string
	UpdatedAt sql.NullTime
}

func (r *SavedEmailRepository) List(ctx context.Context, limit, offset int) ([]SavedEmailListRow, int, error) {
	const listQuery = `
SELECT id, name, updated_at
FROM saved_emails
ORDER BY updated_at DESC
LIMIT ? OFFSET ?`
	const countQuery = `SELECT COUNT(*) FROM saved_emails`

	rows, err := r.db.QueryContext(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []SavedEmailListRow
	for rows.Next() {
		var row SavedEmailListRow
		if err := rows.Scan(&row.ID, &row.Name, &row.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *SavedEmailRepository) getMongoID(ctx context.Context, id int64) (string, error) {
	var mongoID string
	err := r.db.QueryRowContext(ctx, `SELECT content_mongo_id FROM saved_emails WHERE id = ?`, id).Scan(&mongoID)
	if err != nil {
		return "", err
	}
	return mongoID, nil
}

// GetDetail returns metadata and markdown from Mongo.
func (r *SavedEmailRepository) GetDetail(ctx context.Context, id int64) (*SavedEmailDetail, error) {
	const q = `
SELECT id, name, content_mongo_id, created_by, created_at, updated_at
FROM saved_emails WHERE id = ?`
	var d SavedEmailDetail
	var mongoID string
	var createdAt, updatedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&d.ID, &d.Name, &mongoID, &d.CreatedBy, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	if createdAt.Valid {
		d.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		d.UpdatedAt = updatedAt.Time
	}

	markdown, aiJSON, err := r.content.GetMarkdown(ctx, mongoID)
	if err != nil {
		return nil, err
	}
	d.MarkdownContent = markdown
	aiJSON = strings.TrimSpace(aiJSON)
	if aiJSON != "" && aiJSON != "null" && json.Valid([]byte(aiJSON)) {
		d.ComposerAISession = json.RawMessage(aiJSON)
	}
	return &d, nil
}

// Create inserts Mongo body then SQL row. Rolls back Mongo if SQL fails.
func (r *SavedEmailRepository) Create(ctx context.Context, name string, markdown string, composerAISessionJSON string, createdBy int64) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" || strings.TrimSpace(markdown) == "" {
		return 0, errors.New("name and markdown required")
	}

	mongoID, err := r.content.InsertMarkdown(ctx, markdown, composerAISessionJSON)
	if err != nil {
		return 0, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		_ = r.content.DeleteByHexID(ctx, mongoID)
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			_ = r.content.DeleteByHexID(ctx, mongoID)
		}
	}()

	const ins = `
INSERT INTO saved_emails (name, content_mongo_id, created_by, created_at, updated_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	res, err := tx.ExecContext(ctx, ins, name, mongoID, createdBy)
	if err != nil {
		return 0, err
	}
	savedID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return savedID, nil
}

// Update replaces markdown in Mongo and metadata in SQL.
func (r *SavedEmailRepository) Update(ctx context.Context, id int64, name string, markdown string, composerAISessionJSON string) error {
	name = strings.TrimSpace(name)
	if name == "" || strings.TrimSpace(markdown) == "" {
		return errors.New("name and markdown required")
	}

	mongoID, err := r.getMongoID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.content.UpdateMarkdown(ctx, mongoID, markdown, composerAISessionJSON); err != nil {
		return err
	}

	const upd = `
UPDATE saved_emails
SET name = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?`
	res, err := r.db.ExecContext(ctx, upd, name, id)
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

// Delete removes SQL row then Mongo document.
func (r *SavedEmailRepository) Delete(ctx context.Context, id int64) error {
	mongoID, err := r.getMongoID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return err
	}

	res, err := r.db.ExecContext(ctx, `DELETE FROM saved_emails WHERE id = ?`, id)
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
