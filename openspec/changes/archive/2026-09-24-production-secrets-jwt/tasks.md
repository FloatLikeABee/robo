# Tasks

## 1. Startup guard and JWT lifetime

- [x] 1.1 Add failing `dorph/config` tests for production defaults rejected, production real values accepted, local/dev defaults accepted, unrecognized `DORPH_ENV` rejected, short/placeholder/empty JWT secrets rejected without echoing the value, admin password default and short values rejected without echoing the value, and expiry (production unset = 24, production 48 accepted, production 876000 rejected, local unset = 876000). Run the new tests and confirm they fail because the guard is missing.
- [x] 1.2 Implement `DORPH_ENV` parsing, `ValidateStartup`, development warnings, and `JWTExpiryHours` in `dorph/config` per design.md. Normalize a whitespace `JWT_SECRET` to the development default. Verify the tests from 1.1 pass.

## 2. Token signing uses the same lifetime

- [x] 2.1 Add a failing `dorph/auth` test that a production `LoadTokenConfig` with unset `JWT_EXPIRY_HOURS` expires in 24 hours, a production value of 48 is honored, a local unset value stays 876000, and a production value of 876000 does not return 876000 hours. Run it and confirm it fails for the right reason.
- [x] 2.2 Point `LoadTokenConfig` at the shared expiry helper and resolved JWT secret. Verify the tests from 2.1 pass and `cd dorph && go test ./auth ./config` passes.

## 3. Already-seeded admin password

- [x] 3.1 Add failing `dorph/db` tests: stored development admin password is reported; `GuardStoredDefaultAdminPassword` refuses without the rotate flag and does not echo the password; rotate replaces those hashes and keeps the user id; a second Admin on the development password is included; a non-admin who is not the bootstrap identity is ignored; an empty database passes; rotate refuses a weak new password and does not write. Run them and confirm they fail because the guard is missing.
- [x] 3.2 Implement the scan and production-only rotation in `dorph/db` using bcrypt compare and an id-preserving `password_hash` update. Verify the tests from 3.1 pass.

## 4. Process startup wiring

- [x] 4.1 Call the env guard before stores open, decouple invite-schema setup from the admin `else if` chain, fatal in production when `plat_users` cannot be ensured or bootstrap fails, run the stored-password guard before `EnsureBootstrapAdmin`, and log local warnings without secret values. Verify `cd dorph && go test ./...` still passes and `gofmt` is clean on touched Go files. Do not edit `dorph/handlers/authz_middleware.go`.

## 5. Hosting docs

- [x] 5.1 Add `docs/security-hosting-checklist.md` naming `DORPH_ENV`, `JWT_SECRET`, `ADMIN_PASSWORD`, `DORPH_ROTATE_DEFAULT_ADMIN`, `JWT_EXPIRY_HOURS`, and `DORPH_AI_API_KEY`, with no live production secret. Update `.env.example`, the root README, `dorph/README.md`, `docs/agents/01-auth-flow.md`, and `docs/agents/03-dorph.md` so local defaults stay documented and hosting points at the checklist. Do not edit `docs/agents/12-build-deploy.md` or `.github/`. Verify those names are present with a search.

## 6. Integration check

- [x] 6.1 Run `cd dorph && go vet ./... && go test ./...` and confirm both exit 0. No frontend or Rust packages are changed, so do not run those builds.
