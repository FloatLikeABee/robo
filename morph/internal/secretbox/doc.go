// Package secretbox seals and opens secrets with an app-level master key.
//
// MORPH_SECRETS_KEY is 32 random bytes, standard base64, from
// `openssl rand -base64 32`. MORPH_SECRETS_KEY_PREVIOUS is an optional
// comma-separated list of decrypt-only keys used during rotation. The master
// key is independent of JWT_SECRET. This package reads the process environment
// only (it does not load dotenv) and it never logs key material or plaintext.
//
// Morph startup calls openSecretsBox after NewTranSQL. config.ParseMorphEnv
// reads MORPH_ENV (production/prod = production; unset/development/dev/local/test
// = local; anything else refuses to start). config.ValidateStartup(cfg, production)
// is the hosting guard in morph/config/startup.go. Resolve does not read
// MORPH_ENV. NewTranSQL creates filepath.Dir(TRAN_SQLITE_PATH) before the call:
//
//	production, err := config.ParseMorphEnv()
//	if err != nil {
//		log.Fatalf("refusing to start: %s", err.Error())
//	}
//	if err = config.ValidateStartup(cfg, production); err != nil {
//		log.Fatalf("refusing to start: %s", err.Error())
//	}
//	// after NewTranSQL(cfg.TranSQLitePath):
//	if _, err = openSecretsBox(production, cfg.TranSQLitePath); err != nil {
//		log.Fatalf("%s", err.Error())
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
