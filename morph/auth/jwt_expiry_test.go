package auth

import "testing"

func TestLoadTokenConfigExpiryFollowsMode(t *testing.T) {
	// Break this catches: token signing still using the 876000-hour development lifetime after production mode is on.
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("MORPH_ENV", "production")
	t.Setenv("JWT_EXPIRY_HOURS", "")
	if got := LoadTokenConfig().ExpiryHours; got != 24 {
		t.Fatalf("production default expiry = %d, want 24", got)
	}

	t.Setenv("JWT_EXPIRY_HOURS", "48")
	if got := LoadTokenConfig().ExpiryHours; got != 48 {
		t.Fatalf("production configured expiry = %d, want 48", got)
	}

	t.Setenv("JWT_EXPIRY_HOURS", "876000")
	if got := LoadTokenConfig().ExpiryHours; got == 876000 {
		t.Fatal("production token lifetime stayed at 876000 hours")
	}
	if got := LoadTokenConfig().ExpiryHours; got != 24 {
		t.Fatalf("production invalid expiry fallback = %d, want 24", got)
	}

	t.Setenv("MORPH_ENV", "")
	t.Setenv("JWT_EXPIRY_HOURS", "")
	if got := LoadTokenConfig().ExpiryHours; got != 876000 {
		t.Fatalf("local default expiry = %d, want 876000", got)
	}
}
