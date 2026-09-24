package main

import (
	"fmt"
	"log"
	"path/filepath"

	"idongivaflyinfa/internal/secretbox"
)

// providerSecrets is the process master-key box. Provider storage opens
// secrets with it. The master key is not passed into pkg/morphai.
var providerSecrets secretbox.Box

// openSecretsBox loads the box after the Tran data directory exists.
// Production, or a set MORPH_SECRETS_KEY, uses the environment only.
// MORPH_SECRETS_KEY_PREVIOUS is loaded when it is set. Local mode with both
// variables unset uses or creates <dir>/morph-secrets.key at mode 0600 and
// logs a path-only warning when the file was created. Errors name the
// variable and do not include the key value.
func openSecretsBox(production bool, tranSQLitePath string) (secretbox.Box, error) {
	keyPath := filepath.Join(filepath.Dir(tranSQLitePath), "morph-secrets.key")
	box, created, err := secretbox.Resolve(production, keyPath)
	if err != nil {
		return nil, fmt.Errorf("refusing to start: %w", err)
	}
	if created {
		log.Printf("warning: generated dev MORPH_SECRETS_KEY file at %s (mode 0600); set MORPH_SECRETS_KEY before production", keyPath)
	}
	providerSecrets = box
	return box, nil
}
