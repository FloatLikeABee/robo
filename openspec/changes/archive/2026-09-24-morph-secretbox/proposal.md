# Proposal

## Why

A copy of Morph's SQLite database must not expose AI provider API keys. Settings (#28) and provider-key storage (#29) need a master key and a seal/open package in place before the first encrypted row is written, so there is no plaintext-then-migrate step. Issue #24 is adding the production-mode flag in parallel; this change supplies the crypto and the startup contract without editing that code.

## What Changes

- Add `morph/internal/secretbox`: AES-256-GCM seal and open, key ids, associated data, typed fail-closed errors, and rotation helpers. Standard-library crypto only.
- Read the master key from `MORPH_SECRETS_KEY` (32 bytes, standard base64) and optional decrypt-only `MORPH_SECRETS_KEY_PREVIOUS`. Never derive it from `JWT_SECRET`.
- Provide constructors for env keys, raw keys, a dev-only 0600 key file, and a `Resolve` helper that production startup can call to refuse a missing key. Do not call it from `morph/config/config.go` or `main` in this change.
- Document the env vars, `openssl rand -base64 32`, dev key file behavior, and the rotation procedure in `.env.example` and the agent docs.
- Leave Settings, provider storage, HSM/KMS, and per-user envelope keys unwired.

## Capabilities

### New Capabilities

- `morph-secretbox`: Seal, open, and rotate app-level secrets (provider API keys first) with an env master key, row-binding associated data, and fail-closed errors that never contain key material.

### Modified Capabilities

- (none)

## Impact

- New Go package `idongivaflyinfa/internal/secretbox` (directory `morph/internal/secretbox`). No new module dependencies.
- `.env.example` and `docs/agents/` only, besides the package and this change. `morph/config/config.go` and `morph/handlers/authz_middleware.go` stay untouched.
- Consumers #28/#29 import the agreed `Box` interface. Startup wiring waits until #24's production flag exists.
