## Context

See proposal.md for why. Behavior is in `specs/morph-utils-container/spec.md`.

Verified against the tree (after fast-forward to `origin/main`):

- `morph-utils/frontend` is Vite + React. `npm run build` is `tsc -b && vite build`. `vite.config.ts` sets `envDir` to the repo root, dev and preview on port 3040, and proxies `/api` to `http://127.0.0.1:9090`. There is no test script and no Dockerfile.
- `auth.ts` resolves `VITE_MORPH_API_URL ?? VITE_USERS_PANEL_API_URL ?? ''`. A blank string is kept (`??` does not skip it). An empty base makes `morphAuthUrl` return a same-origin path (`/api/auth/user`). It never defaults to localhost.
- `config.ts` defaults embeds with `??` to `http://localhost:19909` (Event Logs, also `VITE_FORMSX_URL`), `:8044` (Content Maker), `:5178` (Data Access), `:5179` (Project, also `VITE_MORPH_ENGI_URL`), and `VITE_MORPH_AI_URL` to `:3031`. Those strings are in the production bundle whenever the variable is unset at build time.
- `App.tsx` skips an iframe when `embedUrl` is empty, and probes only Data Access and Project when a URL is present. An empty embed does not affect a static health response.
- Morph CORS (`morph/main.go`) reflects loopback origins and sends `Access-Control-Allow-Origin: *` otherwise, with `Authorization` allowed and credentials off. Auth fetches send `Authorization: Bearer` and do not set `credentials: 'include'`, so a cross-origin Morph URL works without a proxy.
- Vite inlines `import.meta.env.VITE_*` at build time. `import.meta.env.DEV` is the boolean `false` in a production build, so a `DEV ? 'http://localhost:…' : ''` ternary is removed from the bundle. A string passed as a function argument is not.
- The root `Dockerfile` is the Morph API + Morph AI image. `deploy/check-container-contract.sh` rejects `$$` in that Dockerfile because shell-form `HEALTHCHECK` runs under `/bin/sh -c`. `openspec/specs/morph-container` and `platform-ci` require the Morph image to exclude MorphUtils and require exactly five check names in `ci.yml`. `.github/workflows/docker-image.yml` builds only that Morph image and validates `render.yaml`.

## Goals / Non-Goals

**Goals:**

- One MorphUtils image, optional compose, `/health` and `/ready`, env-only public URLs, no secret layers.
- Local `npm run dev` keeps the Vite proxy and localhost embed defaults.

**Non-Goals:**

- Render service creation, `render.yaml`, or `REACT_APP_MORPH_UTILS_URL` (stories #105 and #106).
- Images for Event Logs, Content Maker, Data Access, Project, AI tools, or Invite Signup.
- Editing the root `Dockerfile`, `deploy/`, or `ci.yml`.
- A new required CI check, or a job on the Morph image workflow (a failure there would fail the Morph image build).

## Decisions

### 1. nginx serves the Vite build

- **Choice:** Multi-stage image. Node 22 builds `morph-utils/frontend`. The runtime stage is `nginx:1.27-alpine`, listening on `${PORT:-3040}` on all interfaces. `/health` and `/ready` are exact locations that return `200` and the text `ok`. Other paths use `try_files` back to `index.html`. Build context is `morph-utils/` so the other apps are not in the context.
- **Rationale:** The shell is static. Health must not depend on Morph or the embeds. nginx already matches the Morph image's "static files + HTTP health" idea without a second application server.
- **Rejected:** `vite preview` as the server. `preview.proxy` targets `127.0.0.1:9090`, which is empty in a shell-only container, and preview still serves a bundle whose `VITE_*` values were fixed at build time. A Node runtime is extra weight for static files.
- **Rejected:** Reverse-proxy `/api` to Morph. That would make an empty API base work, but embed URLs are absolute and still need configuration. Morph already allows cross-origin `Authorization`. A proxy adds an upstream that health must not call, and it fails closed when Morph is down even for the shell document.
- **Rejected:** Putting this app in the root Morph Dockerfile. That image is specified to exclude MorphUtils.

### 2. Runtime `config.js` for public URLs, build-args as a pin

- **Choice:** The image build declares `ARG`/`ENV` for the nine public `VITE_*` names, defaulting each to empty, so the `??` localhost fallbacks do not fire during `vite build`. Call sites pass localhost fallbacks only as `import.meta.env.DEV ? 'http://localhost:…' : ''`, so the production bundle drops those strings. At start, an entrypoint writes `/config.js` (`window.__MORPH_UTILS_CONFIG__`) from the container environment, omitting blank values. `index.html` loads that file with a parser-blocking classic script before the deferred module. The resolver uses the first non-empty value in this order: runtime primary, runtime alias, build-time primary, build-time alias. Dev with everything unset still uses the localhost embed defaults and a same-origin API base.
- **Rationale:** Vite cannot read container env after the build. Render (later, not in this change) injects service env at runtime and also passes declared `ARG`s at build. Runtime wins when non-empty, so `docker run -e VITE_MORPH_API_URL=https://…` works without a rebuild. A build-arg still pins the URL into the bundle when the runtime value is blank, which is what a Render build sees if the same variable is present at build time. Blank runtime values do not wipe a pinned build-arg.
- **Rejected:** Build-args only. `docker run -e` would leave the bundle on the empty or localhost value. Changing a URL would require a rebuild. That fails the "supplied when the container starts" requirement.
- **Rejected:** Runtime injection only, with no `ARG`. A platform that only forwards env into `docker build` would bake nothing and would depend on the entrypoint. Keeping both costs nine empty `ARG`s and matches both local `docker run -e` and a later Render build.
- **Rejected:** Treating an explicit empty `VITE_MORPH_API_URL` as "do not consult the alias", which is what `??` does today. The image sets the primary `ARG` to empty on purpose. If that blocked `VITE_USERS_PANEL_API_URL`, the documented alias would never work in the image. Non-empty primary still wins. Local dev with both unset is unchanged.

### 3. nginx master stays root; workers stay `nginx`

- **Choice:** The entrypoint runs as root, writes `config.js` and the listen port, then `exec`s nginx. The stock image drops workers to `nginx`. `HEALTHCHECK` uses `curl` (already installed in `nginx:1.27-alpine`) against `http://127.0.0.1:${PORT}/health`. No `$$`.
- **Rationale:** The default nginx layout writes the pid and the generated conf as root. There is no data volume to chown. Workers do not run as root. The nginx Alpine Dockerfile installs `curl`, not `wget`. The Morph image's `wget` is BusyBox on a plain Alpine stage and is the wrong client to copy here.
- **Rejected:** `nginxinc/nginx-unprivileged` with a custom pid and temp directory. Same static files, more paths that must be writable by uid 101. Not worth it for a shell with no writable data.
- **Rejected:** `USER` in the Dockerfile before the entrypoint. The process could not rewrite `conf.d` or `config.js` on start, so the port and the URLs would be frozen.

### 4. No secrets in the build

- **Choice:** `.dockerignore` excludes `.env`, `.env.*`, `node_modules`, and `dist`. The Dockerfile has no `ARG` whose name is a secret (`JWT`, `PASSWORD`, `SECRET`, `API_KEY`, `TOKEN`). The runtime stage copies the Vite `dist` and the nginx template only. Compose passes the public URL variables through and does not commit an env file. The runbook says not to put keys in build-args.
- **Rationale:** The shell has no server secret. Public origins are not secrets; they are visible in the browser anyway. A copied repo-root `.env` would still be a leak if the context ever widened, so the ignore file drops env files even though the context is `morph-utils/`.
- **Rejected:** Build-args for any token. They stay in image history.

### 5. Compose is optional and local

- **Choice:** `morph-utils/docker-compose.yml` builds this Dockerfile, sets container `PORT` to 3040, and publishes `${MORPH_UTILS_PUBLISH_PORT:-3040}` on `127.0.0.1`. It does not start Morph or the embeds.
- **Rationale:** Operators can `docker compose up --build` from `morph-utils/` or `docker build` / `docker run` without compose. The host port is not the container `PORT`, same split as the Morph runbook, so a healthcheck that reads `PORT` still hits 3040 inside the container.
- **Rejected:** A service in `deploy/docker-compose.yml`. That file is the Morph stack. `deploy/check-container-contract.sh` locks its shape.

### 6. Contract script plus one node test, no new workflow

- **Choice:** `morph-utils/deploy/check-container-contract.sh` checks the Dockerfile, ignore file, health paths, `${PORT}` (no `$$`), secret `ARG`s, and that the nine `VITE_*` names from `auth.ts` / `config.ts` appear in the Dockerfile, the entrypoint, and `morph-utils/README.md`. `publicUrl.ts` holds the resolver. `node --experimental-strip-types --test` runs it. No new dependency. No edit to `ci.yml` or `docker-image.yml`.
- **Rationale:** The resolver is the behavior that can go wrong without a container. The shell script fails before a slow image build. Node 22 already strips types. A workflow job would either join the five required names or the Morph image workflow.
- **Rejected:** Vitest. New dependency for one pure function.
- **Rejected:** A GitHub workflow in this change. Out of scope, and easy to attach to the wrong required check.

### Grill

Proposer: runtime `config.js` plus empty build-args. Reviewer challenges:

- **Assumption:** "Vite env is enough if we document `--build-arg`." False. The acceptance check supplies the Morph URL at run time. Build-args alone fail that. Kept only as a pin under the runtime value.
- **Assumption:** "`??` localhost defaults are dev-only." False for `vite build` today. Unset variables keep the localhost strings. The image must set the variables to empty and the call site must hide the literal behind `import.meta.env.DEV`.
- **Assumption:** "A proxy is required because of CORS." False for this client. Morph returns `*` for non-loopback origins and allows `Authorization`. The fetch is not credentialed.
- **Assumption:** "Empty `embedUrl` breaks the shell." The nav still renders. Iframes are skipped. Health does not read embeds. A shell-only deploy shows the chrome and a blank module pane until an embed URL is set. Accepted.
- **Assumption:** "Non-root `USER` is mandatory because the Morph image drops privileges." That drop exists to write `/data`. This image has no data. Forcing uid 101 through the stock nginx paths is extra machinery. Workers are already unprivileged.
- **Failure:** entrypoint `envsubst` without a format string would erase nginx `$uri`. The entrypoint substitutes only `${PORT}`.
- **Failure:** `$$PORT` in `HEALTHCHECK` requests `<pid>PORT`. Use `${PORT}`. `wget` is not the client in this base image; `curl` is installed by the nginx Alpine Dockerfile.
- **Failure:** a module script that races `config.js`. Classic script, no `async` or `defer`. Vite hoists the built module into `<head>` and leaves `/config.js` in the body. The classic script still runs during parse, and the module is deferred until after parse, so the config object exists first. Do not add `defer` or `type="module"` to `/config.js`.
- **Failure:** compose interpolation sets `VITE_MORPH_API_URL=` and wipes a build-arg. The resolver skips blank strings.
- **Failure:** a URL containing `"` breaks `config.js`. The entrypoint JSON-escapes values.
- **Out of scope held:** no `render.yaml`, no root Dockerfile edit, no `REACT_APP_MORPH_UTILS_URL`, no sibling Dockerfiles.

## Risks / Trade-offs

- [URL change requires a new process, not a new image] → restart the container after changing env. A build-arg pin is stale until the next image build if the runtime variable is left blank.
- [Same-origin `/api` in the image does not reach Morph] → expected when the origin is unset. The runbook says to set `VITE_MORPH_API_URL`. Health still passes.
- [Blank module pane when embeds are unset] → documented. Setting the embed variables at start fills `config.js` without a rebuild.
- [nginx master is root] → no writable app data. Workers use the image's `nginx` user. Upgrade path is the unprivileged image if a later story forbids a root master.
- [`chown` is not needed] → nothing persists. Do not add a volume.
- [Alias vs empty primary] → behavior change only when the primary is set to an empty string and the alias is non-empty. That is the image default plus an alias, which the old `??` chain would have ignored.

## Migration Plan

There is no running MorphUtils image. Operators follow `morph-utils/README.md`: build from `morph-utils/`, run with env, open `/health`. Rollback is removing the container. `start-all.sh` and `npm run dev` stay as they are. No data to migrate.

## Open Questions

None. Render wiring and the Morph header link are later stories.
