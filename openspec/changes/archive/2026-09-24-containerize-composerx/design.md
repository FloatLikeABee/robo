## Context

See proposal.md for why. Behavior is in `specs/composerx-container/spec.md`, `specs/composerx-render/spec.md`, and the modified `morph-utils-render` requirements.

Verified against this branch after fast-forward to `origin/main` (`b284610`, MorphUtils image #116 and Blueprint #117). Issue #109 is still open and is not in `main`.

- `composerx/backend` is Go + Gin. `go.mod` is `go 1.25.6` and `replace`s `pkg/morphai`, `pkg/assistmd`, `pkg/webresearch`, and `pkg/repoenv`. SQLite is `modernc.org/sqlite` (no CGO). Schema SQL is embedded. `main` listens with `getEnv("COMPOSERX_PORT", "8043")` and `router.Run(":" + port)`, which is all interfaces. It does **not** read `PORT` today. The backend README already claims `PORT` / `COMPOSERX_PORT`.
- Root `.env.example` sets both `PORT=9090` (Morph) and `COMPOSERX_PORT=8043`. `repoenv.Load` does not override existing process env and looks for `start-all.sh`. A container has neither file.
- `GET /health` returns `{"status":"ok"}` and the auth middleware already skips `/health`, `/auth/`, and `/public/`. It does not call `USERS_PANEL_BASE_URL`. Login does. The default base URL in code is `http://127.0.0.1:9090`.
- There is no route for `GET /`. Pages are in-app state in `App.svelte`, not a path router. Public HTML is `/public/p/:slug` on the API.
- The UI production fallback is `import.meta.env.VITE_API_BASE ?? (import.meta.env.DEV ? '' : 'http://localhost:8043')`. An unset variable becomes the localhost string in `vite build`. An empty string is kept (`??` does not skip it). `fetch` sends `Authorization` and does not set `credentials: 'include'`. Dev proxy targets `127.0.0.1:8043`. `vite.config.js` sets `envDir` to the repo root.
- Keys the API actually reads: `MORPH_AI_API_KEY`, legacy `TRAN_QWEN_API_KEY`, `MORPH_AI_BASE_URL` / `MORPH_AI_MODEL`, `TRAN_OPENAI_API_KEY`. It does not read `JWT_SECRET`, `SMTP_PASS`, or the rest of the Morph secret list. `ai.config.json` is gitignored.
- `render.yaml` services are `morph` then `morph-utils`. `deploy/check-container-contract.sh` scopes Morph env to the `morph` block. It still rejects a file-wide `projects:`, `numInstances:`, or `generateValue:`. `morph-utils/deploy/check-container-contract.sh` requires the service names to be exactly `morph`, `morph-utils`.
- Root `.dockerignore` drops the `composerx` directory. Docker uses that one ignore file for a context of `.`. The Morph Dockerfile copies named paths and does not copy `composerx`. The Morph image workflow validates the Blueprint schema and builds only the Morph image. `ci.yml` has five check names.
- Shell-form `HEALTHCHECK` is `/bin/sh -c`. `$$PORT` is the PID. Alpine's `wget` is the client the Morph image uses. `nginx:1.27-alpine` has `curl` instead; this image is not that base.

## Goals / Non-Goals

**Goals:**

- One Content Maker image whose `/health` passes with Morph and MorphUtils absent, and a Blueprint service the product owner can create.
- Local `./start-all.sh start composerx` unchanged: API 8043, Vite 8044, `./data` and `./storage`.

**Non-Goals:**

- Calling Render, or setting `VITE_COMPOSERX_URL` / `REACT_APP_MORPH_UTILS_URL`.
- Event Logs, Data Access, Project, AI tools, Invite Signup, or the Morph header chip.
- A new `ci.yml` check, or building this image inside `.github/workflows/docker-image.yml`.
- Editing Morph CORS. Same-origin UI fetches do not need it.

## Decisions

### 1. One image, one Go process, UI and API on 8043

- **Choice:** Multi-stage image. Node 22 builds `composerx/frontend` with `VITE_API_BASE` empty. Go 1.25 builds `composerx-server` with `CGO_ENABLED=0`. Alpine runs that binary only. The image sets `PORT`, `COMPOSERX_PORT`, and `COMPOSERX_UI_DIR`. Gin serves `/assets` and the files that exist next to `index.html`, plus `GET /` as that file. No catch-all.
- **Rationale:** MorphUtils will iframe one origin. Published pages already live on the API. One listen port means one Render health check. Empty `VITE_API_BASE` keeps same-origin calls because `??` keeps `""`. The dev ternary still bakes `http://localhost:8043` for a checkout `vite build` that does not set the variable.
- **Rejected:** Two Render services (API 8043, nginx UI 8044). The iframe needs the UI origin, and the UI would need the public API origin at build or start. `fetch` is not credentialed and CORS already sends `*` off loopback, so a split would function, but it adds a second host, a second disk decision, and a URL #114 would have to get right twice. Published HTML would not match the UI origin.
- **Rejected:** nginx in front of Go in one container. Two processes, and a supervisor that must forward signals and fail health when Go dies. The shell uses nginx because it has no API. This app's health is the API.
- **Rejected:** `vite preview` as the server. It does not proxy in this config, it is a second process, and it still cannot read env after the build.
- **Rejected:** Serving the UI from the root Morph Dockerfile. That image is specified to exclude Content Maker. A catch-all `NoRoute` to `index.html` would hide API 404s and can swallow `/public/p/:slug` if registered too broadly.

### 2. `COMPOSERX_PORT` wins over `PORT`

- **Choice:** Listen on `COMPOSERX_PORT` when non-empty, else `PORT`, else `8043`. The image and the Blueprint set both to `8043`. Healthcheck calls `http://127.0.0.1:${PORT}/health` with `${PORT}`, not `$$PORT`.
- **Rationale:** Render injects `PORT` unless the Blueprint pins it. The process today ignores `PORT`, so an injected port would make Render's proxy miss 8043. Preferring `PORT` over `COMPOSERX_PORT` would bind a local API to Morph's `9090` from the root `.env`. The image publishes 8043, not the Vite port 8044. 8044 stays the dev server. #114 uses `https://<composerx public host>`, not `:8044`.
- **Rejected:** Leave the binary on `COMPOSERX_PORT` only and pin that variable without teaching it `PORT`. A future edit that drops `COMPOSERX_PORT` from the service would listen on 8043 while Render follows an injected `PORT`. Reading `PORT` as the fallback matches the Morph image and the backend README.

### 3. Root build context, Morph image still omits this app

- **Choice:** `dockerfilePath: ./composerx/Dockerfile`, `dockerContext: .`. Root `.dockerignore` stops excluding the whole `composerx` tree and keeps excluding `composerx/backend/storage`, `composerx/frontend/node_modules`, env files, and `ai.config.json`. The ignore file still contains the word `composerx`. The Morph contract additionally fails if the root Dockerfile copies `composerx`. The Content Maker Dockerfile copies `pkg/` and `composerx/` only.
- **Rationale:** `replace` directives point at `../../pkg` from `composerx/backend`. A context of `composerx/` cannot see `pkg/`. Render supplies one context and the ignore file at its root. The Morph image stays free of Content Maker because its Dockerfile never copies that tree. Sample `storage/` files are not the runtime store.
- **Rejected:** Context `composerx/` plus a committed vendor of `pkg`. Duplicate modules, and they drift.
- **Rejected:** Deleting the `composerx` line from `.dockerignore` with no remaining token. `deploy/check-container-contract.sh` requires that word. Fully excluding the directory makes `COPY composerx` fail.

### 4. UI routes are public; everything else stays gated

- **Choice:** When the UI directory is mounted, skip auth for `GET /`, `GET /index.html`, `/assets/`, and the other files mounted from that directory's root. Do not skip `/templates`, `/emails`, `/ai`, or `/public` beyond the skips that already exist.
- **Rationale:** The login screen is inside `index.html`. A global middleware in front of the route returns 401 for `/` today, which would be a blank iframe. Assets are not API routes.
- **Rejected:** Making `/` public even when no UI is mounted. Harmless 404, but the skip list should match files the process actually serves so a later API route at a static name is not accidentally open. Mounting and the skip list use the same directory listing.

### 5. Disk, non-root server, env-only secrets

- **Choice:** Entrypoint starts as root, creates `/data`, `chown`s it to uid 65532 (`composerx`), and `exec`s `su-exec`. Image defaults: SQLite `/data/composerx.sqlite`, Badger `/data/composerx_badger`, files `/data/storage`. Blueprint disk `composerx-data` at `/data`, 1 GB, `maxShutdownDelaySeconds: 120`, no `numInstances`. `GIN_MODE=release`. `USERS_PANEL_BASE_URL`, `MORPH_AI_API_KEY`, `TRAN_QWEN_API_KEY`, and `TRAN_OPENAI_API_KEY` are `sync: false` with no `value`. No secret `ARG`. Compose for local runs lives at `composerx/docker-compose.yml` and does not edit `deploy/docker-compose.yml`.
- **Rationale:** SQLite is `SetMaxOpenConns(1)` and Badger is a single writer. A Render disk is one instance and not zero-downtime, same as Morph. The chown exists because Render disks mount as root. The API does not read Morph's JWT or SMTP variables; listing them would look required. `USERS_PANEL_BASE_URL` is a public origin, not a secret, and `sync: false` is how this Blueprint prompts without committing a host. `fromService` `host` is a hostname, not `https://…`.
- **Rejected:** No disk. Recreate would wipe templates and published pages.
- **Rejected:** `USER` before the entrypoint. The process could not chown a root-owned mount.
- **Rejected:** Copying the Morph secret inventory onto this service. Those keys are unused here.
- **Rejected:** A guessed `onrender.com` host, or `value: ""`, or `https://morph.example` in the Blueprint.

### 6. Contracts stay beside the image; schema check stays where it is

- **Choice:** `composerx/deploy/check-container-contract.sh` checks the Dockerfile (`CGO_ENABLED=0`, `/health`, `${PORT}`, no `$$`, no secret `ARG`), the ignore exceptions, the example env, the runbook strings, and the `composerx` service block. `morph-utils/deploy/check-container-contract.sh` requires names `morph`, `morph-utils`, `composerx` in that order and still checks the shell service itself. No edit to `ci.yml` or `docker-image.yml`. Schema validation of the whole `render.yaml` remains that workflow's existing step.
- **Rationale:** A Content Maker image job inside the Morph image workflow would fail the check that gates `morph`. The five platform check names stay. #109 will have to extend the same name list; that conflict is a few lines and is why this change stays additive and must merge `main` again before the PR if #109 lands first.
- **Rejected:** A sixth required check in `ci.yml`.
- **Rejected:** Allowing any extra service name. A typo would pass. The list is exactly these three until the next satellite story edits it.

### Grill

Proposer: one Go process on 8043, root context, prompted Morph origin, placeholder for #114. Reviewer challenges:

- **Assumption:** "The server already honors `PORT`." The README says so. `main` reads only `COMPOSERX_PORT`. Shipping a pin of `PORT=8043` without the fallback leaves Render and the process agreed only while `COMPOSERX_PORT` is also 8043. The fallback is in the binary so the pin is real. `COMPOSERX_PORT` still wins so the root `.env` `PORT=9090` does not move the local API.
- **Assumption:** "Health must reach Morph or the iframe is useless." `handleHealth` writes JSON and returns. Login is the call to `/api/auth/login`. Health stays local. A down Morph fails sign-in and still passes `/health`.
- **Assumption:** "The container should listen on 8044 because MorphUtils defaults `VITE_COMPOSERX_URL` to `http://localhost:8044`." That default is the Vite dev server. The image is the production origin, one port, and #114 replaces the whole URL. Publishing 8044 as well would be a second Render port this Blueprint cannot express on one service.
- **Assumption:** "CORS must be opened for the MorphUtils origin." The document is loaded by the iframe. Scripts then call the iframe's own origin when `VITE_API_BASE` is empty. No browser CORS. The existing `*` path is for a split we rejected. Do not edit CORS.
- **Assumption:** "Unset `VITE_API_BASE` is same-origin in production." False. The `DEV` ternary keeps `http://localhost:8043` in the production bundle. The image build sets the variable to empty. A checkout build is unchanged.
- **Assumption:** "`fromService.host` can fill `USERS_PANEL_BASE_URL`." It is a hostname. The client joins it to `/api/auth/login` and needs a scheme. Dashboard prompt only.
- **Assumption:** "Excluding `composerx` in `.dockerignore` is compatible with a root context." `COPY composerx` would be empty. The Morph contract only greps for the word. Narrow the ignore, and forbid `COPY` of `composerx` in the root Dockerfile so the Morph image spec still holds.
- **Assumption:** "A static site is enough because the UI is Vite." The API owns SQLite, Badger, and `/health`. `runtime: static` would not run the entrypoint or the binary.
- **Assumption:** "Optional `VITE_COMPOSERX_URL` belongs on `morph-utils` now so the PO cannot forget it." The shell service was specified to leave embed prompts unset until those origins exist. Setting it here is #114. Document the placeholder. Do not add the key.
- **Failure:** `$$PORT` in `HEALTHCHECK` requests `<pid>PORT`. Use `${PORT}`.
- **Failure:** auth middleware returns 401 for `index.html` and `/assets/*`. The login page never loads. Skip only the mounted UI paths.
- **Failure:** `NoRoute` → `index.html` returns the UI for `GET /templates/missing` and can shadow `/public/p/:slug`. Mount explicit files only.
- **Failure:** a second file-wide `PORT` parser. Morph's parser is already scoped to the `morph` block. Do not widen it. Do not add `numInstances`.
- **Failure:** `chown` skipped. The first Render disk is root-owned and SQLite fails after a healthy-looking start. The entrypoint chowns before exec. `ponytail:` the full walk; upgrade path is to skip it when `/data` is already uid 65532.
- **Failure:** example env or Blueprint contains a real key or `https://something.onrender.com`. Empty example. `sync: false` and no `value`.
- **Out of scope held:** no Render API, no `VITE_COMPOSERX_URL` value, no formx service, no Morph chip, no new CI check name.

## Risks / Trade-offs

- [Login fails until `USERS_PANEL_BASE_URL` is the public Morph origin] → `/health` still returns 200. The runbook says to set it before expecting sign-in. The in-container default `127.0.0.1:9090` is not Morph.
- [UI calls localhost if someone builds the frontend without empty `VITE_API_BASE`] → the image build sets it. The contract checks that the Dockerfile does. Checkout `npm run build` stays the dev-oriented default.
- [Disk blocks zero-downtime deploys] → one instance, same as Morph. Document snapshot backup. Rollback is the previous image on the same disk, or deleting the service. No data migration from a checkout.
- [GitHub checks do not build this image] → the contract and the existing schema step cover the Blueprint. `docker build` is the local proof. A broken Dockerfile fails the product owner's first Render build, not `ci.yml`.
- [`checksPass` does not compile this image] → same trigger as Morph and MorphUtils. A red Content Maker Dockerfile does not block the `morph` deploy.
- [Prompted `USERS_PANEL_BASE_URL` uses the secret-shaped `sync: false`] → the runbook says it is the public Morph origin.
- [Exact service-name list fights #109] → merge `main` again before the PR if that story lands. Extend the list; do not force-push.
- [Shared root context is larger for the Morph image build] → Content Maker source is in the context and not in the Morph image. The new contract line keeps it that way.
- [`chown -R` on every start] → fine for a 1 GB disk. The comment names the ceiling.

## Migration Plan

There is no Content Maker service yet. After merge to `main`, the product owner syncs the Blueprint in project `prj-dahc33dbedkc73a1v8n0`. That adds `composerx` beside `morph` and `morph-utils`. They set `USERS_PANEL_BASE_URL` to `https://<morph public host>` and copy `https://<composerx public host>` for #114. Rollback is deleting that service in the dashboard. Local `start-all.sh` and `npm run dev` stay as they are.

If #109 merges first, rebase or merge `main` and keep both new services. Do not drop `morph` or `morph-utils`. Do not force-push.

## Open Questions

None. Embed wiring is #114. Event Logs is #109.
