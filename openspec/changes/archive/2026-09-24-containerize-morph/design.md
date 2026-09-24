## Context

See proposal.md for why. Behavior is in `specs/morph-container/spec.md`.

Verified against the tree (after fast-forward to `origin/main`):

- `morph/go.mod` is `go 1.25.0`. SQLite is `modernc.org/sqlite`. Badger is v4. No `cgo` import in `morph` or the replaced `pkg/*` modules. `CGO_ENABLED=0` is a valid static build.
- `replace` directives point at `../pkg/morphai`, `../pkg/repoenv`, `../pkg/docextract`, `../pkg/morphgraph`, and `../pkg/webresearch`. A build whose context is only `morph/` cannot resolve them.
- `morph/main.go` serves `./frontend/build/static` and `./frontend/build/index.html`, and the SPA `NoRoute` also returns that `index.html`. `r.Run(":" + cfg.Port)` listens on all interfaces. `GET /health` is registered in `handlers.RegisterAPIRoutes` and returns 200 without checking a live model.
- `morph/frontend` `npm run build` is `node ../../scripts/with-root-env.cjs craco build`. That script loads a repo-root `.env` only when it can see `start-all.sh`, and it does not override variables already set. `src/apiBase.js` uses `REACT_APP_API_URL` or `''` (same origin). An empty value is what the combined image needs.
- `ai.New` returns success when `MORPH_AI_API_KEY` is empty. Login and notes do not need a model key.
- Data env defaults in `morph/config/config.go` are cwd-relative (`./data/badger`, `./data/tran.sqlite`, `./data/entity_details`, `uploads/entity_attachments`). `MORPH_KNOWLEDGE_DIR` is read in `handlers/knowledge.go` (default `data/knowledge`), not on the config struct. Changing those Go defaults would change `start-all.sh`.
- Production refusal lives in `morph/config/startup.go` (`MORPH_ENV=production` rejects the published JWT default, short secrets, and `admin123` / short admin passwords). `JWT_EXPIRY_HOURS=876000` also refuses in production. Unset expiry becomes 24 hours.
- `openspec/specs/platform-ci/spec.md` says the platform CI workflow MUST report exactly these check names: `Go / Morph API`, `Go / Event Logs`, `Go / Content Maker`, `Go / morphai`, `Morph frontend`. A new job in `.github/workflows/ci.yml` would break that.
- `deploy/` does not exist. `docs/agents/12-build-deploy.md` still says so. `scripts/deploy.sh` is not a runbook.
- `products/` is a hardcoded relative directory, not one of the env paths above. It is not part of this capability.

## Goals / Non-Goals

**Goals:**

- One image, one compose command, `/data` volume, env-only secrets, healthcheck, short runbook, image build on every PR without a sixth required check name.
- The Morph process runs as uid 65532. Local checkout defaults stay cwd-relative.

**Non-Goals:**

- Images or compose services for any app other than Morph.
- Choosing or deploying to a host. No Render, Fly, or VPS apply step.
- Finishing TLS (issue #53). The Caddy profile is a stub.
- Moving `products/` onto `/data`.
- Kubernetes, multi-replica SQLite, or a Neo4j sidecar. Graph ingest already no-ops when Neo4j is down and does not block HTTP.

## Decisions

### 1. One image built from the repo root

- **Choice:** Root `Dockerfile`. Node stage runs `npm ci` and `CI=true npm run build` in `morph/frontend`. Go stage builds `CGO_ENABLED=0` with `-trimpath -ldflags "-s -w"`. Final stage copies only the binary and `frontend/build`.
- **Rationale:** `main.go` already serves the CRA build. Same-origin `API_BASE_URL` stays empty, so the UI calls the API on the same host and port. The context must include `pkg/` and `scripts/with-root-env.cjs`.
- **Rejected:** Two images (API plus nginx). Extra proxy and a baked `REACT_APP_API_URL` for no gain. Context limited to `morph/` fails the `replace` directives. Changing Go path defaults to `/data` would break `start-all.sh`.

### 2. Image env defaults, not Go defaults

- **Choice:** The final stage sets `PORT=9090`, `DB_PATH=/data/badger`, `TRAN_SQLITE_PATH=/data/tran.sqlite`, `ENTITY_DETAILS_BADGER=/data/entity_details`, `MORPH_KNOWLEDGE_DIR=/data/knowledge`, `TRAN_ENTITY_ATTACHMENT_DIR=/data/uploads/entity_attachments`, `GIN_MODE=release`. Compose mounts a named volume at `/data`.
- **Rationale:** SQLite `-wal` / `-shm` files sit next to `tran.sqlite`, so they stay on the volume. Uploads sit under `/data` instead of a second mount. Unset env in a local checkout still hits the Go defaults.
- **Rejected:** A second volume for `uploads/`. Operators would back up one tree and miss the other. A bind mount of the host's `./data` as the compose default ties the file to one machine layout; a named volume stays host-agnostic. Bind mounts remain possible later.

### 3. Entrypoint drops to uid 65532 after preparing `/data`

- **Choice:** Final stage is Alpine with `ca-certificates` and `su-exec` (BusyBox already provides `wget`). It creates user `morph` (uid/gid 65532). The Dockerfile does not set `USER`: the entrypoint must start as root, create the `/data` subdirectories, `chown` `/data` to `morph`, and `exec su-exec morph` on the binary. Working directory is `/app`, which is where `./frontend/build` is resolved. A negative secret test uses `docker compose run --rm --no-deps` (or `docker run`) so `restart:` cannot hide the exit.
- **Rationale:** A fresh named volume is root-owned and hides whatever ownership the image put on `/data`. Copying a `.keep` file into the volume does not make the directory writable by 65532. The server process must still be non-root. `wget` is there so `HEALTHCHECK` can call `/health` without a second Go flag.
- **Rejected:** `USER morph` with no entrypoint. Persistence fails on the first named volume. Running the server as root satisfies the volume and fails the non-root requirement. Distroless/scratch has no shell `wget`; adding a `-health` subcommand is Go code this story does not need. `curl` in a Debian runtime is a larger base for the same check.

### 4. No secrets in the build

- **Choice:** `.dockerignore` excludes `.env`, `.env.*`, `deploy/.env.production`, `node_modules`, `**/data`, other app trees, and git metadata. The Dockerfile has no `ARG` for secrets. The Node stage does not receive a repo `.env`, so `with-root-env.cjs` does not bake `REACT_APP_*`. Compose uses `env_file: .env.production` (gitignored). `deploy/.env.production.example` lists names with empty `JWT_SECRET` and `ADMIN_PASSWORD`, and does not set `JWT_EXPIRY_HOURS` (production default is 24; `876000` refuses to start).
- **Rationale:** CRA inlines `REACT_APP_*` at build time. A leaked root `.env` would put keys in the JS bundle even though the Go binary would not. Production startup already refuses default secrets; the example must not contain a value that would boot.
- **Rejected:** Build-args for `JWT_SECRET` or `MORPH_AI_API_KEY`. They land in image history. Committing a filled `.env.production`.

### 5. Compose layout and optional Caddy

- **Choice:** `deploy/docker-compose.yml` builds `context: ..`, `dockerfile: Dockerfile`, publishes `${MORPH_PUBLISH_PORT:-9090}:9090`, mounts named volume `morph-data` at `/data`, and reads `env_file: .env.production`. Caddy is service `caddy` with `profiles: ["tls"]`, image `caddy:2-alpine`, and `deploy/Caddyfile` reverse-proxying to `morph:9090` using `${MORPH_SITE_ADDRESS:-:80}`. `docker compose up` does not start it.
- **Rationale:** Issue #53 needs a place to finish TLS without this story picking a domain or deploying. The app stays host-agnostic.
- **Rejected:** Caddy in the default `up` (would bind 80/443 and look like a finished deploy). Hardcoding a hostname. A root `docker-compose.yml` that mixes future apps; `deploy/` is the tree the docs already expected.

### 6. Image build is a separate workflow

- **Choice:** `.github/workflows/docker-image.yml` runs on pull requests and pushes to `main`. It runs a small contract script, then `docker build`. It does not edit `.github/workflows/ci.yml`. Actions stay pinned to the same SHAs `ci.yml` already uses for checkout. No path filters.
- **Rationale:** `platform-ci` requires the platform workflow to report exactly five check names. A new job in that file is a sixth check. A docker step inside `Morph frontend` would still be that required check, so a daemon or Dockerfile failure would block merge under a name that means "npm test and build", and it would compile the UI twice. A separate workflow still runs every time. No path filter, because a `pkg/morphai` edit can break this image the same way it can break the Go modules.
- **Rejected:** New required check name. Path filters. Putting `docker build` inside an existing job.

### 7. Contract script is the small check

- **Choice:** `deploy/check-container-contract.sh` asserts the Dockerfile, ignore file, compose file, example env, and `ci.yml` job set. The image workflow runs it before `docker build`. It does not need a daemon.
- **Rationale:** The image build is slow. The script fails fast when a later edit drops `CGO_ENABLED=0`, adds a secret `ARG`, or adds a job to `ci.yml`. It is not a sixth required check.
- **Rejected:** A Go test that shells out to Docker. Nothing in `morph` changes, and CI's Go jobs must not require a daemon.

## Risks / Trade-offs

- [Fresh volume is root-owned] → entrypoint `chown` then `su-exec`. Verification must show the server uid is 65532 and that a note survives `down`/`up`. `chown -R` on every start is the ceiling for a large `/data` tree; a later change can skip it when the mount is already uid 65532.
- [Bind mount of a root-owned host directory] → same entrypoint `chown` covers it when the container starts as root. Document that overriding `user:` to 65532 skips the fix and the process cannot write.
- [Single writer] → one replica. SQLite and Badger are on one volume. The runbook says not to scale the service.
- [`products/` writes under the working directory, which is not on `/data`] → accepted. Those routes are outside this capability. `/app` is not writable by `morph`, so a product upload returns an error instead of writing into the image layer.
- [Health is process liveness, not a database probe] → existing `GET /health`. A bad volume still fails startup before listen, which fails the healthcheck.
- [Image build time on every PR] → accepted. No path filter, separate from the required checks so a slow build does not rename or occupy them.
- [Alpine `wget` and `PORT`] → `HEALTHCHECK` uses a shell so it reads `PORT` at runtime via `${PORT}` (`$$` is the shell PID in a Dockerfile, not a Compose escape). Container `PORT` stays 9090; the host publish port is `MORPH_PUBLISH_PORT`.
- [CRA `CI=true` promotes lint warnings] → the image build matches the frontend CI command. Do not set `NODE_ENV=production` before `npm ci` or devDependencies (`craco`) are omitted.

## Migration Plan

There is no running image to migrate. Operators copy `deploy/.env.production.example` to the gitignored `deploy/.env.production`, fill secrets, and `docker compose up`. Rollback is `docker compose down`. The named volume is left in place until an explicit `down -v`. Local `start-all.sh` is unchanged.

Upgrade is a new image and `docker compose up --build` with the same volume. Backup is a copy of the volume (or `docker run --rm` tar of `/data`) while the container is stopped, so SQLite is not mid-write.

## Open Questions

None. Host choice and real TLS certificates belong to later stories.
