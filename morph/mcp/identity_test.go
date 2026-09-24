package mcp_test

import (
	"strings"
	"testing"

	"idongivaflyinfa/auth"
	"idongivaflyinfa/mcp"
)

func testTokenConfig() auth.TokenConfig {
	return auth.TokenConfig{Secret: []byte("test-jwt-secret"), ExpiryHours: 24}
}

func signToken(t *testing.T, cfg auth.TokenConfig, userID, email, username string, roles []string) string {
	t.Helper()
	tok, err := auth.EncodeToken(cfg, userID, email, username, roles, "ch_test")
	if err != nil {
		t.Fatalf("EncodeToken: %v", err)
	}
	return tok
}

func TestResolveIdentity(t *testing.T) {
	cfg := testTokenConfig()
	tok := signToken(t, cfg, "user-1", "ada@example.com", "ada", []string{"Admin"})

	id, err := mcp.ResolveIdentity(tok, cfg)
	if err != nil {
		t.Fatalf("ResolveIdentity: %v", err)
	}
	if id.UserID != "user-1" || id.Email != "ada@example.com" || id.Username != "ada" {
		t.Fatalf("identity = %+v", id)
	}
	if len(id.Roles) != 1 || id.Roles[0] != "Admin" {
		t.Fatalf("roles = %#v", id.Roles)
	}
}

func TestResolveIdentityNormalizesNilRoles(t *testing.T) {
	cfg := testTokenConfig()
	tok := signToken(t, cfg, "user-2", "bea@example.com", "bea", nil)

	id, err := mcp.ResolveIdentity(tok, cfg)
	if err != nil {
		t.Fatalf("ResolveIdentity: %v", err)
	}
	if id.Roles == nil {
		t.Fatal("roles should be an empty slice, not nil")
	}
	if len(id.Roles) != 0 {
		t.Fatalf("roles = %#v", id.Roles)
	}
}

func TestResolveIdentityRejectsMissingToken(t *testing.T) {
	_, err := mcp.ResolveIdentity("   ", testTokenConfig())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "MORPH_MCP_TOKEN") || !strings.Contains(err.Error(), "required") {
		t.Fatalf("error = %q", err)
	}
}

func TestResolveIdentityRejectsInvalidTokenWithoutLeakingIt(t *testing.T) {
	cfg := testTokenConfig()
	tok := signToken(t, cfg, "user-1", "ada@example.com", "ada", []string{"Admin"})
	other := auth.TokenConfig{Secret: []byte("other-secret"), ExpiryHours: 24}

	_, err := mcp.ResolveIdentity(tok, other)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), tok) {
		t.Fatal("error included the token")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("error = %q", err)
	}
}

func TestResolveIdentityRejectsExpiredToken(t *testing.T) {
	cfg := auth.TokenConfig{Secret: []byte("test-jwt-secret"), ExpiryHours: -1}
	tok := signToken(t, cfg, "user-1", "ada@example.com", "ada", nil)

	_, err := mcp.ResolveIdentity(tok, auth.TokenConfig{Secret: cfg.Secret, ExpiryHours: 24})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), tok) {
		t.Fatal("error included the token")
	}
}

func TestResolveIdentityRejectsMissingSubject(t *testing.T) {
	cfg := testTokenConfig()
	tok := signToken(t, cfg, "  ", "ada@example.com", "ada", nil)

	_, err := mcp.ResolveIdentity(tok, cfg)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "subject") {
		t.Fatalf("error = %q", err)
	}
	if strings.Contains(err.Error(), tok) {
		t.Fatal("error included the token")
	}
}

func TestResolveIdentityRejectsEmptySecret(t *testing.T) {
	cfg := testTokenConfig()
	tok := signToken(t, cfg, "user-1", "ada@example.com", "ada", nil)

	_, err := mcp.ResolveIdentity(tok, auth.TokenConfig{})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), tok) {
		t.Fatal("error included the token")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("error = %q", err)
	}
}

func TestIdentityFromEnv(t *testing.T) {
	cfg := testTokenConfig()
	tok := signToken(t, cfg, "user-9", "cy@example.com", "cy", []string{"employee"})
	t.Setenv("JWT_SECRET", string(cfg.Secret))
	t.Setenv(mcp.TokenEnv, tok)

	id, err := mcp.IdentityFromEnv()
	if err != nil {
		t.Fatalf("IdentityFromEnv: %v", err)
	}
	if id.UserID != "user-9" || id.Username != "cy" {
		t.Fatalf("identity = %+v", id)
	}
}

func TestIdentityFromEnvMissing(t *testing.T) {
	t.Setenv(mcp.TokenEnv, "  ")
	_, err := mcp.IdentityFromEnv()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Fatalf("error = %q", err)
	}
}
