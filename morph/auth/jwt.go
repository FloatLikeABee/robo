package auth

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"idongivaflyinfa/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims matches the UsersPanel JWT shape so other apps can keep calling /api/auth/user.
type Claims struct {
	Email            string   `json:"email"`
	Username         string   `json:"username"`
	Roles            []string `json:"roles"`
	DefaultChannelID string   `json:"default_channel_id"`
	jwt.RegisteredClaims
}

type TokenConfig struct {
	Secret      []byte
	ExpiryHours int64
	// Production rejects tokens with no iat or exp, and tokens whose
	// exp-iat window is longer than the production maximum (168 hours).
	Production bool
}

func LoadTokenConfig() TokenConfig {
	cfg := config.GetConfig()
	secret := strings.TrimSpace(cfg.JWTSecret)
	if secret == "" {
		secret = config.DefaultJWTSecret
	}
	hours, err := config.JWTExpiryHours()
	if err != nil {
		// Do not sign a 100-year token when production lifetime is invalid.
		// main refuses to start on this error; 24 hours is the safe fallback.
		log.Printf("warning: %s", err.Error())
		hours = config.ProductionJWTExpiryHours
	}
	production, envErr := config.ParseMorphEnv()
	if envErr != nil {
		production = true
	}
	return TokenConfig{Secret: []byte(secret), ExpiryHours: hours, Production: production}
}

func EncodeToken(cfg TokenConfig, userID, email, username string, roles []string, channelID string) (string, error) {
	if len(cfg.Secret) == 0 {
		return "", errors.New("jwt secret missing")
	}
	now := time.Now()
	claims := Claims{
		Email:            email,
		Username:         username,
		Roles:            roles,
		DefaultChannelID: channelID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(cfg.ExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(cfg.Secret)
}

func DecodeToken(cfg TokenConfig, token string) (*Claims, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("empty token")
	}
	opts := []jwt.ParserOption{}
	if cfg.Production {
		// WithIssuedAt rejects iat in the future. One minute covers clock skew
		// and also applies to exp and nbf.
		opts = append(opts, jwt.WithIssuedAt(), jwt.WithLeeway(productionIssuedAtLeeway))
	}
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return cfg.Secret, nil
	}, opts...)
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	if cfg.Production {
		if err := enforceProductionLifetime(claims); err != nil {
			return nil, err
		}
	}
	return claims, nil
}

// productionIssuedAtLeeway is how far ahead of this process an iat may be.
const productionIssuedAtLeeway = time.Minute

// enforceProductionLifetime rejects tokens that cannot prove they were issued
// inside the production ceiling. EncodeToken always sets iat and exp; tokens
// from before that, or issued for the old 876000-hour lifetime, fail here.
func enforceProductionLifetime(claims *Claims) error {
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return errors.New("token is missing iat or exp")
	}
	lifetime := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
	max := time.Duration(config.ProductionMaxJWTExpiryHours) * time.Hour
	if lifetime <= 0 || lifetime > max {
		return fmt.Errorf("token lifetime exceeds the production maximum of %d hours", config.ProductionMaxJWTExpiryHours)
	}
	return nil
}

// FullPermissions is returned for every authenticated user (compat with old clients).
func FullPermissions() []string {
	return []string{
		"morph_util", "morph_booki", "morph_engi", "inbox_message",
		"view_reports", "export_reports",
		"create_form", "broadcast_form", "compose_email",
	}
}
