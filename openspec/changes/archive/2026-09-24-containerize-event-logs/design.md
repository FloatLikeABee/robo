## Context

See proposal.md for why. Behavior is in `specs/formx-container/spec.md`, `specs/formx-render/spec.md`, and the modified `morph-utils-render` requirements.

Verified against `origin/main` (MorphUtils image `c4ead85`, Blueprint `b284610`):

- `formx/backend` is Gin. `go.mod` is `go 1.25.0`. SQLite is `github.com/glebarez/sqlite` (modernc). Badger is v4. No `cgo` import. `replace` directives point at `../../pkg/{morphai,assistmd,docextract,webresearch,repoenv}`. Those modules have no further `replace`. A build context of only `formx/` cannot resolve them.
- `cmd/server/main.go` listens on `:` + `SERVER_PORT` (default `29909`), which is all interfaces. It serves `/uploads` and `/swagger`. It does not serve the UI and has no `/health`. `repoenv.Load` reads the repo-root `.env` and does not override variables already set. That file sets `PORT=9090` for Morph and `SERVER_PORT=29909` for Event Logs. `start-all.sh` publishes formx on `SERVER_PORT`.
- Login and `/api/v1/auth/me` proxy to `USERS_PANEL_BASE_URL` (default `http://127.0.0.1:9090`) at `/api/auth/login` and `/api/auth/user`. The process starts without that host. There is no local JWT check and no production secret refusal in this binary.
- The UI is Vite. `VITE_API_URL ?? ''` is same-origin when unset. `vite.config.ts` sets `envDir` to the repo root and proxies `/api` to `localhost:29909` only for `server` and `preview`. `npm run build` is `tsc -b && vite build` into `frontend/dist`.
- MorphUtils `config.ts` sets the Event Logs iframe to `${VITE_SHEETX_URL || VITE_FORMSX_URL || 'http://localhost:19909'}/events-info`. One origin, path `/events-info`.
- The root `Dockerfile` is Morph only. Root `.dockerignore` has a `formx` line, and `deploy/check-container-contract.sh` requires that substring. On this main the Morph env parser is scoped to the service named `morph`, so a second `PORT` does not replace `9090`. It still rejects a file-wide `numInstances` or `generateValue`. `morph-utils/deploy/check-container-contract.sh` requires the service names to be exactly `morph` then `morph-utils`.
- BuildKit uses `<Dockerfile>.dockerignore` instead of the context `.dockerignore` when that file exists (`docker build -f formx/Dockerfile .` looks for `formx/Dockerfile.dockerignore`).
- `ci.yml` has five required check names. `.github/workflows/docker-image.yml` builds only the Morph image and validates `render.yaml`.

## Goals / Non-Goals

**Goals:**

- One Event Logs image, optional compose, `/health`, `/data`, env-only secrets.
- One `formx` service in the existing Blueprint, with a runbook and the #114 placeholder.
- Local `npm run dev` and `start-all.sh` keep the Vite proxy and cwd-relative paths.

**Non-Goals:**

- Calling Render or creating the service.
- Setting `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, or `REACT_APP_MORPH_UTILS_URL`.
- Images or services for Content Maker, Data Access, Project, AI tools, or Invite Signup.
- Editing Morph CORS, the Morph SPA chip, or `ci.yml` job names.
- A database probe inside `/health`. Startup already exits if SQLite or Badger cannot open.

## Decisions

### 1. One image, Go serves the Vite build

- **Choice:** Multi-stage `formx/Dockerfile`. Node 22 builds `formx/frontend` with `VITE_API_URL` empty. Go 1.25 builds `CGO_ENABLED=0` with `-trimpath -ldflags "-s -w"`. Alpine runs the binary and copies `dist` to `/app/frontend/dist`. The process serves that tree when `index.html` is there, and falls back to it for client routes. `/api/`, `/uploads`, `/swagger`, and `/health` are never the HTML shell. If `dist` is missing (local `go run`), the API still listens and the SPA is not mounted.
- **Rationale:** MorphUtils iframes one origin plus `/events-info`. Same-origin `VITE_API_URL` means the UI does not need a second host. This is the Morph image shape (API + UI, one port).
- **Rejected:** API and nginx UI as two services. The iframe origin would be the UI, and the UI would need a baked API URL or a proxy. Two disks or a private API hop. UI `/health` could pass while the API is down.
- **Rejected:** `vite preview` as the server. Its proxy targets `127.0.0.1:29909` inside the UI container, and it does not open SQLite.
- **Rejected:** Adding Event Logs to the root Morph Dockerfile. `morph-container` says that image must not contain Event Logs, and the two apps do not share a disk.

### 2. Repo-root context with a Dockerfile-specific ignore file

- **Choice:** `dockerContext: .` and `dockerfilePath: ./formx/Dockerfile`. `formx/Dockerfile.dockerignore` keeps `formx/` and `pkg/` and drops `.env`, `node_modules`, `data`, and the other apps. The root `.dockerignore` still has the `formx` line so the Morph context stays unchanged.
- **Rationale:** `replace` paths live in `pkg/`. The root ignore file cannot both exclude `formx` from Morph and include it for this image. BuildKit's per-Dockerfile ignore file is the split.
- **Rejected:** Context `formx/` only. `go mod download` cannot see `../../pkg`.
- **Rejected:** Deleting the `formx` line from the root ignore file. That uploads the Event Logs tree into every Morph build. The contract wants that line.

### 3. `PORT` is pinned; the process still listens on `SERVER_PORT`

- **Choice:** Image `ENV PORT=29909` and `SERVER_PORT=29909`. The entrypoint sets `SERVER_PORT` from `PORT` (default 29909) after checking that `PORT` is numeric, then execs the binary. Go keeps reading `SERVER_PORT`. The Blueprint sets `PORT` to `29909`. `HEALTHCHECK` calls `http://127.0.0.1:${PORT}/health` with `wget`. No `$$`.
- **Rationale:** Render injects `PORT` unless the Blueprint sets it. Morph pins `9090` for the same reason. Event Logs must not treat `PORT` as its listen port in Go: a checkout `.env` sets `PORT=9090`, and `repoenv.Load` would then steal the listener from Morph's port. The entrypoint is the only place that copies `PORT` onto `SERVER_PORT`, and local `start-all.sh` does not run it. Inside the image there is no `start-all.sh`, so `repoenv.Load` is a no-op and cannot pull a dev `.env`.
- **Rejected:** Changing Go to prefer `PORT`. That collides with Morph on a normal checkout.
- **Rejected:** Leaving `PORT` unset in the Blueprint. Render would inject a different port than 29909 while the binary still used `SERVER_PORT`.

### 4. One disk, non-root server, same entrypoint shape as Morph

- **Choice:** Final stage is Alpine with `ca-certificates` and `su-exec`. User `formx` is uid/gid 65532. The Dockerfile does not set `USER`. The entrypoint creates `/data/uploads` and `/data/formsx_badger`, `chown`s `/data`, and `exec`s `su-exec formx`. Image env: `FORMSX_SQLITE_PATH=/data/formsx.sqlite`, `FORMSX_BADGER_PATH=/data/formsx_badger`, `UPLOAD_DIR=/data/uploads`, `GIN_MODE=release`. Compose uses a named volume `formx-data`. The Blueprint disk is `formx-data` at `/data`, 1 GB, `maxShutdownDelaySeconds: 120`. No `numInstances`.
- **Rationale:** A fresh volume is root-owned. SQLite's WAL sits next to the database file, so one mount covers SQLite, Badger, and uploads. A disk forces a single instance, same as Morph. Go defaults stay cwd-relative when those env vars are unset.
- **Rejected:** A second volume for uploads. Operators would back up one tree and miss the other.
- **Rejected:** `USER formx` with no entrypoint. The first disk mount is not writable.
- **ponytail:** `chown -R` on every start. Ceiling is a large `/data`. Upgrade path is to skip the walk when the mount is already uid 65532.

### 5. Secrets stay out of the build; the auth origin is a prompt

- **Choice:** No `ARG` whose name contains `JWT`, `PASSWORD`, `SECRET`, `API_KEY`, or `TOKEN`. The node stage sets `VITE_API_URL` empty and does not declare it as an `ARG`, so Render will not bake a service env into the bundle. `.env` files are ignored. `formx/deploy/.env.production.example` lists names with empty secret values. The Blueprint sets storage paths and `PORT` as plain values. `USERS_PANEL_BASE_URL`, `PUBLIC_FORM_BASE_URL`, `MORPH_AI_API_KEY`, and `SMTP_PASSWORD` are `sync: false` with no `value`. No `generateValue`.
- **Rationale:** Event Logs does not sign JWTs. It forwards login to Morph. The Morph origin is public but unknown until the dashboard assigns a host, same as MorphUtils `VITE_MORPH_API_URL`. AI and SMTP are optional; listing them as prompts avoids a committed key without pretending they are required to boot. `/health` does not call Morph, so a missing origin still passes the check. Login returns 502 until the origin is set.
- **Rejected:** Copying Morph's full provider-key list. This process does not need those integrations to match the spec, and empty prompts look required.
- **Rejected:** A guessed `onrender.com` host or `value: ""` for the auth origin. An empty value is still a value the sync would ship. `sync: false` with no `value` is the dashboard prompt this repo already uses.
- **Rejected:** Changing the Go default of `USERS_PANEL_BASE_URL`. `start-all.sh` on the host should keep `127.0.0.1:9090`. That default is wrong only inside a container, where the runbook overrides it.

### 6. Blueprint filter, contracts, and a separate image workflow

- **Choice:** Append `formx` after `morph-utils`. Its `buildFilter` lists `formx/**`, `pkg/**`, `formx/Dockerfile`, `formx/Dockerfile.dockerignore`, `formx/deploy/docker-entrypoint.sh`, and `render.yaml`. The `morph` and `morph-utils` filters stay. `formx/deploy/check-container-contract.sh` asserts this service, the Morph service's untouched pins, and the runbook strings. `morph-utils/deploy/check-container-contract.sh` accepts the third name `formx`. `.github/workflows/formx-image.yml` runs that contract and `docker build -f formx/Dockerfile .`. It does not edit `ci.yml`.
- **Rationale:** `pkg/` is an input, so a morphai change must rebuild this image. Adding `formx/**` to the Morph filter would rebuild Morph for files its Dockerfile never copies. A new job inside `ci.yml` would be a sixth required name. A separate workflow fails on a broken Dockerfile without renaming `Build image`.
- **Rejected:** No workflow, local build only. The Morph image has the same class of risk and a workflow. Event Logs should not wait for the product owner's first Render build to find a bad `COPY`.

### 7. Placeholder for #114, not a wired env

- **Choice:** Docs use `https://<event-logs public host>` as the origin to copy into MorphUtils `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`) in story #114. `PUBLIC_FORM_BASE_URL` is the same origin, for broadcast links, and is a dashboard prompt with no committed value. This change does not add either embed variable to `morph-utils` or `REACT_APP_MORPH_UTILS_URL` to `morph`.
- **Rationale:** The iframe path is already `/events-info` on whatever origin the shell is given. The host is assigned in the dashboard. The MorphUtils spec that said Event Logs is not a service is updated so the name list is `morph`, `morph-utils`, `formx`, and so the runbook does not tell the product owner to leave Event Logs out.
- **Rejected:** Setting `VITE_SHEETX_URL` now. That is #114, and the host does not exist until the product owner creates the service.

### Grill

Proposer: one combined image, Dockerfile-specific ignore, port pin via the entrypoint, disk, prompted Morph origin, placeholder URL. Reviewer challenges:

- **Assumption:** "The UI port 19909 should be the container port because that is what MorphUtils defaults to." That default is the Vite dev server. The production iframe uses a public origin with no port. The process that can serve both API and UI is the Go server on 29909. Kept.
- **Assumption:** "Prefer `PORT` in Go so Render's injected port always matches." A checkout `.env` sets `PORT=9090`. Preferring it would bind Event Logs onto Morph. The entrypoint is the adapter. Kept.
- **Assumption:** "Split services make health more honest." The story requires health that does not need MorphUtils or sibling embeds. It does not require the UI to look healthy while its API is absent. A split fails the single `VITE_SHEETX_URL` origin unless the UI proxies, which is a second combined design with more moving parts. Rejected.
- **Assumption:** "Root context plus the existing `.dockerignore` is enough." That ignore file drops `formx`, so `COPY formx` fails. The per-Dockerfile ignore file is the fix. The image workflow is the proof.
- **Assumption:** "`/health` should ping Morph auth." Then a shell-only health check fails closed when Morph is down, which the story forbids. Startup already refuses a bad disk. The handler returns JSON and touches no network.
- **Assumption:** "Morph's `NoRoute` index.html fallback is safe to copy." It also answers unknown `/api` paths with HTML. This mount must not. API misses stay API misses.
- **Assumption:** "A missing `dist` during `go test` or `start-all` should still mount the SPA." Mounting an absent tree turns `/` into a 404 file. Skip the mount when `index.html` is absent. The image always has the file.
- **Failure:** `$$PORT` in `HEALTHCHECK` requests `<pid>PORT`. Use `${PORT}`.
- **Failure:** `envsubst` on an nginx template. There is no nginx template. The entrypoint only exports `SERVER_PORT`.
- **Failure:** a second `PORT` or `dockerContext: .` trips the Morph contract. That script on this main scopes pins to the `morph` block. Do not add `numInstances` or `generateValue` anywhere in the file.
- **Failure:** the MorphUtils contract rejects a third service. Update its expected name list in the same change.
- **Failure:** `Dockerfile.dockerignore` is ignored by a non-BuildKit builder, `COPY formx` fails, and the image workflow goes red. That is the alarm. Do not "fix" it by dropping the root `formx` ignore line.
- **Failure:** a committed example origin or `SMTP_PASSWORD=secret` ships. Prompts have `sync: false` and no value. The example file's secret keys are empty.
- **Out of scope held:** no Render API, no embed env on MorphUtils, no Morph chip URL, no sibling Dockerfiles, no `ci.yml` rename.

## Risks / Trade-offs

- [Login fails until `USERS_PANEL_BASE_URL` is the public Morph origin] → `/health` still returns 200. The runbook says to set `https://<morph public host>` and restart.
- [Inside a container the Go default `127.0.0.1:9090` is this container, not Morph] → the Blueprint prompt has no default value, so production does not silently keep the loopback default unless the product owner skips the prompt. Document that skip as a login failure, not a health failure.
- [`chown -R` every start] → accepted, same ceiling as Morph.
- [Single writer] → one replica. The runbook says not to scale the disk service.
- [GitHub's five required checks do not build this image] → `formx-image.yml` does. It is not one of those five names.
- [Blank `PUBLIC_FORM_BASE_URL` puts `http://localhost:19909` in broadcast mail] → the prompt is documented as `https://<event-logs public host>`. Mail is optional; health does not send mail.
- [nginx is not in this image] → no static-site Render runtime. The Go process is what `/health` and `/events-info` share.
- [`npm run build` is `tsc -b && vite build`] → `ConfirmContext` did not narrow the confirm/alert union (`cancelLabel` / `danger` on the alert arm, unused `isAlert`), so the image build failed before the container started. A type predicate fixes the check and removes the unused local. Dialog behavior is unchanged.

## Migration Plan

There is no Event Logs service and no image in production. After merge, the product owner syncs the Blueprint in project `prj-dahc33dbedkc73a1v8n0`. That adds `formx` beside `morph` and `morph-utils`. Rollback is deleting that service in the dashboard. Local `start-all.sh` and `npm run dev` stay. No data migration.

## Open Questions

None.
