// Package mcp is the Model Context Protocol server for Morph.
//
// The stdio process does not open Badger or SQLite. morph-api holds an
// exclusive Badger directory lock (DB_PATH and ENTITY_DETAILS_BADGER); a
// second process that opens those directories fails and can disrupt the API.
// SQLite is opened by the API in WAL mode, so a later story can read
// TRAN_SQLITE_PATH with mode=ro and query_only without taking that lock.
// This package must not call db.New, db.NewBadgerEntityDetails, or
// db.NewTranSQL: the last one runs schema writes.
//
// Identity is a Morph session JWT from the environment, checked with
// auth.DecodeToken. A raw user id is not accepted.
package mcp

import (
	"errors"
	"os"
	"strings"

	"idongivaflyinfa/auth"
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

// IdentityFromEnv reads TokenEnv and verifies it with auth.LoadTokenConfig.
// JWT_SECRET must be the same value the Morph API used to sign the token.
func IdentityFromEnv() (Identity, error) {
	return ResolveIdentity(os.Getenv(TokenEnv), auth.LoadTokenConfig())
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
