package auth

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims matches the UsersPanel JWT shape so other apps can keep calling /api/auth/user.
type Claims struct {
	Email             string   `json:"email"`
	Username          string   `json:"username"`
	Roles             []string `json:"roles"`
	DefaultChannelID  string   `json:"default_channel_id"`
	jwt.RegisteredClaims
}

type TokenConfig struct {
	Secret       []byte
	ExpiryHours  int64
}

func LoadTokenConfig() TokenConfig {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		secret = "morph-dev-jwt-secret-change-me"
	}
	// Default: no practical session timeout (~100 years). Override with JWT_EXPIRY_HOURS.
	hours := int64(100 * 365 * 24)
	if raw := strings.TrimSpace(os.Getenv("JWT_EXPIRY_HOURS")); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
			hours = n
		}
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
