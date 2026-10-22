## Context

See proposal.md for why. Behavior is in `specs/dorph-utils-render/spec.md` and the modified `dorph-render` requirement.

Verified against this branch (DorphUtils image from #104 is already in the tree):

- `dorph-utils/Dockerfile` builds with context `dorph-utils/`. The runtime stage is nginx, `ENV PORT=3040`, and `HEALTHCHECK` calls `http://127.0.0.1:${PORT}/health`. The entrypoint writes `/config.js` from the public `VITE_*` variables and substitutes only `${PORT}` into the nginx template. `/health` returns plain `ok`.
- `render.yaml` is a flat `services` list with one Docker web service, `dorph` (Singapore, starter, `branch: main`, root Dockerfile, `dockerContext: .`, `healthCheckPath: /health`, `autoDeployTrigger: checksPass`, disk `dorph-data`, `PORT=9090`). There is no `projects` block.
- `deploy/check-container-contract.sh` asserts those Dorph strings and parses every `      - key:` env block in the file into one map. A second `PORT` overwrites `9090`.
- `.github/workflows/docker-image.yml` validates the whole `render.yaml` against `https://render.com/schema/render.yaml.json` and builds only the Dorph image. `ci.yml` has five check names. Do not add a sixth.
- Dorph CORS (`dorph/main.go`) reflects loopback origins and sends `Access-Control-Allow-Origin: *` otherwise, allows `Authorization`, and sets credentials to false. DorphUtils `auth.ts` sends `Authorization: Bearer` and does not set `credentials: 'include'`.
- `resolvePublicUrl` uses the value as an origin and strips one trailing slash. `fromService` `host` is a hostname without a scheme. A scheme-less host is not an origin.
- The shell has no server secret. Embed URLs are optional. Unset embeds leave a blank module pane. Health does not call them.
- The Render schema allows another `type: web` / `runtime: docker` entry, `dockerContext` and `dockerfilePath` relative to the repo root, `sync: false` as a dashboard prompt, and `PORT` as a string value. `runtime: static` is a different service shape and does not run this image's entrypoint.

## Goals / Non-Goals

**Goals:**

- One DorphUtils shell service in the existing Blueprint, using the #104 image paths.
- A runbook the product owner can follow in project `prj-dahc33dbedkc73a1v8n0` to create that service and copy its public HTTPS URL. Placeholder for #106: `https://<dorph-utils public host>`.
- Env documented from the image: pinned `PORT`, required Dorph origin, optional alias, optional embed URLs. No invented secrets.

**Non-Goals:**

- Calling Render, creating the service, or syncing a Blueprint from this repo.
- Setting `REACT_APP_DORPH_UTILS_URL` on Dorph (#106).
- Event Logs, Content Maker, Data Access, Project, or any other service (#109–#112).
- A full create/deploy runbook beyond this shell (#113).
- Dorph CORS changes, a new `ci.yml` check, or a DorphUtils image build inside the Dorph image workflow.

## Decisions

### 1. Second flat Docker web service

- **Choice:** Append `dorph-utils` to `services`. `type: web`, `runtime: docker`, `region: singapore`, `plan: starter`, `branch: main`, `healthCheckPath: /health`, `autoDeployTrigger: checksPass`. `dockerfilePath: ./dorph-utils/Dockerfile`. `dockerContext: ./dorph-utils`. No `projects` block. No disk. No `numInstances`. No `maxShutdownDelaySeconds` (schema default 30). No `dockerCommand`.
- **Rationale:** This matches the Dorph service style and the image that already listens on `PORT` and serves `/health`. The schema's `dockerContext` is repo-root relative. The Dockerfile comment and its `COPY` paths assume context `dorph-utils/`. A disk would force a single instance for a static shell that writes nothing. nginx does not flush a store, so Dorph's 120-second shutdown window stays on `dorph` only.
- **Rejected:** `runtime: static`. Render would publish build output and would not run the entrypoint, so runtime `VITE_*` values would never reach `/config.js`.
- **Rejected:** A `projects` entry for `prj-dahc33dbedkc73a1v8n0`. `dorph-render` forbids `projects`, and a new project block would not attach to the existing project. The project id belongs in the runbook.
- **Rejected:** Building DorphUtils inside `.github/workflows/docker-image.yml`. A utils image failure would fail the Dorph image check that gates the `dorph` service. Schema validation of `render.yaml` already runs there. No new check name in `ci.yml`.

### 2. Pin `PORT` to 3040 and prompt for the Dorph origin

- **Choice:** The service sets `PORT` to `3040`. It lists `VITE_DORPH_API_URL` with `sync: false` and no `value`. It does not list `VITE_USERS_PANEL_API_URL` or the embed variables.
- **Rationale:** Render injects its own `PORT` unless the Blueprint sets one. The image, the healthcheck, and the local runbook all use 3040. The same pin is why Dorph sets `9090`. `VITE_DORPH_API_URL` is required for auth and is a public origin, not a secret. `sync: false` is the schema's dashboard prompt, which is how this Blueprint already avoids committing unknown values. The Dockerfile already declares that name as an `ARG`; Render passes service env into matching build args, and a non-empty runtime value still wins in `/config.js`. The alias and the embed URLs are optional. Prompting for them would look like required secrets for apps this Blueprint does not deploy. Unset embeds stay blank, which the image already specifies.
- **Rejected:** Leaving `PORT` unset. The process would follow Render's injected port and health would still pass, but the documented container port would no longer be 3040. The Dorph service already refused that split.
- **Rejected:** `fromService` `host` (or `hostport`) from `dorph`. Those properties are a hostname and a host:port, not `https://…`. `auth.ts` concatenates the value with `/api/auth/…`. A scheme-less host is not a URL the browser can call on the DorphUtils origin.
- **Rejected:** Committing a guessed `onrender.com` host, an empty `value: ""`, or a sample `https://dorph.example` in the Blueprint. The real Dorph host is assigned in the dashboard. An example value would be a wrong production origin.

### 3. Build filter lists only this image's inputs

- **Choice:** On `dorph-utils` only: `dorph-utils/frontend/**`, `dorph-utils/Dockerfile`, `dorph-utils/.dockerignore`, `dorph-utils/deploy/docker-entrypoint.sh`, `dorph-utils/deploy/nginx.conf.template`, and `render.yaml`. The `dorph` filter is unchanged.
- **Rationale:** Those are the files the Dockerfile copies or the build context needs. `render.yaml` is listed for the same reason as on Dorph: env edits must roll out. Adding DorphUtils paths to the Dorph filter would rebuild Dorph for files the root image never copies (`.dockerignore` excludes `dorph-utils`).

### 4. Scope the Dorph env contract to the `dorph` service

- **Choice:** `deploy/check-container-contract.sh` checks Dorph fields inside the service whose name line is exactly `name: dorph` at service indent: env keys, `dockerContext: .`, `dockerfilePath: ./Dockerfile`, and that service's `buildFilter` paths. `dorph-utils/deploy/check-container-contract.sh` asserts the new service, `PORT` `3040`, the prompted Dorph origin, the absence of a disk and of any other service name, and the runbook strings. The Dorph image workflow keeps the schema check.
- **Rationale:** The current parser is one map for the whole file, and several Dorph needles are substrings of the DorphUtils lines. `PORT` on DorphUtils would replace `9090`. `dockerContext: .` is a prefix of `dockerContext: ./dorph-utils`. `Dockerfile` occurs inside `dorph-utils/Dockerfile`. `name: dorph` is a prefix of `name: dorph-utils` and `name: dorph-data`. The Dorph check must use the service block, not a file-wide substring.

### 5. Runbook in both READMEs, no URL wired on Dorph

- **Choice:** `deploy/README.md` gains a DorphUtils section: sync the Blueprint in project `prj-dahc33dbedkc73a1v8n0`, do not call Render from this repo, set `VITE_DORPH_API_URL` to `https://<dorph public host>` (no path), leave the alias and embeds unset until those origins exist, keep `PORT` at 3040, expect `GET /health` to return `ok`, and copy the dashboard URL. The #106 placeholder is `https://<dorph-utils public host>`. This change does not set `REACT_APP_DORPH_UTILS_URL`. `dorph-utils/README.md` repeats that hosted section and keeps the local `docker build` path independent of `render.yaml`. Both files state the CORS fact from `dorph/main.go`: non-loopback origins get `*`, `Authorization` is allowed, credentials stay false, and this change does not edit CORS.
- **Rationale:** The product owner already uses `deploy/README.md` for Render. The image README is the other file the story names, and it currently says the image does not use `render.yaml`, which remains true for the local build and false if that sentence is left as the only deploy guidance.
- **Rejected:** Putting the public URL into the `dorph` service env. That is #106. The existing comment that names `REACT_APP_DORPH_UTILS_URL` as an optional build arg stays a comment, not an env entry.

### Grill

Proposer: second Docker service, pinned port, prompted Dorph origin, docs for the public URL. Reviewer challenges:

- **Assumption:** "The image assumes port 3040, so an unpinned Render `PORT` breaks health." The entrypoint and `HEALTHCHECK` both read `PORT`. An injected port would still match. The pin is so the Blueprint, the image default, and the runbook stay on 3040, which is the Dorph `PORT=9090` lesson. Kept.
- **Assumption:** "`fromService.host` is the public URL." The schema property is a hostname. The client needs an `https` origin. Rejected.
- **Assumption:** "A static site is the right Render type because the shell is static files." The static runtime does not run `docker-entrypoint.sh`. Runtime config would be stuck at build time. Rejected.
- **Assumption:** "CORS must be opened for the new origin." Non-loopback origins already receive `*`, and the auth fetch is not credentialed. Editing CORS is out of scope and unnecessary. Document only.
- **Assumption:** "Optional embed keys belong in the Blueprint so the PO cannot forget them." Empty prompts look required, and those services are #109–#112. The image already boots with them unset. Document only.
- **Assumption:** "`checksPass` compiles this image." It waits on the existing GitHub checks. Those checks do not build DorphUtils. The first Render build is the image build. Adding a workflow job would couple it to Dorph deploys. Accepted.
- **Failure:** a second `PORT` key makes the Dorph contract assert `3040` or skip the Dorph pin. `dockerContext: .` also matches `dockerContext: ./dorph-utils`, and `Dockerfile` matches `dorph-utils/Dockerfile`. Scope those checks to the `dorph` service block.
- **Failure:** a committed example origin ships to production. `sync: false` and no `value`.
- **Failure:** the runbook tells the PO to guess `https://dorph-utils.onrender.com`. The host is whatever the dashboard shows. The placeholder is `https://<dorph-utils public host>`.
- **Failure:** a disk, a `projects` block, or another service name sneaks in. The DorphUtils contract rejects them.
- **Out of scope held:** no Render API, no `REACT_APP_DORPH_UTILS_URL` value, no embed services, no CORS patch, no new `ci.yml` name.

## Risks / Trade-offs

- [Login fails until `VITE_DORPH_API_URL` is set] → `/health` still returns 200. The runbook says to set the Dorph origin and restart so the entrypoint rewrites `/config.js`. A rebuild is not required for a runtime value.
- [GitHub checks do not build this image] → schema and the contract script catch the Blueprint. A broken Dockerfile fails on the product owner's first Render build, not in `ci.yml`.
- [`checksPass` waits on checks that do not include this image] → same trigger as Dorph, as requested. It does not make a red DorphUtils Dockerfile block `dorph`.
- [Prompted `VITE_DORPH_API_URL` uses the secret-shaped `sync: false`] → the runbook says it is the public Dorph origin. The contract forbids a `value` on that key.
- [Alias left unset] → the image uses it only when the primary is blank. The prompted primary is the one the PO fills.

## Migration Plan

There is no DorphUtils service yet. After merge to `main`, the product owner syncs the Blueprint in project `prj-dahc33dbedkc73a1v8n0`. That adds `dorph-utils` beside the existing `dorph` service. Rollback is deleting that service in the dashboard. This change does not move data. Local `docker build` and `npm run dev` stay as they are.

## Open Questions

None. The Dorph header link, embed services, and the rest of the product-owner runbook are later stories.
