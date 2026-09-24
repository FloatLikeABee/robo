package db

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func openGuardDB(t *testing.T) *TranSQL {
	t.Helper()
	m, err := NewTranSQL(filepath.Join(t.TempDir(), "tran.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m
}

func TestStoredDefaultAdminPasswordIsReported(t *testing.T) {
	// Break this catches: a seeded admin123 hash ignored because only the env password was checked.
	ctx := context.Background()
	m := openGuardDB(t)
	if err := m.EnsureBootstrapAdmin(ctx, "morphadmin@local.com", "morphadmin", "admin123"); err != nil {
		t.Fatal(err)
	}
	hit, err := m.HasStoredDefaultAdminPassword(ctx, "ops@example.com", "ops")
	if err != nil || !hit {
		t.Fatalf("hit=%v err=%v", hit, err)
	}
}

func TestGuardRefusesStoredDefaultPasswordWithoutEcho(t *testing.T) {
	// Break this catches: startup continuing while SQLite still accepts the development admin password, or the error printing that password.
	ctx := context.Background()
	m := openGuardDB(t)
	if err := m.EnsureBootstrapAdmin(ctx, "morphadmin@local.com", "morphadmin", "admin123"); err != nil {
		t.Fatal(err)
	}
	before, err := m.GetPlatUserByUsername(ctx, "morphadmin")
	if err != nil {
		t.Fatal(err)
	}
	const replacement = "correct-horse-battery"
	n, err := m.GuardStoredDefaultAdminPassword(ctx, "morphadmin@local.com", "morphadmin", false, replacement)
	if err == nil {
		t.Fatal("expected refusal")
	}
	if n != 0 {
		t.Fatalf("rotated %d accounts on refusal", n)
	}
	msg := err.Error()
	for _, secret := range []string{"admin123", replacement, "MORPH_ROTATE_DEFAULT_ADMIN"} {
		if secret != "MORPH_ROTATE_DEFAULT_ADMIN" && strings.Contains(msg, secret) {
			t.Fatalf("error leaked %q in %s", secret, msg)
		}
	}
	if !strings.Contains(msg, "MORPH_ROTATE_DEFAULT_ADMIN") || !strings.Contains(msg, "ADMIN_PASSWORD") {
		t.Fatalf("error %q missing rotation guidance", msg)
	}
	after, err := m.GetPlatUserByUsername(ctx, "morphadmin")
	if err != nil {
		t.Fatal(err)
	}
	if after.ID != before.ID || !VerifyPassword(after.PasswordHash, "admin123") {
		t.Fatal("password or id changed without rotation")
	}
}

func TestGuardRotatesStoredDefaultPasswordsAndKeepsID(t *testing.T) {
	// Break this catches: rotation minting a new user id, or leaving a second admin on the development password.
	ctx := context.Background()
	m := openGuardDB(t)
	if err := m.EnsureBootstrapAdmin(ctx, "morphadmin@local.com", "morphadmin", "admin123"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.CreatePlatUser(ctx, "other-admin@example.com", "admin123", true); err != nil {
		t.Fatal(err)
	}
	before, err := m.GetPlatUserByUsername(ctx, "morphadmin")
	if err != nil {
		t.Fatal(err)
	}
	const replacement = "correct-horse-battery"
	n, err := m.GuardStoredDefaultAdminPassword(ctx, "morphadmin@local.com", "morphadmin", true, replacement)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("rotated %d, want 2", n)
	}
	after, err := m.GetPlatUserByUsername(ctx, "morphadmin")
	if err != nil {
		t.Fatal(err)
	}
	if after.ID != before.ID {
		t.Fatalf("id changed %s -> %s", before.ID, after.ID)
	}
	if !VerifyPassword(after.PasswordHash, replacement) || VerifyPassword(after.PasswordHash, "admin123") {
		t.Fatal("bootstrap password was not replaced")
	}
	other, err := m.GetPlatUserByEmail(ctx, "other-admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(other.PasswordHash, replacement) || VerifyPassword(other.PasswordHash, "admin123") {
		t.Fatal("second admin password was not replaced")
	}
	n, err = m.GuardStoredDefaultAdminPassword(ctx, "morphadmin@local.com", "morphadmin", false, replacement)
	if err != nil || n != 0 {
		t.Fatalf("after rotate n=%d err=%v", n, err)
	}
}

func TestGuardIgnoresNonAdminWithDevelopmentPassword(t *testing.T) {
	// Break this catches: a non-admin who chose the development password blocking production or having that password overwritten.
	ctx := context.Background()
	m := openGuardDB(t)
	if _, err := m.CreatePlatUser(ctx, "bob@example.com", "admin123", false); err != nil {
		t.Fatal(err)
	}
	hit, err := m.HasStoredDefaultAdminPassword(ctx, "ops@example.com", "ops")
	if err != nil || hit {
		t.Fatalf("hit=%v err=%v", hit, err)
	}
	n, err := m.GuardStoredDefaultAdminPassword(ctx, "ops@example.com", "ops", true, "correct-horse-battery")
	if err != nil || n != 0 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	bob, err := m.GetPlatUserByEmail(ctx, "bob@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(bob.PasswordHash, "admin123") {
		t.Fatal("non-admin password changed")
	}
}

func TestGuardEmptyDatabasePasses(t *testing.T) {
	// Break this catches: a fresh database refused before the bootstrap admin can be created.
	ctx := context.Background()
	m := openGuardDB(t)
	if err := m.EnsurePlatUsersTable(ctx); err != nil {
		t.Fatal(err)
	}
	n, err := m.GuardStoredDefaultAdminPassword(ctx, "morphadmin@local.com", "morphadmin", false, "correct-horse-battery")
	if err != nil || n != 0 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

func TestGuardRotateRejectsWeakPasswordWithoutWrite(t *testing.T) {
	// Break this catches: MORPH_ROTATE_DEFAULT_ADMIN replacing hashes with the development password.
	ctx := context.Background()
	m := openGuardDB(t)
	if err := m.EnsureBootstrapAdmin(ctx, "morphadmin@local.com", "morphadmin", "admin123"); err != nil {
		t.Fatal(err)
	}
	n, err := m.GuardStoredDefaultAdminPassword(ctx, "morphadmin@local.com", "morphadmin", true, "admin123")
	if err == nil {
		t.Fatal("expected refusal")
	}
	if n != 0 {
		t.Fatalf("rotated %d", n)
	}
	if strings.Contains(err.Error(), "admin123") {
		t.Fatalf("error leaked password: %s", err)
	}
	u, err := m.GetPlatUserByUsername(ctx, "morphadmin")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(u.PasswordHash, "admin123") {
		t.Fatal("hash changed")
	}
}
