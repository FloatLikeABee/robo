# Design

## Context

See proposal.md for why. Behavior contract: `specs/morph-production-secrets/spec.md`.

Checked against current code:

- `morph/config/config.go` `GetConfig` fills `JWT_SECRET` with `morph-dev-jwt-secret-change-me` and `ADMIN_PASSWORD` / `BOOTSTRAP_ADMIN_PASSWORD` with `admin123` when unset. `getEnv` treats whitespace as set. `firstNonEmptyEnv` prefers `ADMIN_PASSWORD` over `BOOTSTRAP_ADMIN_PASSWORD`.
- `morph/auth/jwt.go` `LoadTokenConfig` reads those env vars again on its own. Empty secret becomes the same default. Unset `JWT_EXPIRY_HOURS` becomes `100*365*24` (876000). `handlers.New` calls `LoadTokenConfig` once at startup.
- `morph/db/plat_users.go` `EnsureBootstrapAdmin` inserts the env password only when the email/username is missing. An existing row is left unchanged, so a later env change does not rotate a seeded `admin123` hash. `EnsureBootstrapAdminForce` overwrites the first row matching email OR username, including email, username, and roles. `VerifyPassword` is bcrypt.
- `morph/main.go` chains `EnsurePlatUsersTable`, then `EnsurePlatInviteCodesTable`, then `EnsureBootstrapAdmin` with `else if`. An invite-table error skips admin bootstrap entirely. Schema errors are warnings; the process still serves.
- Root `.env.example` publishes the three development values, including `JWT_EXPIRY_HOURS=876000`. It also sets `APP_ENV=development` for another app. `start-all.sh` does not set `MORPH_ENV`.
- Tokens are HS256 with no server-side revoke list. `morph-engi` reads `JWT_SECRET` and falls back to its own dev secret if unset. That fallback is out of scope to change; the checklist mentions it.

## Goals / Non-Goals

**Goals:**

- One mode flag, one expiry helper, and one stored-password check, wired so startup refusal and token signing cannot disagree.
- Local `./start-all.sh` stays zero-config.
- Errors and warnings name variables and never echo secret values.

**Non-Goals:**

- Secret manager, encryption of stored provider API keys, auth middleware (`morph/handlers/authz_middleware.go`).
- Changing Project's own JWT fallback in Rust.
- Editing `.github/` or `docs/agents/12-build-deploy.md`.
- A password-strength library, refresh tokens, or revocation.

## Decisions

### 1. Mode flag is `MORPH_ENV`, not `APP_ENV` or a boolean

- **Choice**: `production` and `prod` enforce the guard. Unset, `development`, `dev`, `local`, and `test` are local/dev. Any other value refuses startup and names the recognized values. Comparison is case-insensitive. `main` calls `GetConfig` (which loads the root `.env`) before reading `MORPH_ENV`, because `repoenv` does not override variables that are already set, but it does populate unset ones.
- **Rationale**: The issue asks for an explicit flag. `.env.example` already uses `APP_ENV=development` for a different app in the same file. A boolean cannot tell `staging` apart from a typo.
- **Rejected**: Reuse `APP_ENV` — a host setting it for Project/Booki would toggle Morph's guard. `MORPH_PRODUCTION=1` — no fail-closed path for `staging`. Treating every non-dev value as production — the error would claim production mode for a typo.

### 2. Env check and stored-hash check are both required

- **Choice**: `ValidateStartup` runs on the resolved config before stores open. After `plat_users` exists, production scans stored hashes. Then `EnsureBootstrapAdmin` runs. Production treats a `plat_users` schema error as fatal. Invite-code schema stays a warning and is not in the same `else if` chain, so it cannot skip the guard or bootstrap.
- **Rationale**: Env-only misses the seeded row (`EnsureBootstrapAdmin` returns nil when the row exists). DB-only misses the first production boot, which would insert `admin123` if that were still the env value. The invite `else if` in `main.go` is a real bypass if left as-is.
- **Rejected**: Env only. DB only. Auto-updating the hash whenever the env password is strong — that mutates data on every deploy that still has the default hash, without an explicit operator act, and the acceptance criterion is refusal when the default is detected.

### 3. Rotation is explicit, id-preserving, and production-only

- **Choice**: `MORPH_ROTATE_DEFAULT_ADMIN` truthy (`1`, `true`, `yes`, `on`) in production replaces `password_hash` on each matched account with a fresh bcrypt hash of the configured password (one hash per row, not one hash shared across rows). Account ids stay. The flag does nothing in local/dev, so a line left in a dev `.env` cannot change the documented `morphadmin` / `admin123` login. A later start with the flag still set finds no matching hashes and does not write.
- **Matched accounts**: Admin role, or email/username equal to `morphadmin@local.com` / `morphadmin`, or equal to the configured bootstrap email/username. A non-admin who is none of those does not block startup or get rotated.
- **Rejected**: `EnsureBootstrapAdminForce` — it rewrites email, username, and roles on a single `email OR username` match and leaves any other admin on the default password. Deleting the row and reseeding — new user id. Rotating in dev — breaks the documented local login if the flag leaks into `.env`.

### 4. Weak-secret rules stay explainable

- **Choice**: Production JWT secret must be non-empty after trim, not the development default (case-insensitive), at least 32 characters, not a single repeated character, and must not contain `change-me`, `changeme`, `change_me`, `replace-me`, or `replace_me`. Whitespace-only `JWT_SECRET` is normalized to the development default so local signing still works, and production then rejects that default. Admin password must not match `admin123` (case-insensitive) and must be at least 12 characters. The default-password message is used when it matches `admin123`; the length message is used otherwise. Several problems are returned in one error.
- **Rationale**: The published secret is 30 characters and contains `change-me`, so length alone would refuse it but would not say "development default". A 32-character string of one character would pass a length-only check. Twelve characters rejects `admin123` even without a special case, and the special case keeps the error specific.
- **Rejected**: zxcvbn or an entropy estimator — new dependency, harder errors. Minimum length only — the error would not name the published default. Rejecting any secret that merely contains the substring `secret` or `password` — those show up in real passphrases.

### 5. One expiry helper shared by the guard and `LoadTokenConfig`

- **Choice**: Production, unset → 24 hours. Production, integer 1 through 168 → that value. Production, anything else → startup error naming `JWT_EXPIRY_HOURS` and the max of 168. Local/dev, unset or invalid → 876000. Local/dev, any positive integer, including 876000 → that value. `LoadTokenConfig` calls the same helper. If it is ever invoked when the helper returns an error, it signs with 24 hours and logs the error, not with 876000.
- **Decode**: `EncodeToken` sets both `iat` and `exp`. In production, `DecodeToken` rejects a token with either claim missing, or with `exp - iat` longer than 168 hours (the production ceiling, not the current `JWT_EXPIRY_HOURS`). Local/dev decode does not apply that window check, so an existing 876000-hour token still verifies locally. The ceiling is 168 rather than the operator's current setting so shortening `JWT_EXPIRY_HOURS` from 48 to 24 does not invalidate tokens that were already inside the ceiling.
- **Rationale**: `LoadTokenConfig` and `GetConfig` read the env separately today. Validating in config while leaving `LoadTokenConfig` unchanged would let production start (implicit 24h check) and still issue 100-year tokens. There is no revoke list, so the cap is 7 days rather than 30. Checking only new issuances leaves tokens already signed for 876000 hours valid after `MORPH_ENV=production`, which is the same hole. Project validates the same secret on its own and does not apply this window, so the checklist also requires a new `JWT_SECRET` when production mode is first enabled.
- **Rejected**: Change the global default to 24h — local sessions and the current dev default change. Cap at 720 hours — too long for an unrevocable admin bearer. Require an explicit `JWT_EXPIRY_HOURS` with no production default — fails a host that set secrets and expected the documented 24h default.

### 6. Docs and local ergonomics

- **Choice**: New `docs/security-hosting-checklist.md`. Point to it from the root README, `morph/README.md`, `docs/agents/01-auth-flow.md`, `docs/agents/03-morph.md`, and the auth section of `.env.example` (the file header does not send hosts to `deploy/.env.production`). `.env.example` keeps the development secret, password, and `JWT_EXPIRY_HOURS=876000`, with comments that those are local-only and that hosting sets `MORPH_ENV=production` plus the overrides. Do not set `MORPH_ENV` inside `start-all.sh`. Do not enable the rotate flag in the example file. The checklist says to generate a new `JWT_SECRET` when first enabling production mode, and to copy it to every service that validates Morph tokens (Project / morph-engi included). No production secret is committed.
- **Rejected**: Putting `MORPH_ENV=production` in `.env.example` — a copied example would refuse to start locally. Changing `start-all.sh` — unnecessary if unset means dev.

## Risks / Trade-offs

- [Operator flips `MORPH_ENV=production` on a dev database and does not set the rotate flag] → Startup names `MORPH_ROTATE_DEFAULT_ADMIN` and does not change the hash. Checklist documents one successful start, then unsetting the flag.
- [Rotate flag left on] → No-op once hashes no longer match the development password. It does not keep resetting a later custom password.
- [Many Admin rows] → Bcrypt only those candidates, at existing cost 10. This app's user table is small; no parallel compare.
- [Unrecognized `MORPH_ENV` on a laptop] → Process exits with the recognized values. Safer than serving defaults.
- [Invite-table warning no longer skips bootstrap] → A database that previously skipped admin creation because invite schema failed will now create the bootstrap admin. That is the intended auth startup.
- [`LoadTokenConfig` 24h fallback if validation was skipped] → Avoids a 100-year token. `main` still refuses to start when the helper returns an error, so the fallback is not the hosted path.
- [Already-issued 876000-hour token, same strong secret] → Morph rejects it in production because `exp - iat` exceeds 168 hours. Project does not. Checklist requires a new `JWT_SECRET` on first production enable so those other verifiers drop the old sessions too.

## Migration Plan

1. Ship the guard with local mode as the default. Existing `./start-all.sh` checkouts keep working.
2. Before hosting: set `MORPH_ENV=production`, a new 32+ character `JWT_SECRET` (even if the current secret is already strong), a 12+ character `ADMIN_PASSWORD`, and `JWT_EXPIRY_HOURS` from 1 to 168 (or leave it unset for 24). Copy the new secret to every service that validates Morph tokens. If SQLite was seeded earlier, set `MORPH_ROTATE_DEFAULT_ADMIN=1` for one start, confirm the log, then unset it.
3. Rollback: revert the commit. Local data is unchanged unless the rotate flag ran; that password change is a hash update only and is not automatically reversed.

## Open Questions

None that change the spec, the approach, or the task breakdown.

## Design review

Proposer and reviewer pass, against the code above. The reviewer asked for three changes before this document was treated as stable:

1. Unknown `MORPH_ENV` must fail closed. An allowlist of only `production`/`prod` would let `staging` keep the published secrets. An implicit "everything else is production" would mislabel the error. The unrecognized-value error is the one that survived.
2. The invite `else if` must not be able to skip the stored-password check. That is in decision 2, not only in a comment.
3. Expiry resolution must be the same function `LoadTokenConfig` uses. A config-only check was rejected after reading `jwt.go`.

Rejected approaches that did not survive are listed under each decision (mode flag, env-only check, silent auto-rotate, `EnsureBootstrapAdminForce`, global 24h default, 30-day cap, `APP_ENV`). No further spec change is required for those.
