package db

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"
)

const AgentWorkingMemoryCap = 4096

// GetAgentWorkingMemory returns the session summary for an owner, or empty.
func (m *TranSQL) GetAgentWorkingMemory(ctx context.Context, ownerUserID, sessionID string) (string, error) {
	if m == nil || m.DB == nil {
		return "", nil
	}
	ownerUserID = strings.TrimSpace(ownerUserID)
	sessionID = strings.TrimSpace(sessionID)
	if ownerUserID == "" || sessionID == "" {
		return "", nil
	}
	var summary string
	err := m.DB.QueryRowContext(ctx, `
		SELECT summary FROM agent_working_memory WHERE owner_user_id=? AND session_id=?`,
		ownerUserID, sessionID).Scan(&summary)
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(summary), nil
}

// PatchAgentWorkingMemory replaces or merges the session summary. Oversize patches are ignored.
func (m *TranSQL) PatchAgentWorkingMemory(ctx context.Context, ownerUserID, sessionID, patch string) error {
	if m == nil || m.DB == nil {
		return nil
	}
	patch = strings.TrimSpace(patch)
	if patch == "" {
		return nil
	}
	if utf8.RuneCountInString(patch) > AgentWorkingMemoryCap {
		return nil
	}
	ownerUserID = strings.TrimSpace(ownerUserID)
	sessionID = strings.TrimSpace(sessionID)
	if ownerUserID == "" || sessionID == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)
	prev, _ := m.GetAgentWorkingMemory(ctx, ownerUserID, sessionID)
	summary := patch
	if prev != "" && prev != patch {
		summary = strings.TrimSpace(prev + "\n" + patch)
		if utf8.RuneCountInString(summary) > AgentWorkingMemoryCap {
			summary = truncateRunesCap(summary, AgentWorkingMemoryCap)
		}
	}
	_, err := m.DB.ExecContext(ctx, `
		INSERT INTO agent_working_memory (owner_user_id, session_id, summary, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(owner_user_id, session_id) DO UPDATE SET summary=excluded.summary, updated_at=excluded.updated_at`,
		ownerUserID, sessionID, summary, now)
	return err
}

func truncateRunesCap(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}
