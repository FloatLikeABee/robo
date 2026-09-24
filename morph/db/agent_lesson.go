package db

import (
	"context"
	"database/sql"
	"strings"
)

// AgentLesson is one distilled rule from a significant Morph AI session.
// Lessons are owned by a single platform user (OwnerUserID). Enabled controls
// whether the lesson is injected into that user's prompts and Skills listing.
type AgentLesson struct {
	ID              string `json:"id"`
	Trigger         string `json:"trigger"`
	Rule            string `json:"rule"`
	SourceSessionID string `json:"source_session_id"`
	CreatedAt       string `json:"created_at"`
	Enabled         bool   `json:"enabled"`
	OwnerUserID     string `json:"owner_user_id"`
}

type agentLessonScanner interface {
	Scan(dest ...any) error
}

func scanAgentLesson(sc agentLessonScanner) (AgentLesson, error) {
	var l AgentLesson
	var enabled int
	err := sc.Scan(&l.ID, &l.Trigger, &l.Rule, &l.SourceSessionID, &l.CreatedAt, &enabled, &l.OwnerUserID)
	if err != nil {
		return AgentLesson{}, err
	}
	l.Enabled = enabled != 0
	return l, nil
}

// InsertAgentLesson stores a lesson. Uniqueness is (owner_user_id, source_session_id).
func (m *TranSQL) InsertAgentLesson(ctx context.Context, l *AgentLesson) error {
	if m == nil || m.DB == nil || l == nil {
		return sql.ErrConnDone
	}
	enabled := 0
	if l.Enabled {
		enabled = 1
	}
	_, err := m.DB.ExecContext(ctx, `
		INSERT INTO agent_lesson (id, trigger, rule, source_session_id, created_at, enabled, owner_user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		l.ID, l.Trigger, l.Rule, l.SourceSessionID, l.CreatedAt, enabled, strings.TrimSpace(l.OwnerUserID))
	return err
}

// GetAgentLessonBySession returns the owner's lesson for a session, or nil.
// An empty owner matches nothing, so unowned legacy rows are not reused.
func (m *TranSQL) GetAgentLessonBySession(ctx context.Context, ownerUserID, sessionID string) (*AgentLesson, error) {
	if m == nil || m.DB == nil {
		return nil, nil
	}
	ownerUserID = strings.TrimSpace(ownerUserID)
	sessionID = strings.TrimSpace(sessionID)
	if ownerUserID == "" || sessionID == "" {
		return nil, nil
	}
	row := m.DB.QueryRowContext(ctx, `
		SELECT id, trigger, rule, source_session_id, created_at, enabled, owner_user_id
		FROM agent_lesson WHERE owner_user_id=? AND source_session_id=?`, ownerUserID, sessionID)
	l, err := scanAgentLesson(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// GetAgentLessonForOwner returns the lesson when id belongs to ownerUserID.
// A missing id and another user's id both return (nil, nil).
func (m *TranSQL) GetAgentLessonForOwner(ctx context.Context, ownerUserID, id string) (*AgentLesson, error) {
	if m == nil || m.DB == nil {
		return nil, nil
	}
	ownerUserID = strings.TrimSpace(ownerUserID)
	id = strings.TrimSpace(id)
	if ownerUserID == "" || id == "" {
		return nil, nil
	}
	row := m.DB.QueryRowContext(ctx, `
		SELECT id, trigger, rule, source_session_id, created_at, enabled, owner_user_id
		FROM agent_lesson WHERE id=? AND owner_user_id=?`, id, ownerUserID)
	l, err := scanAgentLesson(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// ListAgentLessons returns one user's lessons, newest first.
// enabledOnly keeps rows with enabled=1. limit <= 0 returns the full set (capped at 500).
// An empty owner returns no rows.
func (m *TranSQL) ListAgentLessons(ctx context.Context, ownerUserID string, enabledOnly bool, limit int) ([]AgentLesson, error) {
	if m == nil || m.DB == nil {
		return nil, nil
	}
	ownerUserID = strings.TrimSpace(ownerUserID)
	if ownerUserID == "" {
		return []AgentLesson{}, nil
	}
	if limit < 0 {
		limit = 0
	}
	if limit > 500 {
		limit = 500
	}
	q := `
		SELECT id, trigger, rule, source_session_id, created_at, enabled, owner_user_id
		FROM agent_lesson
		WHERE owner_user_id=?`
	args := []any{ownerUserID}
	if enabledOnly {
		q += ` AND enabled=1`
	}
	q += ` ORDER BY created_at DESC, id DESC`
	if limit > 0 {
		q += ` LIMIT ?`
		args = append(args, limit)
	} else {
		q += ` LIMIT 500`
	}
	rows, err := m.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AgentLesson, 0)
	for rows.Next() {
		l, err := scanAgentLesson(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// SetAgentLessonEnabled toggles enabled for a lesson the owner owns.
// Missing and other users' lessons return sql.ErrNoRows.
func (m *TranSQL) SetAgentLessonEnabled(ctx context.Context, ownerUserID, id string, enabled bool) (*AgentLesson, error) {
	if m == nil || m.DB == nil {
		return nil, sql.ErrConnDone
	}
	ownerUserID = strings.TrimSpace(ownerUserID)
	id = strings.TrimSpace(id)
	if ownerUserID == "" || id == "" {
		return nil, sql.ErrNoRows
	}
	flag := 0
	if enabled {
		flag = 1
	}
	res, err := m.DB.ExecContext(ctx, `
		UPDATE agent_lesson SET enabled=? WHERE id=? AND owner_user_id=?`,
		flag, id, ownerUserID)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, sql.ErrNoRows
	}
	return m.GetAgentLessonForOwner(ctx, ownerUserID, id)
}

// DeleteAgentLessonForOwner removes a lesson the owner owns.
// Missing and other users' lessons return sql.ErrNoRows.
func (m *TranSQL) DeleteAgentLessonForOwner(ctx context.Context, ownerUserID, id string) error {
	if m == nil || m.DB == nil {
		return sql.ErrConnDone
	}
	ownerUserID = strings.TrimSpace(ownerUserID)
	id = strings.TrimSpace(id)
	if ownerUserID == "" || id == "" {
		return sql.ErrNoRows
	}
	res, err := m.DB.ExecContext(ctx, `
		DELETE FROM agent_lesson WHERE id=? AND owner_user_id=?`, id, ownerUserID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
