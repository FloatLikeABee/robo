## 1. Contract check

- [x] 1.1 Add `deploy/check-container-contract.sh` that fails unless the Dockerfile builds with `CGO_ENABLED=0`, health-checks `/health`, drops to `morph` via `su-exec`, and has no secret `ARG`; `.dockerignore` excludes `node_modules`, env files, data, and the other app trees; compose mounts a named volume at `/data`, uses a gitignored `env_file`, and keeps Caddy on a profile; the example env has empty `JWT_SECRET` and `ADMIN_PASSWORD`; `.github/workflows/ci.yml` still has only the five existing check names.
- [x] 1.2 Run the script once before the image files exist and confirm it fails.

## 2. Image and compose

- [x] 2.1 Add the root multi-stage `Dockerfile` and `.dockerignore` from design.md (Node 22 UI build, Go 1.25 static binary, Alpine runtime, `/data` env defaults, entrypoint drop to uid 65532).
- [x] 2.2 Add `deploy/docker-compose.yml`, `deploy/Caddyfile` (profile `tls` only), and `deploy/.env.production.example` with empty secrets and no `JWT_EXPIRY_HOURS`.
- [x] 2.3 Re-run `deploy/check-container-contract.sh` and confirm it passes.

## 3. Docs and CI

- [x] 3.1 Add `deploy/README.md` (build, run, volume, env, backup, upgrade, single replica, optional Caddy profile) and point `docs/agents/12-build-deploy.md` at it without removing the Invite Signup row.
- [x] 3.2 Add `.github/workflows/docker-image.yml` that runs the contract script and `docker build` on pull requests and pushes to `main`, and does not edit `ci.yml`.

## 4. Verify

- [ ] 4.1 `go test ./...` in `morph`.
- [ ] 4.2 Build and `docker compose up` with a throwaway production env (strong secrets generated at runtime, not committed). `GET /health` is 200, `GET /` is the Morph AI index, login works, a note survives `down` and `up`, the server uid is 65532, and production mode refuses a default secret.
