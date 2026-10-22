## Context

Project (`dorph-engi/`) is a Rust Axum API on `DORPH_ENGI_PORT` or `PORT` (default 9096) and a Svelte UI on 5179. `GET /health` already returns HTTP 200. `STATIC_DIR` already serves the built UI with an `index.html` fallback. Uploads are hardcoded to a cwd-relative `uploads/` directory in five call sites, including `serve_upload`, which has no handler state. SQLite comes from `DORPH_ENGI_DATABASE_URL` (default `sqlite://dorph_engi.db`). `sqlx` links system `libsqlite3` (`bundled` is not enabled). `JWT_SECRET` defaults to `dev-dorph-engi-secret` when unset. `USERS_PANEL_BASE_URL` defaults to `http://127.0.0.1:9090`.

The root Blueprint on `main` lists `dorph` then `dorph-utils`. `deploy/check-container-contract.sh` already scopes Dorph env checks to the `dorph` service. `dorph-utils/deploy/check-container-contract.sh` rejects any other name. The root `.dockerignore` excludes `dorph-engi` so the Dorph image stays small. Dorph's shell-form healthcheck must use `${PORT}`; `$$` is the shell PID.

DorphUtils iframe module id is `projects`. The embed URL is `VITE_PROJECTS_URL`, falling back to `VITE_DORPH_ENGI_URL`. Story #114 sets that URL. This change must not.

The product owner applies `render.yaml` to Render project `prj-dahc33dbedkc73a1v8n0`. This repository does not call Render. Sibling stories may add Event Logs, Content Maker, and Data Access to the same file. Edits stay additive.

## Goals / Non-Goals

**Goals:**

- One Project image whose `/health` succeeds and whose secrets are environment-only.
- A `dorph-engi` service in `render.yaml` with a disk, pinned `PORT`, and dashboard prompts for the Dorph API base, `JWT_SECRET`, and `DORPH_AI_API_KEY`.
- Docs that name storage, the Dorph API base, and the placeholder `https://<dorph-engi public host>` for #114.
- Leave `dorph` and `dorph-utils` behaving as they do now, and let later services append.

**Non-Goals:**

- Creating or updating Render resources from this repo.
- Setting `VITE_PROJECTS_URL`, `VITE_DORPH_ENGI_URL`, or any other DorphUtils embed URL.
- Changing the iframe module id (`projects` already matches `dorph-utils/frontend/src/config.ts`).
- Event Logs, Content Maker, Data Access, AI tools, invite-signup, or the Dorph image contract.
- A new CI job name in `ci.yml`.

## Decisions

### One image, one origin

The API process serves `frontend/dist` when `STATIC_DIR` is set. The Dockerfile builds the UI with `VITE_API_BASE_URL` empty, so the browser calls the same host. That host is the iframe origin.

Rejected: a static site plus an API service. DorphUtils has one embed URL. Two origins would bake `VITE_API_BASE_URL` and add CORS. Rejected: folding Project into the Dorph image. The Dorph dockerignore and contract keep other apps out, and the products do not share a disk or port.

### Repo-root context, separate ignore file

`dorphai` is `../../pkg/dorphai-rs`, so the build context is the repo root and the Dockerfile is `dorph-engi/Dockerfile`. `dorph-engi/Dockerfile.dockerignore` is what that build uses. It excludes `.env`, `node_modules`, `target`, `dist`, and `dorph-engi/backend/uploads`. It does not exclude `dorph-engi/` or `pkg/`. The root `.dockerignore` keeps its `dorph-engi` line so the Dorph build is unchanged.

Rejected: `dockerContext: dorph-engi`. The crate path would point outside the context.

### Debian runtime, not Alpine

The binary links glibc and `libsqlite3`. The runtime image is `debian:bookworm-slim` with `libsqlite3-0`, `wget`, and `gosu`. The build stage installs `libsqlite3-dev` and `pkg-config`. No `USER` line: the entrypoint starts as root, creates `/data/uploads`, `chown`s `/data` to uid 65532 (`dorphengi`), then `exec gosu`.

Rejected: Alpine, which is what Dorph uses. A glibc binary will not start there, and enabling `bundled` plus a musl target is a larger change than installing `libsqlite3`.

### Listen on `PORT` only

`Settings::from_env` prefers `DORPH_ENGI_PORT` over `PORT`. The image and the Blueprint set `PORT=9096` and do not set `DORPH_ENGI_PORT`. The healthcheck is `wget` against `http://127.0.0.1:${PORT}/health`. `$$` is forbidden.

Local `./start-all.sh` still uses `DORPH_ENGI_PORT=9096` from the repo `.env`. That path is outside the image.

### Storage env, not a symlink

`DORPH_ENGI_UPLOAD_DIR` defaults to `uploads` when unset or blank, so local runs and the existing import test keep writing `uploads/<org>/`. The image and Blueprint set `/data/uploads`. `DORPH_ENGI_DATABASE_URL` in the image is `sqlite:///data/dorph_engi.db`. The Blueprint disk is `dorph-engi-data` at `/data`, 1 GB. Do not set `numInstances`; a disk is a single instance. Document that.

One helper reads the upload directory. All five `PathBuf::from("uploads")` sites use it. `serve_upload` can call the helper without handler state.

Rejected: a symlink from `/app/uploads` to `/data/uploads`. It depends on the process cwd staying `/app` and is invisible in the env docs the operator follows.

### Production refuses a dev secret and a loopback Dorph base

When `APP_ENV=production`, startup fails before listen if `JWT_SECRET` is empty, shorter than 32 characters, `dev-dorph-engi-secret`, or `dorph-dev-jwt-secret-change-me`; if the Dorph API base is empty or its host is `localhost`, `127.0.0.1`, `::1`, or `0.0.0.0`; or if the database file or upload directory is not an absolute path under `/data`. Other `APP_ENV` values keep today's defaults. The check is a pure function of `Settings` so tests do not mutate process env.

`JWT_SECRET` and `DORPH_AI_API_KEY` are Blueprint prompts (`sync: false`, no value). `USERS_PANEL_BASE_URL` is the same kind of prompt: the public Dorph origin, not a secret, and not invented here. The operator pastes the same `JWT_SECRET` Dorph already uses.

Rejected: `fromService` to copy Dorph's `JWT_SECRET`. Dashboard secrets are not a value we can prove the Blueprint reference will read, and a prompt matches the Event Logs pattern. Rejected: baking the dev secret into the image so the container boots. That is the footgun the security checklist already names.

### Additive Blueprint

Service order stays `dorph`, `dorph-utils`, then `dorph-engi`. The DorphUtils contract changes from an exclusive pair to "the list starts with `dorph` then `dorph-utils`". The Project contract requires `dorph-engi` and forbids setting `VITE_PROJECTS_URL` and `VITE_DORPH_ENGI_URL` anywhere in the file. It does not require those to be the only three names, so a sibling service can land without this check going red.

Rejected: a top-level `projects:` block that names `prj-dahc33dbedkc73a1v8n0`. The Dorph contract forbids that key. The owner attaches this file to the existing project. Rejected: an exclusive three-name list. Open Blueprint PRs for Event Logs and Content Maker already assume they are the third name; an exclusive list makes the next merge fail the other's check.

The old DorphUtils scenario titled "No other product services are declared" is renamed. Keeping that title while allowing `dorph-engi` would leave a false heading in the main spec.

### Docs and the #114 placeholder

`dorph-engi/README.md` and `deploy/README.md` list `PORT`, `APP_ENV`, `STATIC_DIR`, `DORPH_ENGI_DATABASE_URL`, `DORPH_ENGI_UPLOAD_DIR`, `USERS_PANEL_BASE_URL`, `JWT_SECRET`, and `DORPH_AI_API_KEY`. Empty example values only. Both name `https://<dorph-engi public host>` as the value #114 will put in `VITE_PROJECTS_URL` / `VITE_DORPH_ENGI_URL`. `dorph-utils/README.md` gains that same placeholder sentence. Neither README nor `render.yaml` sets those variables.

### Checks

`dorph-engi/deploy/check-container-contract.sh` locks the Dockerfile, ignore file, example env, and Blueprint shape. `.github/workflows/dorph-engi-image.yml` runs that script, schema-validates `render.yaml`, and builds the image. It is not a job in `ci.yml`. `cargo test` covers the production guard and the upload-dir default.

## Risks / Trade-offs

- [Render injects its own `PORT` if the Blueprint omits it, and a set `DORPH_ENGI_PORT` would win] → Pin `PORT=9096` in the image and the Blueprint, and do not set `DORPH_ENGI_PORT` in either.
- [`Dockerfile.dockerignore` is ignored by an old builder, so the root ignore drops `dorph-engi` and `COPY` fails] → The image build is the proof. The contract also requires the ignore file to keep `pkg/` and `dorph-engi/` and to drop `uploads` and `.env`.
- [A disk mounts as root, so the server cannot create the SQLite file] → Entrypoint `chown` before `gosu`. Docs say one instance and to snapshot the disk before an upgrade.
- [Unset `JWT_SECRET` used to boot with `dev-dorph-engi-secret`] → Production now exits. Local development does not set `APP_ENV=production`.
- [Sibling PRs conflict on `render.yaml`] → Only append a service and relax the exclusive name check to a prefix. Do not rewrite `dorph` or `dorph-utils` env.
- [Health body is `status: ok`, while Dorph's runbook says `healthy`] → `wget` and Render only require HTTP 200. Leave the JSON alone.
- [`APP_ENV=development` in the shared `.env.example` is what local Project reads] → The production guard keys off `production` only, so that shared flag does not change local startup.

## Migration Plan

After merge, the product owner syncs the existing Blueprint in project `prj-dahc33dbedkc73a1v8n0`. Render creates the `dorph-engi` service from the new block. Before the first boot they set `USERS_PANEL_BASE_URL` to `https://<dorph public host>` (no path), `JWT_SECRET` to Dorph's secret, and `DORPH_AI_API_KEY` if AI documents should run. Health does not need the AI key. They copy `https://<dorph-engi public host>` for story #114. They do not set the DorphUtils variables in this step.

Rollback is reverting the Blueprint service block. Do not destroy the disk from this repo. Local `start-all.sh` is unchanged.

## Open Questions

None. The module id, the port, and the decision not to call Render are fixed by the story and the code.
