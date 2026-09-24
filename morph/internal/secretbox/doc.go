// Package secretbox seals and opens secrets with an app-level master key.
//
// MORPH_SECRETS_KEY is 32 random bytes, standard base64, from
// `openssl rand -base64 32`. MORPH_SECRETS_KEY_PREVIOUS is an optional
// comma-separated list of decrypt-only keys used during rotation. The master
// key is independent of JWT_SECRET. This package reads the process environment
// only (it does not load dotenv) and it never logs key material or plaintext.
//
// Startup wiring belongs in the Morph process after config.GetConfig and
// config.ParseMorphEnv (MORPH_ENV=production or prod). Call Resolve only after
// the data directory exists; NewTranSQL creates filepath.Dir(TRAN_SQLITE_PATH).
// This package is not called from main:
//
//	dataDir := filepath.Dir(cfg.TranSQLitePath)
//	keyPath := filepath.Join(dataDir, "morph-secrets.key")
//	box, created, err := secretbox.Resolve(production, keyPath)
//	if err != nil {
//		log.Fatalf("secrets key: %v", err)
//	}
//	if created {
//		log.Printf("warning: generated dev MORPH_SECRETS_KEY file at %s (mode 0600); set MORPH_SECRETS_KEY before production", keyPath)
//	}
//
// Production (production == true) refuses a missing or invalid MORPH_SECRETS_KEY
// and does not read morph-secrets.key. Copy that file into MORPH_SECRETS_KEY
// before enabling production. Dev with both env vars unset creates the file
// with mode 0600. Decrypt provider keys in the Morph API and pass plaintext
// into pkg/morphai. Do not pass the master key across that boundary.
//
// Rotation: set the new key as MORPH_SECRETS_KEY and the old key as
// MORPH_SECRETS_KEY_PREVIOUS, restart, Reseal every value for which
// NeedsRotation is true, then remove MORPH_SECRETS_KEY_PREVIOUS.
package secretbox
