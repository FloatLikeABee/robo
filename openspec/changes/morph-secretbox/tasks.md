# Tasks

## 1. Seal and open

- [x] 1.1 Add `morph/internal/secretbox` tests (external test package) for round trip, distinct ciphertexts, empty plaintext, wrong associated data, tampered and truncated payloads, unknown key id versus decrypt failure, unsupported version, and error text that does not contain plaintext or key material. Use obviously fake 32-byte keys such as repeated `0x11` bytes. Verify `go test` in that package fails because the API is missing, not because of a typo in the test.
- [x] 1.2 Implement `Box`, `New`, `Seal`, and `Open` with AES-256-GCM, the `v1` token, SHA-256 key ids, and sentinels `ErrUnknownKeyID`, `ErrUnsupportedVersion`, and `ErrDecrypt`. Verify the tests from 1.1 pass and `go test` prints no plaintext or key bytes.

## 2. Environment keys and rotation

- [x] 2.1 Add failing tests for `FromEnv`: missing key, invalid base64, wrong length, URL-safe alphabet, empty previous slot, whitespace-tolerant standard base64, identical previous key ignored, and a generated pair of distinct keys that share an 8-hex id. Add failing tests that a previous key opens, `NeedsRotation` is true only for a well-formed non-current key id, and `Reseal` rewrites only those tokens. Verify these tests fail on missing symbols.
- [x] 2.2 Implement `FromEnv`, `ErrKeyMissing`, `ErrInvalidKey`, `NeedsRotation`, and `Reseal` as specified in design.md. Verify the section 2 tests pass with `go test` and that `FromEnv` does not read dotenv files.

## 3. Provider-key associated data

- [x] 3.1 Add failing tests that `ProviderKeyAAD` returns exactly `provider_key:<scope>:<ownerID>:<provider>` for `workspace` and `user`, rejects empty fields, other scopes, and colons, and that a token does not open under another row's associated data. Verify the tests fail on the missing helper.
- [x] 3.2 Implement `ProviderKeyAAD`, `ErrInvalidAAD`, `ScopeWorkspace`, and `ScopeUser`. Verify the section 3 tests pass.

## 4. Dev key file and process resolver

- [x] 4.1 Add failing tests for `LoadOrCreateDevKeyFile` and `Resolve`: create mode `0600`, do not overwrite, reuse an existing file, production missing key creates nothing, invalid env does not fall back to a file, previous-without-current does not generate, loose permissions refused, symlink refused, missing parent directory refused. Verify the tests fail on missing symbols.
- [x] 4.2 Implement `LoadOrCreateDevKeyFile`, `ErrKeyFile`, and `Resolve` without logging. Verify the section 4 tests pass, including a second resolve that reports `createdDevKey == false`.

## 5. Operator documentation

- [x] 5.1 Document `MORPH_SECRETS_KEY` and `MORPH_SECRETS_KEY_PREVIOUS` in `.env.example` with `openssl rand -base64 32` and no key value. Add `docs/agents/14-secrets-key.md` and pointers from `docs/agents/03-morph.md`, `docs/agents/12-build-deploy.md`, and `docs/agents/13-conventions.md` covering the dev file path, the warning, production refusal (the file is not read; copy it into the env var first), the startup `Resolve` snippet, and the rotation steps. Verify those files contain the env names and the openssl command and do not contain a 32-byte base64 secret.

## 6. Module verification

- [ ] 6.1 Run `cd morph && go vet ./... && go test ./...` and verify both exit 0. Verify `git diff` does not touch `morph/config/config.go` or `morph/handlers/authz_middleware.go`.
