# 14 — Master secrets key

Stored secrets, starting with AI provider API keys, are sealed in the Morph API by `morph/internal/secretbox` (import `idongivaflyinfa/internal/secretbox`). A copy of `tran.sqlite` does not include the master key. The master key is not derived from `JWT_SECRET`.

Startup calls `openSecretsBox` after `NewTranSQL`. That passes the bool from `config.ParseMorphEnv` (`config.ValidateStartup` already ran) into `secretbox.Resolve`. Do not pass the master key into `pkg/morphai`; decrypt in the Morph API and pass the provider key as ordinary config.

## Environment

| Variable | Role |
|----------|------|
| `MORPH_SECRETS_KEY` | Current key. 32 random bytes, standard base64. |
| `MORPH_SECRETS_KEY_PREVIOUS` | Optional. Comma-separated older keys. Decrypt only. |

Generate a key:

```bash
openssl rand -base64 32
```

`.env.example` names both variables and does not set a value. `FromEnv` reads the process environment after `config.GetConfig()` has loaded the root `.env`. It does not load dotenv itself.

## Local development

When both variables are unset and the process is not in production mode, create the key file next to the Tran database:

`<directory of TRAN_SQLITE_PATH>/morph-secrets.key`

The default database path is `./data/tran.sqlite` (cwd-relative), so the default file is `./data/morph-secrets.key`. The file mode is `0600`. It holds the standard-base64 key plus a newline. `*.key` and `data/` are gitignored.

The package does not log. Startup must log this warning, and only the path:

```text
warning: generated dev MORPH_SECRETS_KEY file at %s (mode 0600); set MORPH_SECRETS_KEY before production
```

A later dev start reuses the file and does not warn again. The file is not overwritten. A symlink, a non-regular file, a missing parent directory, or any group/other permission bit is refused. An invalid `MORPH_SECRETS_KEY` is refused in dev too; the file is not a fallback.

## Production

`MORPH_ENV` is parsed by `config.ParseMorphEnv()` (`morph/config/startup.go`). Comparison is case-insensitive.

| `MORPH_ENV` | Result |
|-------------|--------|
| `production`, `prod` | production (`true`) |
| unset, `development`, `dev`, `local`, `test` | local (`false`) |
| anything else | error; the process refuses to start |

`secretbox.Resolve` does not read `MORPH_ENV`. `main` calls `openSecretsBox(prod, cfg.TranSQLitePath)` after `NewTranSQL`, which creates the data directory. `Resolve(true, keyPath)` uses only the environment. A missing or invalid `MORPH_SECRETS_KEY` refuses the process. The error names `MORPH_SECRETS_KEY` (or `MORPH_SECRETS_KEY_PREVIOUS` when that value is the problem) and does not print the value. Production does not read `morph-secrets.key`. Before enabling production, copy that file's base64 line into `MORPH_SECRETS_KEY`. A set `MORPH_SECRETS_KEY_PREVIOUS` is loaded for decrypt.

`morph/config` is unchanged. The `main.go` edit is the call after Tran SQLite opens, outside the CORS block:

```go
cfg := config.GetConfig()
production, err := config.ParseMorphEnv()
if err != nil {
    log.Fatalf("refusing to start: %s", err.Error())
}
if err = config.ValidateStartup(cfg, production); err != nil {
    log.Fatalf("refusing to start: %s", err.Error())
}
// NewTranSQL creates filepath.Dir(cfg.TranSQLitePath).
if _, err = openSecretsBox(production, cfg.TranSQLitePath); err != nil {
    log.Fatalf("%s", err.Error())
}
```

## Associated data

Provider-key rows use `provider_key:<scope>:<ownerID>:<provider>`. Scope is `workspace` (ship first) or `user` (later per-user keys). Build it with `ProviderKeyAAD`. Fields must be non-empty and must not contain `:`, because the colons are separators; `owner-1` + `open:ai` would look like a different row. A ciphertext opened with another row's associated data fails.

## Errors

`Open` returns only `ErrUnknownKeyID`, `ErrUnsupportedVersion`, or `ErrDecrypt`. Callers show "key unreadable, re-enter it" for all three, not an internal server error. Error text never includes the master key, the sealed secret, or the plaintext.

## Rotation

1. Generate a new key with `openssl rand -base64 32`. Set it as `MORPH_SECRETS_KEY`. Put the old key, and any older keys still required, in `MORPH_SECRETS_KEY_PREVIOUS` (comma-separated). Restart.
2. For each stored token, if `NeedsRotation` is true, call `Reseal` with that row's associated data and store the result. Count `Reseal` errors on their own. A corrupt token does not report rotation. A tampered previous-key token still reports rotation, and `Reseal` then returns `ErrDecrypt` with no plaintext.
3. When a full scan finds nothing left to rotate, remove `MORPH_SECRETS_KEY_PREVIOUS` and restart.

There is no admin HTTP endpoint for this scan yet. `Reseal` is the function that endpoint will call. New seals always use the current key.

## Token format

`v1:<keyid>:<base64(nonce || ciphertext || tag)>`. AES-256-GCM, 96-bit random nonce. `keyid` is the first 8 lowercase hex characters of SHA-256 over the raw 32-byte key.
