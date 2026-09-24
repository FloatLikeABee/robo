package db

import (
	"context"
	"database/sql"
	"strings"
)

// AgentLesson is one distilled rule from a significant Morph AI session.
type AgentLesson struct {
	ID              string `json:"id"`
	Trigger         string `json:"trigger"`
	Rule            string `json:"rule"`
	SourceSessionID string `json:"source_session_id"`
	CreatedAt       string `json:"created_at"`
}

// InsertAgentLesson stores a lesson. Unique source_session_id prevents duplicates.
func (m *TranSQL) InsertAgentLesson(ctx context.Context, l *AgentLesson) error {
	if m == nil || m.DB == nil || l == nil {
		return sql.ErrConnDone
	}
	_, err := m.DB.ExecContext(ctx, `
		INSERT INTO agent_lesson (id, trigger, rule, source_session_id, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		l.ID, l.Trigger, l.Rule, l.SourceSessionID, l.CreatedAt)
	return err
}

// GetAgentLessonBySession returns the lesson for a session, or nil.
func (m *TranSQL) GetAgentLessonBySession(ctx context.Context, sessionID string) (*AgentLesson, error) {
	if m == nil || m.DB == nil {
		return nil, nil
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, nil
	}
	var l AgentLesson
	err := m.DB.QueryRowContext(ctx, `
		SELECT id, trigger, rule, source_session_id, created_at
		FROM agent_lesson WHERE source_session_id=?`, sessionID).
		Scan(&l.ID, &l.Trigger, &l.Rule, &l.SourceSessionID, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// ListRecentAgentLessons returns newest lessons first.
func (m *TranSQL) ListRecentAgentLessons(ctx context.Context, limit int) ([]AgentLesson, error) {
	if m == nil || m.DB == nil {
		return nil, nil
	}
	if limit < 1 || limit > 50 {
		limit = 8
	}
	rows, err := m.DB.QueryContext(ctx, `
		SELECT id, trigger, rule, source_session_id, created_at
		FROM agent_lesson
		ORDER BY created_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AgentLesson
	for rows.Next() {
		var l AgentLesson
		if err := rows.Scan(&l.ID, &l.Trigger, &l.Rule, &l.SourceSessionID, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
