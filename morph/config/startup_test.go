package config

import (
	"strings"
	"testing"
)

const (
	devJWTSecret     = "morph-dev-jwt-secret-change-me"
	devAdminPassword = "admin123"
	strongJWTSecret  = "0123456789abcdef0123456789abcdef"
	strongAdminPass  = "correct-horse-battery"
)

func strongConfig() Config {
	return Config{JWTSecret: strongJWTSecret, AdminPassword: strongAdminPass}
}

func devConfig() Config {
	return Config{JWTSecret: devJWTSecret, AdminPassword: devAdminPassword}
}

func TestProductionRejectsDevelopmentDefaults(t *testing.T) {
	// Break this catches: production start allowed while the published secret, password, or 876000-hour lifetime is still configured.
	t.Setenv("JWT_EXPIRY_HOURS", "876000")
	err := ValidateStartup(devConfig(), true)
	if err == nil {
		t.Fatal("expected production startup to fail")
	}
	msg := err.Error()
	for _, want := range []string{"JWT_SECRET", "ADMIN_PASSWORD", "JWT_EXPIRY_HOURS", "32", "12", "168"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %s", msg, want)
		}
	}
	if strings.Contains(msg, devJWTSecret) || strings.Contains(msg, devAdminPassword) {
		t.Fatalf("error leaked secret material: %s", msg)
	}
}

func TestProductionAcceptsRealValues(t *testing.T) {
	// Break this catches: a unique secret, unique password, and 24-hour lifetime still refused in production.
	t.Setenv("JWT_EXPIRY_HOURS", "24")
	if err := ValidateStartup(strongConfig(), true); err != nil {
		t.Fatal(err)
	}
}

func TestLocalModeAllowsDevelopmentDefaults(t *testing.T) {
	// Break this catches: local/dev startup refused when the published defaults are still in use.
	t.Setenv("JWT_EXPIRY_HOURS", "876000")
	if err := ValidateStartup(devConfig(), false); err != nil {
		t.Fatal(err)
	}
}

func TestUnrecognizedMorphEnvRefusesStartup(t *testing.T) {
	// Break this catches: an unknown MORPH_ENV such as staging treated as local and allowed to keep defaults.
	const canary = "staging-canary-mode"
	t.Setenv("MORPH_ENV", canary)
	prod, err := ParseMorphEnv()
	if err == nil {
		t.Fatalf("expected error, production=%v", prod)
	}
	msg := err.Error()
	for _, want := range []string{"MORPH_ENV", "production", "development"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %s", msg, want)
		}
	}
	if strings.Contains(msg, canary) || strings.Contains(msg, devJWTSecret) || strings.Contains(msg, devAdminPassword) {
		t.Fatalf("error leaked mode or secret material: %s", msg)
	}
}

func TestParseMorphEnvRecognizedValues(t *testing.T) {
	// Break this catches: production/prod not enforced, or development/dev/local/test/unset treated as production.
	cases := []struct {
		in   string
		prod bool
	}{
		{in: "", prod: false},
		{in: "development", prod: false},
		{in: "Dev", prod: false},
		{in: "local", prod: false},
		{in: "test", prod: false},
		{in: "production", prod: true},
		{in: "PROD", prod: true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			t.Setenv("MORPH_ENV", tc.in)
			prod, err := ParseMorphEnv()
			if err != nil {
				t.Fatal(err)
			}
			if prod != tc.prod {
				t.Fatalf("MORPH_ENV=%q production=%v want %v", tc.in, prod, tc.prod)
			}
		})
	}
}

func TestProductionRejectsWeakJWTSecretsWithoutEcho(t *testing.T) {
	// Break this catches: empty, short, placeholder, or repeated-character JWT secrets accepted, or the value copied into the error.
	t.Setenv("JWT_EXPIRY_HOURS", "24")
	secrets := []string{
		"",
		"   ",
		"short-canary-secret",
		"change-me-change-me-change-me-change-me",
		strings.Repeat("a", 40),
		strings.ToUpper(devJWTSecret),
	}
	for _, secret := range secrets {
		t.Run(secret, func(t *testing.T) {
			cfg := strongConfig()
			cfg.JWTSecret = secret
			err := ValidateStartup(cfg, true)
			if err == nil {
				t.Fatal("expected refusal")
			}
			msg := err.Error()
			if !strings.Contains(msg, "JWT_SECRET") {
				t.Fatalf("error %q missing JWT_SECRET", msg)
			}
			trimmed := strings.TrimSpace(secret)
			if trimmed != "" && strings.Contains(msg, trimmed) {
				t.Fatalf("error leaked secret %q in %s", trimmed, msg)
			}
		})
	}
}

func TestProductionRejectsWeakAdminPasswordWithoutEcho(t *testing.T) {
	// Break this catches: admin123 or a short password accepted in production, or the password copied into the error.
	t.Setenv("JWT_EXPIRY_HOURS", "24")
	passwords := []string{"", "admin123", "Admin123", "short-pw"}
	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			cfg := strongConfig()
			cfg.AdminPassword = password
			err := ValidateStartup(cfg, true)
			if err == nil {
				t.Fatal("expected refusal")
			}
			msg := err.Error()
			if !strings.Contains(msg, "ADMIN_PASSWORD") || !strings.Contains(msg, "12") {
				t.Fatalf("error %q missing password guidance", msg)
			}
			if password != "" && strings.Contains(msg, password) {
				t.Fatalf("error leaked password %q in %s", password, msg)
			}
		})
	}
}

func TestJWTExpiryHours(t *testing.T) {
	// Break this catches: production unset lifetime staying at 876000, a 48-hour setting ignored, or 876000 allowed in production.
	t.Setenv("JWT_EXPIRY_HOURS", "")
	hours, err := JWTExpiryHoursFor(true)
	if err != nil || hours != 24 {
		t.Fatalf("production default hours=%d err=%v", hours, err)
	}
	t.Setenv("JWT_EXPIRY_HOURS", "48")
	hours, err = JWTExpiryHoursFor(true)
	if err != nil || hours != 48 {
		t.Fatalf("configured hours=%d err=%v", hours, err)
	}
	t.Setenv("JWT_EXPIRY_HOURS", "168")
	hours, err = JWTExpiryHoursFor(true)
	if err != nil || hours != 168 {
		t.Fatalf("cap boundary hours=%d err=%v", hours, err)
	}
	t.Setenv("JWT_EXPIRY_HOURS", "876000")
	_, err = JWTExpiryHoursFor(true)
	if err == nil || !strings.Contains(err.Error(), "JWT_EXPIRY_HOURS") || !strings.Contains(err.Error(), "168") {
		t.Fatalf("got %v", err)
	}
	t.Setenv("JWT_EXPIRY_HOURS", "")
	hours, err = JWTExpiryHoursFor(false)
	if err != nil || hours != 876000 {
		t.Fatalf("local default hours=%d err=%v", hours, err)
	}
	t.Setenv("JWT_EXPIRY_HOURS", "876000")
	hours, err = JWTExpiryHoursFor(false)
	if err != nil || hours != 876000 {
		t.Fatalf("local explicit hours=%d err=%v", hours, err)
	}
}

func TestDevelopmentWarningsOmitSecretValues(t *testing.T) {
	// Break this catches: local startup silent about defaults, or the warning printing the secret or password.
	warns := DevelopmentWarnings(devConfig(), false)
	joined := strings.Join(warns, "\n")
	if !strings.Contains(joined, "JWT_SECRET") || !strings.Contains(joined, "ADMIN_PASSWORD") {
		t.Fatalf("warnings %q", joined)
	}
	if strings.Contains(joined, devJWTSecret) || strings.Contains(joined, devAdminPassword) {
		t.Fatalf("warning leaked secret material: %s", joined)
	}
	if got := DevelopmentWarnings(strongConfig(), false); len(got) != 0 {
		t.Fatalf("unexpected warnings for strong local config: %v", got)
	}
	if got := DevelopmentWarnings(devConfig(), true); len(got) != 0 {
		t.Fatalf("production path should not emit dev warnings: %v", got)
	}
}

func TestGetConfigBlankJWTSecretUsesDevelopmentDefault(t *testing.T) {
	// Break this catches: a whitespace JWT_SECRET kept as-is, so production would not see the development default and local signing would use a blank key.
	t.Setenv("JWT_SECRET", "   ")
	cfg := GetConfig()
	if cfg.JWTSecret != devJWTSecret {
		t.Fatalf("JWTSecret=%q", cfg.JWTSecret)
	}
}
