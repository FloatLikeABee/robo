package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"idongivaflyinfa/auth"
)

func ResolvePAT(ctx context.Context, db *sql.DB, token string) (Identity, error) {
	token = strings.TrimSpace(token)
	if !auth.IsPAT(token) {
		return Identity{}, errors.New("DORPH_MCP_TOKEN is invalid")
	}
	if db == nil {
		return Identity{}, errors.New("note store is not open")
	}
	id := patID(token)
	if id == "" {
		return Identity{}, errors.New("DORPH_MCP_TOKEN is invalid")
	}
	var userID, hash, scopeJSON, expires, revoked string
	err := db.QueryRowContext(ctx, `
		SELECT user_id, token_hash, scopes, expires_at, COALESCE(revoked_at, '')
		FROM agent_token WHERE id = ?`, id).Scan(&userID, &hash, &scopeJSON, &expires, &revoked)
	if err != nil {
		return Identity{}, errors.New("DORPH_MCP_TOKEN is invalid")
	}
	if !auth.EqualHash(hash, auth.HashToken(token)) {
		return Identity{}, errors.New("DORPH_MCP_TOKEN is invalid")
	}
	if strings.TrimSpace(revoked) != "" {
		return Identity{}, errors.New("DORPH_MCP_TOKEN is revoked")
	}
	if exp, err := time.Parse(time.RFC3339, strings.TrimSpace(expires)); err == nil && time.Now().UTC().After(exp) {
		return Identity{}, errors.New("DORPH_MCP_TOKEN is expired or invalid")
	}
	idnt, err := platIdentity(ctx, db, userID, parseTokenScopes(scopeJSON), expires)
	if err != nil {
		return Identity{}, err
	}
	_, _ = db.ExecContext(ctx, `UPDATE agent_token SET last_used_at = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), id)
	return idnt, nil
}

func platIdentity(ctx context.Context, db *sql.DB, userID string, scopes []string, expires string) (Identity, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return Identity{}, errors.New("DORPH_MCP_TOKEN subject is not a Dorph user")
	}
	var email, username, rolesJSON string
	err := db.QueryRowContext(ctx, `SELECT email, username, roles FROM plat_users WHERE id = ?`, userID).
		Scan(&email, &username, &rolesJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return Identity{}, errors.New("DORPH_MCP_TOKEN subject is not a Dorph user")
	}
	if err != nil {
		return Identity{}, errors.New("DORPH_MCP_TOKEN is invalid")
	}
	var roles []string
	_ = json.Unmarshal([]byte(rolesJSON), &roles)
	if roles == nil {
		roles = []string{}
	}
	if scopes == nil {
		scopes = []string{}
	}
	return Identity{UserID: userID, Email: email, Username: username, Roles: roles, Scopes: scopes, ExpiresAt: expires}, nil
}

func parseTokenScopes(raw string) []string {
	var scopes []string
	if json.Unmarshal([]byte(strings.TrimSpace(raw)), &scopes) != nil {
		return nil
	}
	return scopes
}

func patID(token string) string {
	rest := auth.StripPATPrefix(token)
	id, _, ok := strings.Cut(rest, "_")
	if !ok {
		return ""
	}
	return id
}
