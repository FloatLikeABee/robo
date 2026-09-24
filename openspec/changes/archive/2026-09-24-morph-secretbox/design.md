# Design

## Context

See proposal.md for why this exists. Constraints from the repo and the agreed Aegis/Prism note:

- The Go module is `idongivaflyinfa`. The package directory is `morph/internal/secretbox`. Import path: `idongivaflyinfa/internal/secretbox`.
- `morph/config/config.go` loads dotenv inside `GetConfig()` and defaults `JWT_SECRET` to `morph-dev-jwt-secret-change-me`, which is also in `.env.example`. There is no production-mode flag in that file yet (issue #24, another agent). `main` calls `GetConfig()` then opens `TRAN_SQLITE_PATH` (default `./data/tran.sqlite`). Paths are cwd-relative (`docs/agents/13-conventions.md`).
- `.gitignore` already ignores `*.key` and `**/data/`. `deploy/` is absent, so docs must not describe a finished Render/Alibaba runbook. An invented on-disk key still dies on an ephemeral disk, which is why production must not generate one.
- Do not edit `morph/config/config.go` or `morph/handlers/authz_middleware.go`.
- Crypto is `crypto/aes`, `crypto/cipher`, and `crypto/rand` only. No new dependencies.
- The Prism contract (function names on `Box`, sealed format, env names, AAD spelling, the three Open errors) stays as agreed. One real flaw is documented below and constrained in the helper; the wire format does not change.

## Goals / Non-Goals

**Goals:**

- A tested `Box` plus constructors and a reseal helper that #28/#29 can import without a migration.
- A `Resolve(production bool, devKeyPath string)` helper so the #24 follow-up does not re-decide policy.
- Operator docs for generation and rotation.

**Non-Goals:**

- Calling `Resolve` from `main` or `config.GetConfig`.
- An HTTP reseal endpoint, Settings UI, or provider-key columns.
- HSM/KMS, envelope keys, or moving the package into `pkg/` (ComposerX and FormX do not decrypt provider keys).
- Guessing #24's env flag name.

## Decisions

### Keep AES-256-GCM and the v1 token

Agreed token: `v1:<keyid>:<base64(nonce || ciphertext || tag)>`. Nonce is 12 bytes from `cipher.NewGCM` (not a custom nonce size). `keyid` is `hex(SHA-256(rawKey)[:4])`, lowercase. The hash input is the raw 32 bytes after base64 decode, so insignificant whitespace in the env var does not change the id.

Rejected:

- **XChaCha20-Poly1305** (`golang.org/x/crypto`). The issue limits crypto to the standard library, and the agreed algorithm is AES-256-GCM.
- **Per-row data key wrapped by the master key.** A single string is what Prism will store. Envelope encryption can be a later `v2` token. It is not needed to ship workspace provider keys.
- **Trial decryption with no key id.** The contract tells Open to select by key id. Trial decryption would also collapse "unknown key" and "tamper" into one error, and the acceptance test requires those to stay distinct.
- **Derive the master key from `JWT_SECRET`.** Forbidden by the agreed note. The JWT default is committed in `.env.example` and returned by `getEnv` in `config.go`. Deriving from it would seal secrets with a public constant and tie JWT rotation to every provider key.

32-bit key ids collide around 2^16 keys under the birthday bound. A ring holds the current key plus a short previous list, so this is acceptable. `New` fails closed if two *different* keys share an id. The same key repeated in `MORPH_SECRETS_KEY_PREVIOUS` is ignored so a duplicated paste still boots. The 8-hex width is not changed.

### AAD format stays, with a helper that rejects colons

Agreed AAD is `provider_key:<scope>:<ownerID>:<provider>` (`workspace` first, `user` reserved for #32). No provider-key table exists yet, so nothing in `morph/db` picks owner-id or provider character sets. The spelling is ambiguous if any field contains `:`.

That is a real flaw. It is not fixed by changing the string: Prism already agreed the spelling. `ProviderKeyAAD` returns `ErrInvalidAAD` for an empty field, a scope other than `workspace` or `user`, or a field containing `:`. `Seal` / `Open` pass the caller's bytes straight to GCM and do not parse them, so a future non-provider secret can use other AAD. Callers who concatenate the string by hand can still collide; the docs say to use the helper.

Rejected: length-prefixed or JSON AAD. Safer, and it would break the agreed format.

### Fail-closed errors, and what `NeedsRotation` does not know

Exported sentinels:

| Sentinel | When |
| --- | --- |
| `ErrUnknownKeyID` | Well-formed `v1` key id that is not in the ring |
| `ErrUnsupportedVersion` | First field is not `v1`, including an empty string |
| `ErrDecrypt` | Malformed `v1` token, bad base64, short blob, wrong AAD, tamper, GCM failure |
| `ErrKeyMissing` | `MORPH_SECRETS_KEY` unset or blank |
| `ErrInvalidKey` | Bad encoding or length, empty previous slot, distinct key-id collision, previous keys set without a current key |
| `ErrInvalidAAD` | `ProviderKeyAAD` rejected the fields |
| `ErrKeyFile` | Dev key file missing directory, symlink, loose mode, corrupt contents, or a file that changed between stat and open |

Open returns only the first three. Wrappers add static text (`MORPH_SECRETS_KEY must be 32 bytes of standard base64`) and never the value, the plaintext, or a key id taken from caller data that might be the secret itself. Key ids are not secrets, but they are still omitted from errors so logs stay boring.

`NeedsRotation` is a bool, as agreed, so it cannot return a parse error. It is true only when the token is exactly three colon-separated fields, the version is `v1`, the key id matches `^[0-9a-f]{8}$`, and that id is not the current key. Otherwise it is false. It does not decrypt. A tampered previous-key token still reports rotation; `Reseal` then fails with `ErrDecrypt`. Garbage does not look like "needs rotation", because resealing it cannot succeed. The admin scan should count `Reseal` errors separately from `NeedsRotation == false`. Changing the signature to `(bool, error)` was rejected because it breaks the agreed interface.

`Reseal` is a function, not a `Box` method, so the interface Prism imports stays `Seal`, `Open`, and `NeedsRotation`. It opens first. If the token is already current it returns the same string and `rotated=false` (no nonce churn). If the key id is not current it seals with the current key and returns `rotated=true`.

### Constructors and the startup split

```go
type Box interface {
    Seal(plaintext, aad []byte) (string, error)
    Open(sealed string, aad []byte) ([]byte, error)
    NeedsRotation(sealed string) bool
}

func New(current []byte, previous ...[]byte) (Box, error)
func FromEnv() (Box, error)
func LoadOrCreateDevKeyFile(path string) (key []byte, created bool, err error)
func Resolve(production bool, devKeyPath string) (box Box, createdDevKey bool, err error)
func Reseal(box Box, sealed string, aad []byte) (out string, rotated bool, err error)
func ProviderKeyAAD(scope, ownerID, provider string) ([]byte, error)
```

`FromEnv` does not call `repoenv`. `main` already calls `config.GetConfig()`, which loads the root `.env` before anything else. A second loader here would surprise tests and the #24 startup guard.

`Resolve` policy:

1. Production, or `MORPH_SECRETS_KEY` set: `FromEnv` only. Do not create or read the dev file.
2. Not production, both vars unset: `LoadOrCreateDevKeyFile`, then `New`.
3. Not production, current key unset, previous keys set: `ErrInvalidKey`. Do not generate a fresh current key beside leftover previous keys.

The production flag is a `bool` argument. `config.go` has no flag yet, and the issue does not name the env var. Reading a guessed name would race the #24 agent.

Rejected: generating a key inside `FromEnv`. That hides the production/dev split and would mint a key on a disk that does not survive restart.

Key decode strips ASCII whitespace, accepts standard base64 with or without padding, and rejects the URL-safe alphabet. Each previous-list entry is trimmed; an empty entry after trim is `ErrInvalidKey`. Length must be 32 bytes. AES-128 and AES-192 are rejected.

`ScopeWorkspace` (`workspace`) and `ScopeUser` (`user`) are exported constants for `ProviderKeyAAD`.

### Dev key file

Recommended path, computed by the caller after `GetConfig()`: `filepath.Join(filepath.Dir(cfg.TranSQLitePath), "morph-secrets.key")`. With the default sqlite path that is `./data/morph-secrets.key`.

Create with `O_EXCL` and mode `0600`, write standard base64 and a newline, `Sync`, close, then stat again and refuse if group/other bits are set. Do not `MkdirAll`: sqlite startup already owns the data directory, and creating parent directories would hide a bad path. A missing parent is `ErrKeyFile`.

On read: `Lstat`, reject symlinks, open, `Stat`, require `os.SameFile`, reject non-regular files and any mode with group/other bits, decode, require 32 bytes. Do not `chmod` a loose file and continue; the bytes may already have been readable. Do not overwrite an existing or corrupt file. Two creators race on `O_EXCL`; the loser reads the winner's file.

The package does not log. `createdDevKey == true` means the caller logs the path only:

```text
warning: generated dev MORPH_SECRETS_KEY file at %s (mode 0600); set MORPH_SECRETS_KEY before production
```

### Startup wiring for the follow-up (not in this change)

After `cfg := config.GetConfig()` and once #24 exposes its production bool:

```go
dataDir := filepath.Dir(cfg.TranSQLitePath)
keyPath := filepath.Join(dataDir, "morph-secrets.key")
box, created, err := secretbox.Resolve(production, keyPath)
if err != nil {
    log.Fatalf("secrets key: %v", err)
}
if created {
    log.Printf("warning: generated dev MORPH_SECRETS_KEY file at %s (mode 0600); set MORPH_SECRETS_KEY before production", keyPath)
}
```

Hold `box` for #29. Decrypt provider keys inside the Morph API. Pass plaintext into `pkg/morphai` the same way `MORPH_AI_API_KEY` is passed today. Do not pass the master key into `pkg/morphai`.

Production with a missing or invalid key refuses the process and does not read `morph-secrets.key`. Before turning production on, the operator copies that file's base64 line into `MORPH_SECRETS_KEY`. Dev with a missing key creates the 0600 file and warns. An invalid env value refuses in both modes. A later dev start with the file present reuses it and does not warn.

### Rotation procedure

1. Generate a new key with `openssl rand -base64 32`. Set it as `MORPH_SECRETS_KEY`. Put the previous key, and any older ones still required, in `MORPH_SECRETS_KEY_PREVIOUS` as a comma-separated list. Restart.
2. For each stored token, if `NeedsRotation` is true, call `Reseal` with that row's AAD (`ProviderKeyAAD`) and store the result. Rows that fail `Reseal` stay in place and surface as "key unreadable, re-enter it".
3. When a full scan reports no remaining rotations, remove `MORPH_SECRETS_KEY_PREVIOUS` and restart.

No admin endpoint is added here. The helper is what that endpoint will call.

### Design review

Proposer and reviewer pass, against `config.go`, `main.go`, `.gitignore`, and the agreed note:

- Generating the dev key in production fails on an ephemeral disk and looks like success until restart. `Resolve(true, ...)` never creates a file.
- `FromEnv` must not treat "unset" and "invalid" the same, or dev startup would overwrite a typo with a new file. Invalid env never falls back.
- `NeedsRotation == false` on garbage can hide a corrupt row from a naive scan. Accepted because the agreed signature is a bool; the rotation doc tells the admin path to count `Reseal` errors.
- Colon-separated AAD can alias rows. Format unchanged; helper rejects `:`.
- 8-hex key ids can collide. Constructor fails closed. Width unchanged.
- Key id sits outside the GCM tag, so an attacker can edit it. The edit either selects no key (`ErrUnknownKeyID`) or the wrong key (`ErrDecrypt`). It does not forge plaintext. Binding the key id into the internal AAD was rejected as extra behavior the contract does not ask for.
- `config.go` must not grow a default for `MORPH_SECRETS_KEY` the way it did for `JWT_SECRET`. This change does not edit that file, and `.env.example` leaves the variable commented with no value.
- The warning log must not include the key. The package returns a bool; the documented format string prints the path only.

No open question changes the spec, the approach, or the tasks. #24's flag name is deliberately unknown.

## Risks / Trade-offs

- [32-bit key id] → Collision of distinct keys fails `New`. Not a silent wrong-key open.
- [AAD field injection] → Helper rejects `:`. Hand-built AAD can still alias; docs require the helper for provider keys.
- [Dev key file on a shared machine] → Mode `0600` required; group/other bits and symlinks are errors. Dev only.
- [Corrupt partial write] → Existing file is never replaced. Operator deletes a file that was never used to seal.
- [Random 96-bit nonce birthday bound] → Fine for provider-key volume. `Seal` fails if `crypto/rand` cannot fill the nonce.
- [`NeedsRotation` cannot describe a corrupt token] → Documented. `Reseal` still fails closed.
- [Master key lives in process memory] → Same as `JWT_SECRET` today. No HSM in this story.

## Migration Plan

No sealed rows exist. Shipping the package before #28 means the first write is already sealed. Rollback is reverting the package. The dev key file is gitignored. There is no database migration.

## Open Questions

None.
