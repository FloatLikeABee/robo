# morph-production-secrets Specification

## Purpose

Stops a hosted Morph process from signing sessions or accepting the published development admin password, while leaving a zero-config local `start-all.sh` checkout able to start.

## Requirements

### Requirement: Explicit production mode

Morph MUST treat `MORPH_ENV` of `production` or `prod` (case-insensitive) as production mode. Morph MUST treat an unset value, or `development`, `dev`, `local`, or `test` (case-insensitive), as local/dev mode. Any other `MORPH_ENV` value MUST prevent startup. The refusal MUST name `MORPH_ENV` and the recognized values. It MUST NOT print JWT secret or admin password values.

#### Scenario: Production flag enables the guard

- **WHEN** `MORPH_ENV` is `production` or `prod` and a development JWT secret or development admin password is configured
- **THEN** the process exits before it serves HTTP
- **AND** the error says what to set

#### Scenario: Local mode is the default

- **WHEN** `MORPH_ENV` is unset or set to `development`, `dev`, `local`, or `test`
- **AND** the development JWT secret and development admin password are configured
- **THEN** Morph still starts

#### Scenario: Unrecognized mode fails closed

- **WHEN** `MORPH_ENV` is set to a value other than the recognized production and local/dev values
- **THEN** the process exits before it serves HTTP
- **AND** the error names `MORPH_ENV` and the recognized values

### Requirement: Production rejects a weak JWT secret

In production mode, Morph MUST refuse to start when `JWT_SECRET` is empty, is the development default `morph-dev-jwt-secret-change-me` (case-insensitive, ignoring surrounding whitespace), is shorter than 32 characters, is a single repeated character, or contains a placeholder fragment `change-me`, `changeme`, `change_me`, `replace-me`, or `replace_me` (case-insensitive). The error MUST name `JWT_SECRET` and MUST NOT include the secret value.

#### Scenario: Development default secret

- **WHEN** production mode is on and `JWT_SECRET` is missing or is the development default
- **THEN** the process exits before it serves HTTP
- **AND** the error tells the operator to set `JWT_SECRET` to a unique random string of at least 32 characters
- **AND** the error text does not contain the secret value

#### Scenario: Short or placeholder secret

- **WHEN** production mode is on and `JWT_SECRET` is non-empty but shorter than 32 characters, a repeated character, or a known placeholder
- **THEN** the process exits before it serves HTTP
- **AND** the error names `JWT_SECRET` and does not echo the configured value

#### Scenario: Strong secret is accepted

- **WHEN** production mode is on and `JWT_SECRET` is a unique random string of at least 32 characters that is not a placeholder
- **AND** the admin password and JWT lifetime also satisfy production rules
- **THEN** the secret check does not prevent startup

### Requirement: Production rejects a weak admin password from the environment

In production mode, Morph MUST refuse to start when the effective bootstrap admin password (`ADMIN_PASSWORD`, otherwise `BOOTSTRAP_ADMIN_PASSWORD`) is empty, matches `admin123` case-insensitively, or is shorter than 12 characters. The error MUST name `ADMIN_PASSWORD` and MUST NOT include the password value. Local/dev mode MUST still accept the development password.

#### Scenario: Development default password

- **WHEN** production mode is on and the effective admin password is missing or is `admin123`
- **THEN** the process exits before it serves HTTP
- **AND** the error tells the operator to set `ADMIN_PASSWORD` to a unique password of at least 12 characters
- **AND** the error text does not contain the password

#### Scenario: Unique password is accepted

- **WHEN** production mode is on and `ADMIN_PASSWORD` is a unique password of at least 12 characters
- **AND** the JWT secret and JWT lifetime also satisfy production rules
- **THEN** the environment password check does not prevent startup

### Requirement: Production detects an already-seeded default admin password

In production mode, after the login user table exists and before serving HTTP, Morph MUST refuse to start when any of these accounts still accepts the development password `admin123`: a user with the Admin role, the development identity `morphadmin` / `morphadmin@local.com`, or the configured bootstrap email or username. Checking only the environment variable is not sufficient. The error MUST NOT include the password. A non-admin user who is not that bootstrap identity and who happens to use the same password MUST NOT by itself block startup.

#### Scenario: Database still has the development admin password

- **WHEN** production mode is on and the environment admin password is already unique and at least 12 characters
- **AND** an Admin account in SQLite still accepts the development password
- **THEN** the process exits before it serves HTTP
- **AND** the error tells the operator to set `ADMIN_PASSWORD` and start once with `MORPH_ROTATE_DEFAULT_ADMIN=1`

#### Scenario: Fresh database with a strong password

- **WHEN** production mode is on, the environment admin password is unique and at least 12 characters, and no stored account accepts the development password
- **THEN** Morph creates the bootstrap admin if missing and starts

#### Scenario: One-shot rotation replaces stored development passwords

- **WHEN** production mode is on, `MORPH_ROTATE_DEFAULT_ADMIN` is truthy (`1`, `true`, `yes`, or `on`), and `ADMIN_PASSWORD` meets the production password rules
- **AND** one or more of the accounts described above still accept the development password
- **THEN** Morph replaces those stored password hashes with the configured admin password, keeps each account id, and continues startup
- **AND** a later start without the flag succeeds because those accounts no longer accept the development password

#### Scenario: Rotation flag does not excuse a weak new password

- **WHEN** production mode is on and `MORPH_ROTATE_DEFAULT_ADMIN` is set but `ADMIN_PASSWORD` is still the development default or shorter than 12 characters
- **THEN** the process exits before it serves HTTP
- **AND** no stored hash is replaced

#### Scenario: Login table cannot be checked

- **WHEN** production mode is on and the login user table cannot be created or opened
- **THEN** the process exits before it serves HTTP
- **AND** a failure of an unrelated schema step does not skip this check when the login table itself is usable

### Requirement: Local mode warns and still starts

In local/dev mode, Morph MUST start when the development JWT secret and development admin password are in use, including when an already-seeded admin account still accepts the development password. Morph MUST log a warning that names the variables involved and points at the hosting checklist. The warning MUST NOT include secret values. `./start-all.sh` MUST NOT need `MORPH_ENV`, `JWT_SECRET`, or `ADMIN_PASSWORD` overrides for a fresh checkout to start Morph.

#### Scenario: Fresh local checkout

- **WHEN** an operator runs Morph the way `./start-all.sh` runs it, with no extra environment beyond the repo defaults
- **THEN** Morph starts
- **AND** the development admin can sign in with the documented local password

#### Scenario: Local warning

- **WHEN** local/dev mode starts with the development JWT secret or development admin password
- **THEN** the process logs a warning that these defaults are allowed only outside production
- **AND** the warning does not print the secret or password

### Requirement: Production JWT lifetime is short and configurable

Issued Morph JWTs MUST use one lifetime rule for both the startup check and token signing. In production mode, when `JWT_EXPIRY_HOURS` is unset, the lifetime MUST be 24 hours. A configured integer from 1 through 168 hours MUST be used as-is. A non-integer, a value below 1, or a value above 168 MUST prevent startup, and the error MUST name `JWT_EXPIRY_HOURS` and the allowed range. In local/dev mode, an unset or invalid `JWT_EXPIRY_HOURS` MUST keep the current development lifetime of 876000 hours, and an explicit positive value (including 876000) MUST be honored.

#### Scenario: Production default lifetime

- **WHEN** production mode is on and `JWT_EXPIRY_HOURS` is unset
- **AND** the JWT secret and admin password satisfy production rules
- **THEN** Morph starts
- **AND** newly issued tokens expire in 24 hours

#### Scenario: Production rejects the development lifetime

- **WHEN** production mode is on and `JWT_EXPIRY_HOURS` is `876000`
- **THEN** the process exits before it serves HTTP
- **AND** the error names `JWT_EXPIRY_HOURS` and the maximum of 168 hours

#### Scenario: Configured production lifetime

- **WHEN** production mode is on and `JWT_EXPIRY_HOURS` is `48`
- **AND** the JWT secret and admin password satisfy production rules
- **THEN** newly issued tokens expire in 48 hours

#### Scenario: Local lifetime stays long

- **WHEN** local/dev mode is on and `JWT_EXPIRY_HOURS` is unset or `876000`
- **THEN** newly issued tokens expire in 876000 hours

### Requirement: Hosting checklist lists secrets to set

The repository MUST include a hosting checklist that tells an operator to set `MORPH_ENV=production`, replace `JWT_SECRET`, replace `ADMIN_PASSWORD` (and rotate an already-seeded development password), and set `JWT_EXPIRY_HOURS` within the production range. The checklist MUST also list `MORPH_AI_API_KEY` and state that `.env` must not be committed. Root `.env.example` MUST document `MORPH_ENV`, the production overrides, and that the checked-in development values are for local runs only. The root README MUST state that local/dev still starts with the development admin login and that hosting uses the checklist.

#### Scenario: Operator follows the checklist

- **WHEN** an operator reads the hosting checklist before a hosted deploy
- **THEN** it names `MORPH_ENV`, `JWT_SECRET`, `ADMIN_PASSWORD`, `MORPH_ROTATE_DEFAULT_ADMIN`, `JWT_EXPIRY_HOURS`, and `MORPH_AI_API_KEY`
- **AND** it does not contain a live production secret
