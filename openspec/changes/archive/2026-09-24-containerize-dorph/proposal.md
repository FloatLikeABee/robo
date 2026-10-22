## Why

Dorph API and the Dorph AI UI still start only from a checkout (`start-all.sh` or a local `go build` plus the CRA dev server). Issue #52 needs a reproducible image so an operator can run that one app on a host that is not chosen yet, without baking secrets into the image.

## What Changes

- Add a multi-stage image that builds the CRA UI and a static Dorph API binary, then serves both on `PORT` (default 9090) with `GET /health`.
- Point the image's data env defaults at a single `/data` mount (SQLite, Badger, knowledge, uploads).
- Add compose that starts the stack with one command, a named volume, and a gitignored env file. TLS (Caddy) stays an optional profile for issue #53.
- Document build, run, volume, env, backup, and upgrade in `deploy/README.md`, and point `docs/agents/12-build-deploy.md` at it.
- Check the image build in CI without adding or renaming a required check. The five existing check names stay as they are.

## Capabilities

### New Capabilities

- `dorph-container`: One image and compose file run Dorph API plus the Dorph AI UI, persist data on `/data`, take secrets from the environment only, and are build-checked outside the five required CI check names.

### Modified Capabilities

- (none)

## Impact

- New files: root `Dockerfile` and `.dockerignore`, `deploy/docker-compose.yml`, optional Caddy profile, `deploy/.env.production.example`, `deploy/README.md`, a non-required image-build workflow.
- Docs pointer in `docs/agents/12-build-deploy.md`.
- No Go or frontend behavior change. Local `./data` defaults and `start-all.sh` stay. Other apps (Event Logs, Content Maker, Data Access, Project, AI tools, DorphUtils, Invite Signup) are out of the image.
- Production startup still refuses default `JWT_SECRET` / `ADMIN_PASSWORD` when `DORPH_ENV=production` (existing `dorph-production-secrets` spec).
