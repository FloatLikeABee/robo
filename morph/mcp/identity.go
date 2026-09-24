// Package mcp is the Model Context Protocol server for Morph.
//
// The stdio process does not open Badger. morph-api holds an exclusive
// Badger directory lock (DB_PATH and ENTITY_DETAILS_BADGER). SQLite is
// opened read-only (mode=ro and query_only) so a WAL writer in morph-api
// can keep the file. This package must not call db.New,
// db.NewBadgerEntityDetails, or db.NewTranSQL: the last one runs schema writes.
//
// Identity is a Morph session JWT from the environment, checked with
// auth.DecodeToken. A raw user id is not accepted.
package mcp

import (
	"errors"
	"os"
	"strings"

	"idongivaflyinfa/auth"
	"idongivaflyinfa/config"
)

// TokenEnv is the environment variable that carries a Morph session JWT.
// The value is a credential: never log it.
const TokenEnv = "MORPH_MCP_TOKEN"

// Identity is the Morph user a server acts as. It is taken from verified
// JWT claims, not from a database lookup.
type Identity struct {
	UserID   string
	Email    string
	Username string
	Roles    []string
}

// IdentityFromEnv reads TokenEnv and verifies it with auth.LoadTokenConfig
// and auth.DecodeToken (signature, exp, and the production lifetime cap).
// JWT_SECRET must be the same value the Morph API used to sign the token.
// An empty secret or the built-in development default is refused before
// LoadTokenConfig can substitute that default. In production the secret
// strength and JWT_EXPIRY_HOURS rules from config.ValidateStartup apply too.
func IdentityFromEnv() (Identity, error) {
	if err := validateStartupAuth(); err != nil {
		return Identity{}, err
	}
	return ResolveIdentity(os.Getenv(TokenEnv), auth.LoadTokenConfig())
}

// validateStartupAuth applies the same JWT gate as morph-api main before
// DecodeToken runs. Admin password rules stay on the API: this process
// never reads ADMIN_PASSWORD.
func validateStartupAuth() error {
	prod, err := config.ParseMorphEnv()
	if err != nil {
		return err
	}
	secret := os.Getenv("JWT_SECRET")
	if prod {
		if msg := config.JWTSecretProblem(secret); msg != "" {
			return errors.New(msg)
		}
		if _, err := config.JWTExpiryHoursFor(true); err != nil {
			return err
		}
		return nil
	}
	return RejectJWTSecret(secret)
}

// RejectJWTSecret fails closed when the secret is empty or the value
// auth.LoadTokenConfig would use if JWT_SECRET were unset. The error does
// not include the secret.
func RejectJWTSecret(secret string) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if strings.EqualFold(secret, config.DefaultJWTSecret) {
		return errors.New("JWT_SECRET is still the development default")
	}
	return nil
}

// RecheckFromEnv verifies the process token again. Tool handlers call it
// so an expired token becomes a tool error without restarting.
func RecheckFromEnv() error {
	return RecheckToken(os.Getenv(TokenEnv), auth.LoadTokenConfig())
}

// RecheckToken verifies the session JWT again for one tool call.
// A failure does not include the token text.
func RecheckToken(token string, cfg auth.TokenConfig) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("MORPH_MCP_TOKEN is required")
	}
	if len(cfg.Secret) == 0 {
		return errors.New("JWT_SECRET is required")
	}
	if _, err := auth.DecodeToken(cfg, token); err != nil {
		return errors.New("MORPH_MCP_TOKEN is expired or invalid")
	}
	return nil
}

// ResolveIdentity verifies a Morph session JWT. Missing and invalid tokens
// return errors that do not include the token text.
func ResolveIdentity(token string, cfg auth.TokenConfig) (Identity, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return Identity{}, errors.New("MORPH_MCP_TOKEN is required")
	}
	if len(cfg.Secret) == 0 {
		return Identity{}, errors.New("JWT_SECRET is required")
	}
	claims, err := auth.DecodeToken(cfg, token)
	if err != nil {
		return Identity{}, errors.New("MORPH_MCP_TOKEN is invalid")
	}
	userID := strings.TrimSpace(claims.Subject)
	if userID == "" {
		return Identity{}, errors.New("MORPH_MCP_TOKEN is missing a subject")
	}
	roles := claims.Roles
	if roles == nil {
		roles = []string{}
	} else {
		roles = append([]string(nil), roles...)
	}
	return Identity{
		UserID:   userID,
		Email:    claims.Email,
		Username: claims.Username,
		Roles:    roles,
	}, nil
}
