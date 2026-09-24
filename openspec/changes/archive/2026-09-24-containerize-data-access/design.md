## Context

See proposal.md. Data Access is `SharpReport/backend` (Axum, package `datapulse`, edition 2024) plus a SvelteKit UI on Vite port 5178. The API listens on `SHARPREPORT_PORT` (default 3050) via `Settings::new`. It does not read `PORT`. `GET /api/v1/health` returns the text `OK`. There is no `GET /health` or `GET /ready`. `GET /` in a debug build redirects to Vite. The release fallback is a stub HTML page, not the built UI.

`resolve_database_url` strips a leading `/` before the absolute-path check, so `sqlite:///data/datapulse.db` is treated as a relative path. `config/production.toml` points at Postgres and contains placeholder passwords. `RUN_ENV` defaults to `development`, which loads `config/development.toml`. Metabase `autostart` is false.

`Cargo.toml` path-depends on `pkg/morphai-rs`. The root `.dockerignore` excludes the whole `SharpReport` tree, and the Morph contract still requires the word `SharpReport` in that file. `morph-utils/deploy/check-container-contract.sh` requires `render.yaml` service names to be exactly `morph`, `morph-utils`.

`SharpReport/deploy/Dockerfile` copies `frontend/build` (not what SvelteKit emits here), health-checks a hardcoded port 3050, and the compose file sets `JWT_SECRET=change-me-in-production` plus a Postgres password.

Morph and MorphUtils are already in `render.yaml`. Render injects `PORT` unless the Blueprint pins it. Shell-form `HEALTHCHECK` is `/bin/sh -c`, so `$$PORT` is the PID. SQLx `sqlite` links libsqlite3 dynamically. The Rust image must be glibc if the runtime is Debian.

## Goals / Non-Goals

**Goals:**

- One local command starts Data Access and `GET /health` and `GET /ready` succeed without Morph.
- The Blueprint adds `sharpreport` without removing `morph` or `morph-utils`.
- Secrets and the Morph auth base URL are environment prompts with no committed values.
- The public URL placeholder for #114 is written down. `VITE_DATAX_URL` is not set.

**Non-Goals:**

- Creating the Render service, calling the Render API, or guessing an `onrender.com` host.
- Setting `VITE_DATAX_URL` or `REACT_APP_MORPH_UTILS_URL`.
- Packaging Metabase or Java.
- Changing Event Logs, Content Maker, Project, AI tools, or Invite Signup application code.
- Changing `start-all.sh` ports (API 3050, UI 5178).

## Decisions

### 1. One process serves the API and the production UI

- **Choice:** Multi-stage image. Node 22 builds the SvelteKit app with `@sveltejs/adapter-static` and `fallback: 'index.html'`. Rust 1.89 builds `datapulse` in release. Debian bookworm runs that binary only. The image sets `SHARPREPORT_UI_DIR` to the built files. The API serves `index.html` for `GET /` and for other GET paths that are not files and do not start with `/api`, `/public`, `/metabase`, `/health`, or `/ready`. `PUBLIC_API_URL` stays unset so the UI calls same-origin `/api`.
- **Rationale:** MorphUtils will iframe one origin. The UI already uses relative `/api` when that variable is empty. One listen port is one Render health check.
- **Rejected:** Two Render services (API 3050, nginx UI 5178). The iframe needs the UI origin, and the UI would need the public API origin. A second host and a second disk decision. #114 would have to be right twice.
- **Rejected:** nginx in front of the API in one container. Two processes. The shell uses nginx because it has no API. This app's health is the API.
- **Rejected:** Keep the stub HTML and only containerize the API. The embed story needs the UI on that origin.

### 2. `SHARPREPORT_PORT` wins over `PORT`

- **Choice:** Listen on `SHARPREPORT_PORT` when it parses as a non-zero `u16`, else `PORT`, else the file default `3050`. The image and the Blueprint set both to `3050`. The healthcheck calls `http://127.0.0.1:${PORT}/health`.
- **Rationale:** Render injects `PORT` unless it is pinned. The process ignores `PORT` today, so an injected port would miss 3050. Preferring `PORT` over `SHARPREPORT_PORT` would bind a checkout that loaded the root `.env` (`PORT=9090`) to Morph's port. `.env.example` already sets `SHARPREPORT_PORT=3050`, and `start-all.sh` launches the API with that variable, so the supported local path stays on 3050. A shell that exports only `PORT=9090` and then runs this binary will listen on 9090; that is the Render fallback, not the launcher. The image publishes 3050, not Vite's 5178.
- **Rejected:** Pin `PORT` in the Blueprint and leave the binary on `SHARPREPORT_PORT` only. Dropping `SHARPREPORT_PORT` later would listen on 3050 while Render follows an injected `PORT`.

### 3. Build context is the repo root

- **Choice:** `dockerfilePath: ./SharpReport/Dockerfile`, `dockerContext: .`. Stop excluding the whole `SharpReport` tree. Keep a `SharpReport/...` line so the Morph contract still sees the word. Do not `COPY` SharpReport from the root Morph Dockerfile. Copy `pkg/morphai-rs` and `SharpReport/` only.
- **Rationale:** `morphai = { path = "../../pkg/morphai-rs" }` from `SharpReport/backend`. A `SharpReport/` context cannot see that crate.
- **Rejected:** Vendor `morphai-rs` into `SharpReport/`. Duplicates the shared crate.

### 4. SQLite on `/data`, no Metabase in the image

- **Choice:** `SHARPREPORT_DATABASE_URL=sqlite:///data/datapulse.db`. Fix `resolve_database_url` so a path that is absolute after the `sqlite://` prefix is returned unchanged. Entrypoint starts as root, `chown`s `/data` to uid 65532 (`sharpreport`), then execs the binary. Blueprint disk `sharpreport-data` at `/data`, 1 GB, `maxShutdownDelaySeconds: 120`, no `numInstances`. Do not set `RUN_ENV=production` (that file selects Postgres and placeholder passwords). Metabase stays `autostart = false`. No JRE.
- **Rationale:** Imported tables live in this SQLite file. A Render disk is one instance, same as Morph. The chown exists because Render disks mount as root. Java plus a Metabase download would make the first boot slow and would put another database on the same disk without a story asking for it.
- **Rejected:** Ship the old compose stack (app + Postgres + Metabase). It commits a JWT and a database password, and it is not the app MorphUtils embeds.
- **Rejected:** No disk. Recreate would wipe tables.

### 5. Health is local; the Morph origin is a prompt

- **Choice:** `GET /health` and `GET /ready` return `{"status":"ok"}` and do not call `USERS_PANEL_BASE_URL`. Keep `GET /api/v1/health`. `USERS_PANEL_BASE_URL`, `JWT_SECRET`, and `MORPH_AI_API_KEY` are `sync: false` with no `value`. No secret `ARG`. An empty JWT still lets the process listen; the development toml secret remains the local default until the dashboard sets `JWT_SECRET`. Sign-in fails until `USERS_PANEL_BASE_URL` is `https://<morph public host>`.
- **Rationale:** `/health` must pass in a local container before Morph exists. `sync: false` is how this Blueprint prompts without committing a host. `fromService` would yield a hostname, not `https://…`.
- **Rejected:** Fail startup when `JWT_SECRET` is empty. That breaks the local health check the story requires. Document the key instead.
- **Rejected:** Copy Morph's full secret list. This API reads `JWT_SECRET` (its own crypto key), `USERS_PANEL_BASE_URL`, and `MORPH_AI_API_KEY`. Listing unused keys makes them look required.

### 6. Contracts stay additive

- **Choice:** `SharpReport/deploy/check-container-contract.sh` checks this image and the `sharpreport` service block. `morph-utils/deploy/check-container-contract.sh` requires `morph` and `morph-utils` first and requires `sharpreport` to be present. It does not require the list to be only those three. No new job in `ci.yml`. The existing image workflow still schema-validates `render.yaml`. Add one step in that same job to run the Data Access contract. The job name stays `Build image`.
- **Rationale:** Sibling PRs for Event Logs and Content Maker also append services. An exact three-name list would fight those diffs. A "must include" check still fails if `sharpreport` is dropped.
- **Rejected:** Leave the MorphUtils contract at exact `["morph", "morph-utils"]`. It would fail as soon as this service is added.

### 7. Docs name the placeholder and do not set the embed

- **Choice:** `deploy/README.md` tells the product owner to sync the Blueprint in project `prj-dahc33dbedkc73a1v8n0`, set `USERS_PANEL_BASE_URL` to `https://<morph public host>`, and copy `https://<sharpreport public host>` for story #114. This change does not set `VITE_DATAX_URL`. MorphUtils README keeps that variable unset and points at the same placeholder.
- **Rationale:** #114 wires the shell. Inventing an `onrender.com` host would be wrong.

## Risks / Trade-offs

- [Login fails until `USERS_PANEL_BASE_URL` is the public Morph origin] → `/health` still returns 200. The runbook says to set it before expecting sign-in. The in-container default `127.0.0.1:9090` is not Morph.
- [Empty `JWT_SECRET` uses the development toml secret] → the process starts. The runbook says to set `JWT_SECRET` before storing database connection passwords. The example file leaves it empty.
- [Disk blocks zero-downtime deploys] → one instance, same as Morph. Document snapshot backup.
- [`chown -R` on every start] → fine for a 1 GB disk. `ponytail:` the full walk; upgrade path is to skip it when `/data` is already uid 65532.
- [Absolute SQLite URLs were rewritten] → a unit test locks `sqlite:///data/datapulse.db`.
- [`$$PORT` in `HEALTHCHECK`] → use `${PORT}`. The contract rejects `$$`.
- [Alpine runtime with a glibc binary] → the binary would not start. Runtime is Debian with `libsqlite3-0`.
- [SPA fallback returns HTML for unknown API paths] → fallback skips `/api`, `/public`, and `/metabase`.
- [Sibling Blueprint edits] → service block is appended. Re-merge `main` before the PR if Event Logs or Content Maker land, and keep their services.
- [Morph build context grows] → `**/node_modules` and `**/target` stay ignored. The Morph Dockerfile still does not copy SharpReport.

## Migration Plan

1. Merge latest `main` before opening the PR. If `formx` or `composerx` services are present, keep them and keep `sharpreport` in the list. Do not force-push.
2. Product owner syncs the Blueprint in project `prj-dahc33dbedkc73a1v8n0` and fills `USERS_PANEL_BASE_URL`, `JWT_SECRET`, and optionally `MORPH_AI_API_KEY`.
3. Rollback is the previous image on the same disk. Deleting the service deletes the disk. No data migration from a checkout.

## Open Questions

None. Metabase in a later image is out of scope and does not change these tasks.
