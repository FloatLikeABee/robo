# morph-secretbox Specification

## Purpose

Seals stored secrets, starting with AI provider API keys, under an app-level master key so a copy of the Morph database does not reveal key material, and so a later key rotation can still open old rows.

## Requirements

### Requirement: AES-256-GCM seal format

The package MUST seal plaintext with AES-256-GCM and a fresh 96-bit random nonce, and MUST return a token `v1:<keyid>:<payload>` where `payload` is standard base64 of `nonce || ciphertext || tag`. `keyid` MUST be the first 8 lowercase hex characters of SHA-256 over the raw 32-byte key, not over the base64 text. Two seals of the same plaintext and associated data MUST differ. Open MUST return the original plaintext when the version is `v1`, the key id is known, the payload is intact, and the associated data matches. New seals MUST use only the current key.

#### Scenario: Round trip

- **WHEN** a caller seals plaintext and associated data under a 32-byte current key and opens that token with the same associated data
- **THEN** Open returns the same plaintext
- **AND** a second seal of the same inputs is a different token

#### Scenario: Empty plaintext

- **WHEN** a caller seals an empty plaintext
- **THEN** Open returns an empty plaintext

### Requirement: Open fails closed with typed errors

Open MUST return one of three sentinel errors, and no other error, for a failed open: unknown key id, unsupported version, or decrypt failure. Decrypt failure MUST cover wrong associated data, truncated or non-canonical payloads, a malformed `v1` token, and any GCM authentication failure. A token whose version field is not `v1` MUST return the unsupported-version error, including an empty token. Open MUST choose the key solely by key id and MUST NOT trial-decrypt with any other key. Error text MUST NOT contain the plaintext, the raw key, or the base64 form of any key.

#### Scenario: Unknown key id is not a decrypt failure

- **WHEN** a token sealed under key A is opened by a box that does not hold A's key id
- **THEN** the error is the unknown-key-id sentinel
- **AND** it is not the decrypt-failure sentinel

#### Scenario: Unsupported version

- **WHEN** the token's first field is not `v1`
- **THEN** Open returns the unsupported-version sentinel

#### Scenario: Tamper or wrong associated data

- **WHEN** the payload is altered, truncated, or opened with different associated data
- **THEN** Open returns the decrypt-failure sentinel
- **AND** the error text does not contain the plaintext or the key

### Requirement: Env master key

`MORPH_SECRETS_KEY` MUST decode to exactly 32 bytes of standard base64 (padding optional; ASCII whitespace ignored; URL-safe alphabet rejected). `MORPH_SECRETS_KEY_PREVIOUS`, when set, MUST be a comma-separated list of keys in that same form and those keys MUST be decrypt-only. A missing or blank current key MUST be a distinct sentinel from an invalid encoding, length, empty list slot, or two different keys that share a key id. Error text MUST NOT echo the env value. The master key MUST NOT be derived from `JWT_SECRET`. An identical key listed again in the previous set MUST be ignored. Construction from the environment MUST read the process environment only and MUST NOT load dotenv files itself.

#### Scenario: Invalid env value

- **WHEN** `MORPH_SECRETS_KEY` is not 32 bytes of standard base64
- **THEN** construction from the environment fails with the invalid-key sentinel
- **AND** the error text does not contain the env value

#### Scenario: Missing env key

- **WHEN** `MORPH_SECRETS_KEY` is unset or blank
- **THEN** construction from the environment fails with the missing-key sentinel

#### Scenario: Empty previous slot

- **WHEN** `MORPH_SECRETS_KEY_PREVIOUS` has an empty comma-separated slot
- **THEN** construction fails with the invalid-key sentinel

#### Scenario: Distinct keys with the same key id

- **WHEN** two different 32-byte keys in the ring hash to the same 8-hex key id
- **THEN** construction fails with the invalid-key sentinel

### Requirement: Rotation

A token sealed with a previous key MUST open when the associated data matches. `NeedsRotation` MUST be true for a well-formed `v1` token whose key id is not the current key id, including a key id this box does not hold. `NeedsRotation` MUST be false for the current key id and for any token that is not a well-formed `v1` token with an 8-character lowercase hex key id. `NeedsRotation` MUST NOT decrypt the payload. The reseal helper MUST open the token and, when rotation is needed, return a new token sealed with the current key that opens to the same plaintext. When the token is already on the current key, reseal MUST return the original string and report that it did not rotate. A tampered token MUST fail reseal with the same error Open would return.

#### Scenario: Previous key opens and needs rotation

- **WHEN** plaintext was sealed under key P and the box's current key is a different key C with P in the previous set
- **THEN** Open returns the plaintext
- **AND** `NeedsRotation` is true
- **AND** reseal returns a token sealed under C that opens to the same plaintext and does not need rotation

#### Scenario: Current key is stable

- **WHEN** a token was sealed with the current key and the payload is intact
- **THEN** `NeedsRotation` is false
- **AND** reseal returns that same token string

### Requirement: Associated data binds the owning row

Canonical associated data for a provider-key row MUST be the UTF-8 string `provider_key:<scope>:<ownerID>:<provider>`. Scope MUST be `workspace` or `user`. The helper that builds this string MUST reject an empty field, any other scope, and any field that contains `:`. Seal and Open MUST authenticate the associated-data bytes the caller passes and MUST NOT invent or rewrite them. Opening a token with another row's associated data MUST fail with the decrypt-failure sentinel.

#### Scenario: Ciphertext moved to another row

- **WHEN** a secret sealed with `provider_key:workspace:owner-1:openai` is opened with `provider_key:workspace:owner-2:openai`
- **THEN** Open returns the decrypt-failure sentinel

#### Scenario: Helper rejects ambiguous fields

- **WHEN** the scope is not `workspace` or `user`, or any field is empty, or any field contains `:`
- **THEN** the helper returns an error and does not return associated-data bytes

### Requirement: Dev key file and production refusal

The process resolver MUST take an explicit production flag supplied by the caller. It MUST NOT read `MORPH_ENV` itself. The caller supplies the bool from `config.ParseMorphEnv()` (`production`/`prod` = production; unset/`development`/`dev`/`local`/`test` = local; anything else refuses to start) and `config.ValidateStartup(cfg, prod)` in `morph/config/startup.go`. When production is true, or when `MORPH_SECRETS_KEY` is set, the resolver MUST use only the environment and MUST NOT create or read a key file. A missing or invalid environment key in that case MUST be returned to the caller so startup can refuse. When production is false and both env vars are unset, the resolver MUST load or create a key file at the caller-supplied path. Creation MUST use mode `0600`, MUST write standard base64 of 32 random bytes plus a trailing newline, and MUST report that the file was created. An existing file MUST NOT be overwritten. The resolver MUST refuse a symlink, a non-regular file, a missing parent directory, a file it cannot prove is the same file it stat-ed, a file whose mode grants any group or other permission bits, and a file that does not decode to 32 bytes. When production is false, the current key is unset, and previous keys are set, the resolver MUST fail with the invalid-key sentinel and MUST NOT generate a file. The package MUST NOT log. Morph startup MUST call the resolver after `NewTranSQL` with that bool. A resolver error MUST refuse process start. The refusal MUST name `MORPH_SECRETS_KEY` when that variable is missing or invalid, MUST name `MORPH_SECRETS_KEY_PREVIOUS` when that variable is the problem, and MUST NOT include either value. When the resolver creates the dev file, startup MUST log a warning that contains the path and MUST NOT log the key. A set `MORPH_SECRETS_KEY_PREVIOUS` MUST be loaded for decrypt.

#### Scenario: Dev creates a private key file

- **WHEN** production is false, both secrets env vars are unset, and the path does not exist inside an existing directory
- **THEN** the resolver creates a regular file with mode `0600`
- **AND** the file decodes to the 32-byte key the returned box uses to seal
- **AND** the resolver reports that it created the file

#### Scenario: Production does not invent a key

- **WHEN** production is true and `MORPH_SECRETS_KEY` is unset
- **THEN** the resolver returns the missing-key sentinel
- **AND** no key file is created

#### Scenario: Invalid env does not fall back to a file

- **WHEN** production is false and `MORPH_SECRETS_KEY` is set but invalid
- **THEN** the resolver returns the invalid-key sentinel
- **AND** no key file is created

#### Scenario: Existing dev key is reused

- **WHEN** production is false, both secrets env vars are unset, and a valid mode-`0600` key file already exists
- **THEN** the resolver does not rewrite the file
- **AND** it reports that it did not create the file
- **AND** tokens sealed by the returned box open under the key stored in that file

#### Scenario: Loose permissions are refused

- **WHEN** an existing key file grants group or other permission bits
- **THEN** the resolver returns an error and does not use the file's contents as a key

#### Scenario: Startup refuses a bad production key and keeps a local file private

- **WHEN** the Morph process reaches startup after the Tran data directory exists
- **THEN** production with a missing or invalid `MORPH_SECRETS_KEY` refuses to start with an error that names that variable and does not include the value
- **AND** local mode with both secrets env vars unset uses or creates `morph-secrets.key` at mode `0600` in that directory and logs a warning that contains the path and not the key
- **AND** a set `MORPH_SECRETS_KEY_PREVIOUS` is available to open tokens sealed with that older key

#### Scenario: Symlink is refused

- **WHEN** the key path is a symlink, including a symlink to a valid key file
- **THEN** the resolver returns an error and does not use the target as a key

### Requirement: Operator documentation

`.env.example` MUST name `MORPH_SECRETS_KEY` and `MORPH_SECRETS_KEY_PREVIOUS`, MUST show the generation command `openssl rand -base64 32`, and MUST NOT contain a usable key. Agent docs MUST state that the key is independent of `JWT_SECRET`, that local dev creates `<data-dir>/morph-secrets.key` with mode `0600` and a warning when the env key is unset, that production startup refuses a missing or invalid key and does not read that file (copy the file into `MORPH_SECRETS_KEY` before enabling production), and the rotation steps: put the new key in `MORPH_SECRETS_KEY` and the old key in `MORPH_SECRETS_KEY_PREVIOUS`, restart, reseal every row that needs rotation, then remove the previous key. The recommended data directory is the directory of `TRAN_SQLITE_PATH` (default `./data`, cwd-relative).

#### Scenario: Example env has no real key

- **WHEN** an operator reads `.env.example`
- **THEN** it names both secrets env vars and the command `openssl rand -base64 32`
- **AND** it does not assign a 32-byte key value

#### Scenario: Docs describe dev, production, and rotation

- **WHEN** an operator reads the secrets-key agent doc
- **THEN** it explains the `0600` dev file under the data directory, the warning, production refusal without reading that file, and the reseal-then-drop-previous rotation steps
