package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"idongivaflyinfa/auth"
)

// AgentToken is one personal access token row. Token is set only on create.
type AgentToken struct {
	ID         string
	UserID     string
	Name       string
	Token      string
	Scopes     []string
	CreatedAt  string
	ExpiresAt  string
	LastUsedAt string
	RevokedAt  string
}

func (m *TranSQL) CreateAgentToken(ctx context.Context, userID, name string, scopes []string, days int) (*AgentToken, error) {
	if m == nil || m.DB == nil {
		return nil, errors.New("sql store not configured")
	}
	userID = strings.TrimSpace(userID)
	name = strings.Join(strings.Fields(name), " ")
	if userID == "" || name == "" {
		return nil, errors.New("name is required")
	}
	if len(name) > 80 {
		return nil, errors.New("name is too long")
	}
	scopes, err := auth.NormalizeScopes(scopes)
	if err != nil {
		return nil, err
	}
	id, plaintext, err := auth.NewPATSecret()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	row := AgentToken{
		ID:        id,
		UserID:    userID,
		Name:      name,
		Token:     plaintext,
		Scopes:    scopes,
		CreatedAt: now.Format(time.RFC3339),
		ExpiresAt: auth.PATExpiry(days).Format(time.RFC3339),
	}
	scopeJSON, _ := json.Marshal(scopes)
	_, err = m.DB.ExecContext(ctx, `
		INSERT INTO agent_token (id, user_id, name, token_hash, scopes, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.UserID, row.Name, auth.HashToken(plaintext), string(scopeJSON), row.CreatedAt, row.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *TranSQL) ListAgentTokens(ctx context.Context, userID string) ([]AgentToken, error) {
	if m == nil || m.DB == nil {
		return nil, errors.New("sql store not configured")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("user is required")
	}
	rows, err := m.DB.QueryContext(ctx, `
		SELECT id, user_id, name, scopes, created_at, expires_at,
			COALESCE(last_used_at, ''), COALESCE(revoked_at, '')
		FROM agent_token WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AgentToken{}
	for rows.Next() {
		t, err := scanAgentToken(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (m *TranSQL) RevokeAgentToken(ctx context.Context, userID, id string) error {
	if m == nil || m.DB == nil {
		return errors.New("sql store not configured")
	}
	userID = strings.TrimSpace(userID)
	id = strings.TrimSpace(id)
	if userID == "" || id == "" {
		return sql.ErrNoRows
	}
	res, err := m.DB.ExecContext(ctx, `
		UPDATE agent_token SET revoked_at = ? WHERE id = ? AND user_id = ? AND revoked_at IS NULL`,
		time.Now().UTC().Format(time.RFC3339), id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (m *TranSQL) LookupAgentToken(ctx context.Context, token string) (*AgentToken, error) {
	if m == nil || m.DB == nil {
		return nil, errors.New("sql store not configured")
	}
	token = strings.TrimSpace(token)
	if !auth.IsPAT(token) {
		return nil, errors.New("not a personal access token")
	}
	id := patID(token)
	if id == "" {
		return nil, errors.New("invalid token")
	}
	row := m.DB.QueryRowContext(ctx, `
		SELECT id, user_id, name, token_hash, scopes, created_at, expires_at,
			COALESCE(last_used_at, ''), COALESCE(revoked_at, '')
		FROM agent_token WHERE id = ?`, id)
	var t AgentToken
	var hash, scopeJSON string
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &hash, &scopeJSON, &t.CreatedAt, &t.ExpiresAt, &t.LastUsedAt, &t.RevokedAt)
	if err != nil {
		return nil, err
	}
	if !auth.EqualHash(hash, auth.HashToken(token)) {
		return nil, errors.New("invalid token")
	}
	t.Scopes = parseScopesJSON(scopeJSON)
	return &t, nil
}

func (m *TranSQL) TouchAgentToken(ctx context.Context, id string) {
	if m == nil || m.DB == nil || strings.TrimSpace(id) == "" {
		return
	}
	_, _ = m.DB.ExecContext(ctx, `UPDATE agent_token SET last_used_at = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), id)
}

func (t AgentToken) Live() error {
	if strings.TrimSpace(t.RevokedAt) != "" {
		return errors.New("token is revoked")
	}
	if exp, err := time.Parse(time.RFC3339, strings.TrimSpace(t.ExpiresAt)); err == nil && time.Now().UTC().After(exp) {
		return errors.New("token is expired")
	}
	return nil
}

func (t AgentToken) Public() map[string]any {
	return map[string]any{
		"id":           t.ID,
		"name":         t.Name,
		"scopes":       t.Scopes,
		"created_at":   t.CreatedAt,
		"expires_at":   t.ExpiresAt,
		"last_used_at": emptyToNil(t.LastUsedAt),
		"revoked_at":   emptyToNil(t.RevokedAt),
	}
}

func scanAgentToken(s interface{ Scan(dest ...any) error }) (AgentToken, error) {
	var t AgentToken
	var scopeJSON string
	err := s.Scan(&t.ID, &t.UserID, &t.Name, &scopeJSON, &t.CreatedAt, &t.ExpiresAt, &t.LastUsedAt, &t.RevokedAt)
	if err != nil {
		return t, err
	}
	t.Scopes = parseScopesJSON(scopeJSON)
	return t, nil
}

func parseScopesJSON(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var scopes []string
	if err := json.Unmarshal([]byte(raw), &scopes); err != nil {
		return nil
	}
	return scopes
}

func patID(token string) string {
	rest := auth.StripPATPrefix(token)
	id, _, ok := strings.Cut(rest, "_")
	if !ok || id == "" {
		return ""
	}
	return id
}

func emptyToNil(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
