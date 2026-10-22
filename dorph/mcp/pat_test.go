package mcp_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"idongivaflyinfa/auth"
	"idongivaflyinfa/mcp"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func mintPATRow(t *testing.T, exec func(string, ...any) error, userID string, scopes []string, expires time.Time, revoked string) (id, token string) {
	t.Helper()
	id, token, err := auth.NewPATSecret()
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(scopes)
	rev := any(nil)
	if revoked != "" {
		rev = revoked
	}
	if err := exec(`INSERT INTO agent_token (id, user_id, name, token_hash, scopes, created_at, expires_at, revoked_at)
		VALUES (?, ?, 'test', ?, ?, ?, ?, ?)`,
		id, userID, auth.HashToken(token), string(raw),
		time.Now().UTC().Format(time.RFC3339), expires.Format(time.RFC3339), rev); err != nil {
		t.Fatal(err)
	}
	return id, token
}

func TestPATStartsWithoutJWTSecret(t *testing.T) {
	path, writer := seedTwoUsers(t)
	_, tok := mintPATRow(t, func(q string, args ...any) error {
		_, err := writer.Exec(q, args...)
		return err
	}, "ada-id", []string{auth.ScopeNotesRead, auth.ScopeNotesWrite}, time.Now().UTC().Add(24*time.Hour), "")
	t.Setenv("DORPH_ENV", "development")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("TRAN_SQLITE_PATH", path)
	t.Setenv(mcp.TokenEnv, tok)

	id, err := mcp.IdentityFromEnv()
	if err != nil {
		t.Fatalf("IdentityFromEnv: %v", err)
	}
	if id.UserID != "ada-id" || !auth.HasScope(id.Scopes, auth.ScopeNotesWrite) {
		t.Fatalf("identity = %+v", id)
	}
	if id.ExpiresAt == "" {
		t.Fatal("missing expiry")
	}

	ro, err := mcp.OpenReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ro.Close() })
	server, cleanup, err := mcp.NewServer(id, ro, mcp.RecheckWithStore(ro), nil, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	session := connectInMemory(t, server, "2025-06-18")
	res, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: "whoami"})
	if err != nil || res.IsError {
		t.Fatalf("whoami err=%v res=%+v", err, res)
	}
	got := structuredMap(t, res.StructuredContent)
	if got["id"] != "ada-id" {
		t.Fatalf("whoami = %#v", got)
	}
	scopes, _ := got["scopes"].([]any)
	if !containsAny(scopes, auth.ScopeNotesRead) || !containsAny(scopes, auth.ScopeNotesWrite) {
		t.Fatalf("scopes = %#v", got["scopes"])
	}
	if got["expires_at"] == "" {
		t.Fatal("whoami missing expiry")
	}
}

func TestRevokedAndExpiredPATFailNextCall(t *testing.T) {
	path, writer := seedTwoUsers(t)
	id, tok := mintPATRow(t, func(q string, args ...any) error {
		_, err := writer.Exec(q, args...)
		return err
	}, "ada-id", []string{auth.ScopeNotesRead, auth.ScopeNotesWrite}, time.Now().UTC().Add(24*time.Hour), "")
	t.Setenv("DORPH_ENV", "development")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("TRAN_SQLITE_PATH", path)
	t.Setenv(mcp.TokenEnv, tok)

	ident, err := mcp.IdentityFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	ro, err := mcp.OpenReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ro.Close() })
	server, cleanup, err := mcp.NewServer(ident, ro, mcp.RecheckWithStore(ro), nil, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	session := connectInMemory(t, server, "2025-06-18")
	ctx := context.Background()
	if res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "whoami"}); err != nil || res.IsError {
		t.Fatalf("live whoami err=%v res=%+v", err, res)
	}

	if _, err := writer.Exec(`UPDATE agent_token SET revoked_at = ? WHERE id = ?`, time.Now().UTC().Format(time.RFC3339), id); err != nil {
		t.Fatal(err)
	}
	res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "whoami"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("revoked PAT still worked")
	}
	if strings.Contains(toolText(t, res), tok) {
		t.Fatal("revoke error leaked the token")
	}

	_, expired := mintPATRow(t, func(q string, args ...any) error {
		_, err := writer.Exec(q, args...)
		return err
	}, "ada-id", []string{auth.ScopeNotesRead}, time.Now().UTC().Add(-time.Hour), "")
	t.Setenv(mcp.TokenEnv, expired)
	if _, err := mcp.IdentityFromEnv(); err == nil {
		t.Fatal("expired PAT started")
	} else if strings.Contains(err.Error(), expired) {
		t.Fatal("expiry error leaked the token")
	}
}

func TestDeletedUserFailsNextPATCall(t *testing.T) {
	path, writer := seedTwoUsers(t)
	_, tok := mintPATRow(t, func(q string, args ...any) error {
		_, err := writer.Exec(q, args...)
		return err
	}, "ada-id", []string{auth.ScopeNotesRead, auth.ScopeNotesWrite}, time.Now().UTC().Add(24*time.Hour), "")
	t.Setenv("DORPH_ENV", "development")
	t.Setenv("TRAN_SQLITE_PATH", path)
	t.Setenv(mcp.TokenEnv, tok)
	_ = os.Unsetenv("JWT_SECRET")

	ident, err := mcp.IdentityFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	ro, err := mcp.OpenReadOnly(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ro.Close() })
	server, cleanup, err := mcp.NewServer(ident, ro, mcp.RecheckWithStore(ro), nil, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	session := connectInMemory(t, server, "2025-06-18")
	if _, err := writer.Exec(`DELETE FROM plat_users WHERE id = 'ada-id'`); err != nil {
		t.Fatal(err)
	}
	res, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: "whoami"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatal("deleted user still worked")
	}
	if strings.Contains(toolText(t, res), tok) {
		t.Fatal("deleted-user error leaked the token")
	}
}

func containsAny(items []any, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
