package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

const (
	PATPrefix         = "dorph_pat_"
	legacyPATPrefix   = "morph_pat_"
	DefaultPATDays    = 90
	ScopeNotesRead    = "notes:read"
	ScopeNotesWrite   = "notes:write"
	ScopePoolRead     = "pool:read"
	ScopePoolPublish  = "pool:publish"
	ScopeAdminPool    = "admin:pool"
)

var defaultPATScopes = []string{ScopeNotesRead, ScopeNotesWrite, ScopePoolRead}

var allowedScopes = map[string]bool{
	ScopeNotesRead:   true,
	ScopeNotesWrite:  true,
	ScopePoolRead:    true,
	ScopePoolPublish: true,
	ScopeAdminPool:   true,
}

func IsPAT(token string) bool {
	t := strings.TrimSpace(token)
	return strings.HasPrefix(t, PATPrefix) || strings.HasPrefix(t, legacyPATPrefix)
}

func StripPATPrefix(token string) string {
	t := strings.TrimSpace(token)
	t = strings.TrimPrefix(t, PATPrefix)
	return strings.TrimPrefix(t, legacyPATPrefix)
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func EqualHash(stored, got string) bool {
	a := []byte(strings.TrimSpace(stored))
	b := []byte(strings.TrimSpace(got))
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}

func DefaultPATScopes() []string {
	return append([]string(nil), defaultPATScopes...)
}

func NormalizeScopes(in []string) ([]string, error) {
	if len(in) == 0 {
		return DefaultPATScopes(), nil
	}
	seen := map[string]bool{}
	var out []string
	for _, raw := range in {
		s := strings.ToLower(strings.TrimSpace(raw))
		if s == "" {
			continue
		}
		if !allowedScopes[s] {
			return nil, errors.New("invalid scope")
		}
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	if len(out) == 0 {
		return DefaultPATScopes(), nil
	}
	return out, nil
}

func HasScope(scopes []string, want string) bool {
	if want == "" {
		return true
	}
	if len(scopes) == 0 {
		return want == ScopeNotesRead || want == ScopeNotesWrite
	}
	for _, s := range scopes {
		if s == want {
			return true
		}
	}
	return false
}

func NewPATSecret() (id, token string, err error) {
	var raw [40]byte
	if _, err = rand.Read(raw[:]); err != nil {
		return "", "", err
	}
	id = hex.EncodeToString(raw[:8])
	secret := hex.EncodeToString(raw[8:])
	return id, PATPrefix + id + "_" + secret, nil
}

func PATExpiry(days int) time.Time {
	if days <= 0 {
		days = DefaultPATDays
	}
	if days > 365 {
		days = 365
	}
	return time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
}
