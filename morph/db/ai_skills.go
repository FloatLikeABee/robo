package db

import (
	"context"
	"database/sql"
	"strings"
)

// AISkill is the SQLite index row for a Morph AI skill.
type AISkill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	OwnerUserID string `json:"owner_user_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CountAISkills returns total skills in the index.
func (m *TranSQL) CountAISkills(ctx context.Context) (int64, error) {
	if m == nil || m.DB == nil {
		return 0, nil
	}
	var n int64
	err := m.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_skills`).Scan(&n)
	return n, err
}

// InsertAISkill inserts a skill index row.
func (m *TranSQL) InsertAISkill(ctx context.Context, s *AISkill) error {
	if m == nil || m.DB == nil || s == nil {
		return sql.ErrConnDone
	}
	enabled := 0
	if s.Enabled {
		enabled = 1
	}
	_, err := m.DB.ExecContext(ctx, `
		INSERT INTO ai_skills (id, name, description, enabled, owner_user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Name, s.Description, enabled, s.OwnerUserID, s.CreatedAt, s.UpdatedAt)
	return err
}

// UpdateAISkill updates mutable skill index fields.
func (m *TranSQL) UpdateAISkill(ctx context.Context, s *AISkill) error {
	if m == nil || m.DB == nil || s == nil {
		return sql.ErrConnDone
	}
	enabled := 0
	if s.Enabled {
		enabled = 1
	}
	res, err := m.DB.ExecContext(ctx, `
		UPDATE ai_skills SET name=?, description=?, enabled=?, updated_at=? WHERE id=?`,
		s.Name, s.Description, enabled, s.UpdatedAt, s.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetAISkill returns one skill by id, or nil if missing.
func (m *TranSQL) GetAISkill(ctx context.Context, id string) (*AISkill, error) {
	if m == nil || m.DB == nil {
		return nil, nil
	}
	id = strings.TrimSpace(id)
	var s AISkill
	var enabled int
	err := m.DB.QueryRowContext(ctx, `
		SELECT id, name, description, enabled, owner_user_id, created_at, updated_at
		FROM ai_skills WHERE id=?`, id).
		Scan(&s.ID, &s.Name, &s.Description, &enabled, &s.OwnerUserID, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.Enabled = enabled != 0
	return &s, nil
}

// ListAISkills lists skills; when enabledOnly is true, only enabled rows.
func (m *TranSQL) ListAISkills(ctx context.Context, enabledOnly bool) ([]AISkill, error) {
	if m == nil || m.DB == nil {
		return nil, nil
	}
	q := `
		SELECT id, name, description, enabled, owner_user_id, created_at, updated_at
		FROM ai_skills`
	if enabledOnly {
		q += ` WHERE enabled=1`
	}
	q += ` ORDER BY name COLLATE NOCASE ASC`
	rows, err := m.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AISkill
	for rows.Next() {
		var s AISkill
		var enabled int
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &enabled, &s.OwnerUserID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Enabled = enabled != 0
		out = append(out, s)
	}
	return out, rows.Err()
}

// DeleteAISkill removes a skill index row.
func (m *TranSQL) DeleteAISkill(ctx context.Context, id string) error {
	if m == nil || m.DB == nil {
		return nil
	}
	res, err := m.DB.ExecContext(ctx, `DELETE FROM ai_skills WHERE id=?`, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
