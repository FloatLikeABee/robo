package db

import (
	"database/sql"
	"strings"
	"time"
)

// GetAgentContextCache returns the cached blob when fingerprint matches.
func (t *TranSQL) GetAgentContextCache(userID, sessionID, fingerprint string) (string, bool) {
	if t == nil || t.DB == nil {
		return "", false
	}
	var blob, fp string
	err := t.DB.QueryRow(
		`SELECT fingerprint, blob FROM morph_agent_context_cache WHERE user_id = ? AND session_id = ?`,
		strings.TrimSpace(userID), strings.TrimSpace(sessionID),
	).Scan(&fp, &blob)
	if err == sql.ErrNoRows || err != nil {
		return "", false
	}
	if fp != fingerprint || strings.TrimSpace(blob) == "" {
		return "", false
	}
	return blob, true
}

// PutAgentContextCache stores assembled context for a user/session fingerprint.
func (t *TranSQL) PutAgentContextCache(userID, sessionID, fingerprint, blob string) error {
	if t == nil || t.DB == nil {
		return nil
	}
	_, err := t.DB.Exec(
		`INSERT INTO morph_agent_context_cache (user_id, session_id, fingerprint, blob, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(user_id, session_id) DO UPDATE SET fingerprint = excluded.fingerprint, blob = excluded.blob, updated_at = excluded.updated_at`,
		strings.TrimSpace(userID), strings.TrimSpace(sessionID), fingerprint, blob, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}
