package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// DefaultJWTSecret is the published local signing key. Production refuses it.
	DefaultJWTSecret = "morph-dev-jwt-secret-change-me"
	// DefaultAdminPassword is the published local bootstrap password. Production refuses it.
	DefaultAdminPassword = "admin123"
	// DefaultAdminUsername is the published local bootstrap username.
	DefaultAdminUsername = "morphadmin"
	// DefaultAdminEmail is the published local bootstrap email.
	DefaultAdminEmail = "morphadmin@local.com"

	MinJWTSecretLength                = 32
	MinAdminPasswordLength            = 12
	DevJWTExpiryHours           int64 = 876000
	ProductionJWTExpiryHours    int64 = 24
	ProductionMaxJWTExpiryHours int64 = 168
)

// ParseMorphEnv reports whether this process is in production mode.
// Unrecognized MORPH_ENV values return an error and do not echo the raw value.
func ParseMorphEnv() (bool, error) {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("MORPH_ENV"))) {
	case "", "development", "dev", "local", "test":
		return false, nil
	case "production", "prod":
		return true, nil
	default:
		return false, fmt.Errorf("MORPH_ENV is not recognized. Use production for hosting or development for local runs (also accepted: prod, dev, local, test)")
	}
}

// ValidateStartup refuses production boot when the JWT secret, admin password,
// or JWT lifetime is not safe to host. Local/dev mode always returns nil.
// Error text names the variables to set and does not include secret values.
func ValidateStartup(cfg Config, production bool) error {
	if !production {
		return nil
	}
	var problems []string
	if msg := JWTSecretProblem(cfg.JWTSecret); msg != "" {
		problems = append(problems, msg)
	}
	if msg := adminPasswordProblem(cfg.AdminPassword); msg != "" {
		problems = append(problems, msg)
	}
	if _, err := JWTExpiryHoursFor(true); err != nil {
		problems = append(problems, err.Error())
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%s", strings.Join(problems, "; "))
}

func resolvedJWTSecret() string {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return DefaultJWTSecret
	}
	return secret
}

// JWTSecretProblem reports why a JWT secret is unsafe to host.
// An empty string means the secret passes. The text does not include the secret.
func JWTSecretProblem(secret string) string {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return fmt.Sprintf("JWT_SECRET is empty. Set JWT_SECRET to a unique random string of at least %d characters", MinJWTSecretLength)
	}
	if strings.EqualFold(secret, DefaultJWTSecret) {
		return fmt.Sprintf("JWT_SECRET is still the development default. Set JWT_SECRET to a unique random string of at least %d characters", MinJWTSecretLength)
	}
	lower := strings.ToLower(secret)
	for _, frag := range []string{"change-me", "changeme", "change_me", "replace-me", "replace_me"} {
		if strings.Contains(lower, frag) {
			return fmt.Sprintf("JWT_SECRET is a known placeholder. Set JWT_SECRET to a unique random string of at least %d characters", MinJWTSecretLength)
		}
	}
	if singleRepeatedCharacter(secret) {
		return fmt.Sprintf("JWT_SECRET is a single repeated character. Set JWT_SECRET to a unique random string of at least %d characters", MinJWTSecretLength)
	}
	if len(secret) < MinJWTSecretLength {
		return fmt.Sprintf("JWT_SECRET is shorter than %d characters. Set JWT_SECRET to a unique random string of at least %d characters", MinJWTSecretLength, MinJWTSecretLength)
	}
	return ""
}

func singleRepeatedCharacter(secret string) bool {
	var first rune
	for i, r := range secret {
		if i == 0 {
			first = r
			continue
		}
		if r != first {
			return false
		}
	}
	return secret != ""
}

func adminPasswordProblem(password string) string {
	password = strings.TrimSpace(password)
	if password == "" || strings.EqualFold(password, DefaultAdminPassword) {
		return fmt.Sprintf("ADMIN_PASSWORD is missing or still the development default. Set ADMIN_PASSWORD to a unique password of at least %d characters", MinAdminPasswordLength)
	}
	if len(password) < MinAdminPasswordLength {
		return fmt.Sprintf("ADMIN_PASSWORD is shorter than %d characters. Set ADMIN_PASSWORD to a unique password of at least %d characters", MinAdminPasswordLength, MinAdminPasswordLength)
	}
	return ""
}

// JWTExpiryHours resolves the process lifetime from MORPH_ENV and JWT_EXPIRY_HOURS.
func JWTExpiryHours() (int64, error) {
	production, err := ParseMorphEnv()
	if err != nil {
		return 0, err
	}
	return JWTExpiryHoursFor(production)
}

// JWTExpiryHoursFor resolves JWT lifetime for an already-known mode.
// Local/dev keeps 876000 hours when unset or invalid. Production defaults to 24
// and rejects values outside 1..168.
func JWTExpiryHoursFor(production bool) (int64, error) {
	raw := strings.TrimSpace(os.Getenv("JWT_EXPIRY_HOURS"))
	if raw == "" {
		if production {
			return ProductionJWTExpiryHours, nil
		}
		return DevJWTExpiryHours, nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 1 {
		if production {
			return 0, fmt.Errorf("JWT_EXPIRY_HOURS must be a whole number of hours from 1 to %d (production default is %d when unset)", ProductionMaxJWTExpiryHours, ProductionJWTExpiryHours)
		}
		return DevJWTExpiryHours, nil
	}
	if production && n > ProductionMaxJWTExpiryHours {
		return 0, fmt.Errorf("JWT_EXPIRY_HOURS is %d, which is longer than the production maximum of %d hours. Set JWT_EXPIRY_HOURS to a value from 1 to %d (for example %d)", n, ProductionMaxJWTExpiryHours, ProductionMaxJWTExpiryHours, ProductionJWTExpiryHours)
	}
	return n, nil
}

// DevelopmentWarnings lists local-only secret warnings. It returns nothing in production.
// Messages name variables and do not include secret values.
func DevelopmentWarnings(cfg Config, production bool) []string {
	if production {
		return nil
	}
	var out []string
	if JWTSecretProblem(cfg.JWTSecret) != "" {
		out = append(out, "JWT_SECRET is a development default, placeholder, or shorter than 32 characters. Morph will still start because MORPH_ENV is not production. Set a unique secret before hosting. See docs/security-hosting-checklist.md")
	}
	if adminPasswordProblem(cfg.AdminPassword) != "" {
		out = append(out, "ADMIN_PASSWORD is the development default or shorter than 12 characters. Morph will still start because MORPH_ENV is not production. Set a unique password before hosting. See docs/security-hosting-checklist.md")
	}
	return out
}

// RotateDefaultAdmin reports whether this production start should replace
// stored development admin passwords. The flag is ignored by callers outside production.
func RotateDefaultAdmin() bool {
	return envTruthy("MORPH_ROTATE_DEFAULT_ADMIN")
}
