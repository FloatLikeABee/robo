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
	t.Setenv("MORPH_ENV", "development")
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

func TestIdentityFromEnvProductionRules(t *testing.T) {
	secret := "0123456789abcdef0123456789abcdef"
	t.Setenv("MORPH_ENV", "production")
	t.Setenv("JWT_EXPIRY_HOURS", "24")

	for _, bad := range []string{
		"stdio-test-secret",
		strings.Repeat("a", 32),
		"change-me-but-long-enough-for-rules",
	} {
		t.Setenv("JWT_SECRET", bad)
		t.Setenv(mcp.TokenEnv, "ignored")
		_, err := mcp.IdentityFromEnv()
		if err == nil {
			t.Fatalf("production accepted %q", bad)
		}
		if !strings.Contains(err.Error(), "JWT_SECRET") || strings.Contains(err.Error(), bad) {
			t.Fatalf("secret %q error = %q", bad, err)
		}
	}

	t.Setenv("JWT_SECRET", secret)
	t.Setenv("JWT_EXPIRY_HOURS", "876000")
	_, err := mcp.IdentityFromEnv()
	if err == nil || !strings.Contains(err.Error(), "JWT_EXPIRY_HOURS") {
		t.Fatalf("expiry error = %v", err)
	}

	t.Setenv("JWT_EXPIRY_HOURS", "24")
	longLived := signToken(t, auth.TokenConfig{Secret: []byte(secret), ExpiryHours: 876000}, "user-1", "a@example.com", "a", nil)
	t.Setenv(mcp.TokenEnv, longLived)
	_, err = mcp.IdentityFromEnv()
	if err == nil {
		t.Fatal("production accepted a 876000-hour token")
	}
	if strings.Contains(err.Error(), longLived) {
		t.Fatal("error included the token")
	}

	t.Setenv("MORPH_ENV", "staging")
	_, err = mcp.IdentityFromEnv()
	if err == nil || !strings.Contains(err.Error(), "MORPH_ENV") || strings.Contains(err.Error(), "staging") {
		t.Fatalf("env error = %v", err)
	}

	t.Setenv("MORPH_ENV", "production")
	good := signToken(t, auth.TokenConfig{Secret: []byte(secret), ExpiryHours: 24}, "user-1", "a@example.com", "a", nil)
	t.Setenv(mcp.TokenEnv, good)
	id, err := mcp.IdentityFromEnv()
	if err != nil {
		t.Fatalf("production token: %v", err)
	}
	if id.UserID != "user-1" {
		t.Fatalf("identity = %+v", id)
	}

	t.Setenv("MORPH_ENV", "development")
	if err := mcp.RecheckFromEnv(); err != nil {
		t.Fatalf("development recheck of production-signed token: %v", err)
	}
	t.Setenv(mcp.TokenEnv, longLived)
	if err := mcp.RecheckFromEnv(); err != nil {
		t.Fatalf("development rejected long-lived token: %v", err)
	}
	t.Setenv("MORPH_ENV", "production")
	if err := mcp.RecheckFromEnv(); err == nil {
		t.Fatal("production recheck accepted a long-lived token")
	} else if strings.Contains(err.Error(), longLived) {
		t.Fatal("recheck error included the token")
	}
}

func TestIdentityFromEnvMissing(t *testing.T) {
	t.Setenv("MORPH_ENV", "development")
	t.Setenv("JWT_SECRET", "stdio-test-secret")
	t.Setenv(mcp.TokenEnv, "  ")
	_, err := mcp.IdentityFromEnv()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Fatalf("error = %q", err)
	}
}
