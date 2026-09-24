package auth

import (
	"errors"
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
	return TokenConfig{Secret: []byte(secret), ExpiryHours: hours}
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
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return cfg.Secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func IsAdminRoles(roles []string) bool {
	for _, r := range roles {
		if strings.EqualFold(strings.TrimSpace(r), "Admin") {
			return true
		}
	}
	return false
}

// FullPermissions is returned for every authenticated user (compat with old clients).
func FullPermissions() []string {
	return []string{
		"morph_util", "morph_booki", "morph_engi", "inbox_message",
		"view_reports", "export_reports",
		"create_form", "broadcast_form", "compose_email",
	}
}
