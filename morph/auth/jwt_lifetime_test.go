package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func testSecret() []byte {
	return []byte("0123456789abcdef0123456789abcdef")
}

func TestProductionDecodeRejectsLegacyLongLivedToken(t *testing.T) {
	// Break this catches: production still accepting a token issued for 876000 hours with the same secret.
	secret := testSecret()
	legacy, err := EncodeToken(TokenConfig{Secret: secret, ExpiryHours: 876000}, "user-1", "a@example.com", "a", []string{"Admin"}, "ch")
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("MORPH_ENV", "")
	t.Setenv("JWT_SECRET", string(secret))
	t.Setenv("JWT_EXPIRY_HOURS", "")
	local := LoadTokenConfig()
	claims, err := DecodeToken(local, legacy)
	if err != nil {
		t.Fatalf("local mode rejected legacy token: %v", err)
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatal("tokens issued by EncodeToken must include iat and exp")
	}

	t.Setenv("MORPH_ENV", "production")
	t.Setenv("JWT_EXPIRY_HOURS", "24")
	prod := LoadTokenConfig()
	_, err = DecodeToken(prod, legacy)
	if err == nil {
		t.Fatal("production accepted a token whose lifetime is 876000 hours")
	}
	if strings.Contains(err.Error(), string(secret)) {
		t.Fatalf("error leaked secret: %s", err)
	}
}

func TestProductionDecodeLifetimeBoundary(t *testing.T) {
	// Break this catches: a 169-hour token accepted in production, or a 168-hour token rejected.
	secret := testSecret()
	within, err := EncodeToken(TokenConfig{Secret: secret, ExpiryHours: 168}, "user-1", "a@example.com", "a", nil, "ch")
	if err != nil {
		t.Fatal(err)
	}
	over, err := EncodeToken(TokenConfig{Secret: secret, ExpiryHours: 169}, "user-1", "a@example.com", "a", nil, "ch")
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("MORPH_ENV", "production")
	t.Setenv("JWT_SECRET", string(secret))
	t.Setenv("JWT_EXPIRY_HOURS", "24")
	prod := LoadTokenConfig()
	if _, err := DecodeToken(prod, within); err != nil {
		t.Fatalf("168h token rejected: %v", err)
	}
	if _, err := DecodeToken(prod, over); err == nil {
		t.Fatal("169h token accepted in production")
	}
	issued, err := EncodeToken(prod, "user-1", "a@example.com", "a", nil, "ch")
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeToken(prod, issued)
	if err != nil {
		t.Fatalf("fresh production token rejected: %v", err)
	}
	if got.IssuedAt == nil || got.ExpiresAt == nil {
		t.Fatal("production tokens must include iat and exp")
	}

	t.Setenv("MORPH_ENV", "development")
	local := LoadTokenConfig()
	if _, err := DecodeToken(local, over); err != nil {
		t.Fatalf("local mode rejected 169h token: %v", err)
	}
}

func TestProductionDecodeRejectsMissingIatOrExp(t *testing.T) {
	// Break this catches: production accepting a token that omits iat or exp, so the lifetime cap cannot be checked.
	secret := testSecret()
	now := time.Now()
	noIat, err := signRaw(secret, Claims{
		Email: "a@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	noExp, err := signRaw(secret, Claims{
		Email: "a@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  "user-1",
			IssuedAt: jwt.NewNumericDate(now),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("MORPH_ENV", "development")
	t.Setenv("JWT_SECRET", string(secret))
	t.Setenv("JWT_EXPIRY_HOURS", "24")
	local := LoadTokenConfig()
	if _, err := DecodeToken(local, noIat); err != nil {
		t.Fatalf("local mode rejected token without iat: %v", err)
	}
	if _, err := DecodeToken(local, noExp); err != nil {
		t.Fatalf("local mode rejected token without exp: %v", err)
	}

	t.Setenv("MORPH_ENV", "production")
	prod := LoadTokenConfig()
	if _, err := DecodeToken(prod, noIat); err == nil {
		t.Fatal("production accepted a token with no iat")
	}
	if _, err := DecodeToken(prod, noExp); err == nil {
		t.Fatal("production accepted a token with no exp")
	}
}

func signRaw(secret []byte, claims Claims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}
