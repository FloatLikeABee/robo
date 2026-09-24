# Proposal

## Why

Morph starts with a published admin password (`admin123`), a published JWT signing secret, and a JWT lifetime of about 100 years. Nothing stops those defaults from being used on a hosted process. An operator who already seeded SQLite on an earlier dev run would keep the default password even after changing the env var, because bootstrap does not update an existing account.

## What Changes

- Add an explicit `MORPH_ENV` mode. `production` (and `prod`) refuses to start when the JWT secret or admin password is still a development default, empty, or obviously weak, and when JWT lifetime is longer than the production cap.
- Local/dev mode (unset, `development`, `dev`, `local`, `test`) keeps starting with the defaults and logs a warning. `./start-all.sh` and a fresh checkout need no new config.
- Unrecognized `MORPH_ENV` values refuse to start, so a typo or `staging` cannot silently keep the defaults.
- In production, also refuse when an admin account already stored in SQLite still accepts the development password. A one-shot `MORPH_ROTATE_DEFAULT_ADMIN=1` replaces those stored hashes with the configured password, then the process continues.
- Production JWT lifetime defaults to 24 hours when unset, stays configurable, and rejects values above 168 hours.
- Document the flag, the required overrides, and a hosting checklist of secrets to set or rotate. Startup errors name the variables to set and never print secret values.

## Capabilities

### New Capabilities

- `morph-production-secrets`: Production startup guard for Morph JWT secret, admin password (env and already-seeded SQLite), and JWT lifetime, plus the local/dev exception and hosting checklist.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baseline)

## Impact

- `morph/config` (mode, secret checks, expiry resolution), `morph/auth/jwt.go` (issue tokens with the same expiry rules), `morph/db` (detect and optionally rotate a stored default admin password), `morph/main.go` (fail closed before serving).
- Root `.env.example`, root README, Morph README, `docs/agents/01-auth-flow.md`, `docs/agents/03-morph.md`, new `docs/security-hosting-checklist.md`.
- Does not change auth middleware, provider-key encryption, secret managers, `.github/`, or `docs/agents/12-build-deploy.md`. Does not change how `./start-all.sh` launches Morph.
