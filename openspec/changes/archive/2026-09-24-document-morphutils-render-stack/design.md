# Design

## Context

See proposal.md for why. `render.yaml` already declares `morph`, `morph-utils`, `formx`, `composerx`, `sharpreport`, and `morph-engi`. Per-service sections in `deploy/README.md` say how to sync one service and which keys stay out of git. They do not number a stack create, and they defer `REACT_APP_MORPH_UTILS_URL` without saying when the image must be rebuilt.

The root `Dockerfile` sets `ARG REACT_APP_MORPH_UTILS_URL` and `ENV` before `npm run build`. CRA inlines that value. `morph/frontend/src/lib/headerAppLinks.js` omits the header link in production when the URL is empty or loopback. Render passes service env vars into the Docker build. A container restart does not change the built bundle.

MorphUtils embed origins are different. `morph-utils/deploy/docker-entrypoint.sh` writes non-empty `VITE_*` values into `/config.js` at process start. A restart applies them. An image rebuild is not required.

Existing specs forbid listing `REACT_APP_MORPH_UTILS_URL` as an env key in `render.yaml`, and they forbid committing embed URL values. This change does not edit `render.yaml`.

## Goals / Non-Goals

**Goals:**

- One numbered create order in `deploy/README.md` that a product owner can follow without Forge using Render.
- Record each public HTTPS origin, then rebuild Morph with `REACT_APP_MORPH_UTILS_URL`.
- An env matrix of placeholders for that URL and each embed origin story #114 applies.
- A contract check that fails if those statements disappear, and that fails if the example hosts or `key: REACT_APP_MORPH_UTILS_URL` show up in `render.yaml`.

**Non-Goals:**

- Creating or updating Render services, calling the Render API, or changing Dockerfiles.
- Setting `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL`, or `VITE_MORPH_ENGI_URL` on `morph-utils`. Story #114 does that.
- Replacing the per-service port, disk, and secret tables.

## Decisions

### One section in `deploy/README.md`

**Choice:** Add `## MorphUtils stack on Render` after the Morph "After a deploy" section and before `## MorphUtils on Render`. Morph is the prerequisite (#53); the next heading is the rest of the stack. Point to that heading from `docs/agents/12-build-deploy.md`. In `morph-utils/README.md`, add one sentence that this Blueprint section does not set `REACT_APP_MORPH_UTILS_URL` and that recording the URL and the later image rebuild are under `MorphUtils stack on Render` in `deploy/README.md`. Do not add "set the variable now" to the per-service MorphUtils section. That section's contract stays "the Blueprint does not set it."

The stack section opens with the sentence `Forge must not create services on Render.` Then it says the product owner creates the services in project `prj-dahc33dbedkc73a1v8n0` and that this repository does not call Render.

**Rejected:** A new `deploy/MORPHUTILS-STACK.md`. Agent docs and the container contracts already treat `deploy/README.md` as the hosting runbook. A second file drifts.

**Rejected:** Copying the full matrix into every app README. The per-service files already have their own env tables. One matrix is the fill-in sheet.

**Rejected:** Putting `REACT_APP_MORPH_UTILS_URL` or the example hosts into `render.yaml`. `openspec/specs/morph-utils-render/spec.md` says that key is not an env key in the file. A committed host would be a guessed value. The dashboard holds the value; git holds the placeholder `https://<morph-utils public host>`.

### Numbered order

1. Sync `render.yaml` on `main` in project `prj-dahc33dbedkc73a1v8n0`. That is how MorphUtils, Event Logs, Content Maker, Data Access, and Project are created. If a service already exists, do not create another.
2. Copy each public HTTPS origin from the dashboard after `GET /health` succeeds. Do not guess a host from the service name. `formx` is the counterexample: the live host is not `formx.onrender.com`.
3. If create dropped nested env vars, set `USERS_PANEL_BASE_URL` on `formx`, `composerx`, `sharpreport`, and `morph-engi` in the dashboard when the key is missing or blank. The value is `https://<morph public host>` with no path. Set `VITE_MORPH_API_URL` on `morph-utils` the same way and restart so `/config.js` is rewritten. A MorphUtils image rebuild is not required for that key. Health stays green without these keys.
4. Only after step 2 recorded the MorphUtils origin, set `REACT_APP_MORPH_UTILS_URL` on `morph` to `https://<morph-utils public host>` and deploy so the image rebuilds. If the header link is still missing, clear the build cache and deploy again. Do not add the key to `render.yaml`.
5. Leave the embed `VITE_*` keys unset. The matrix is the sheet story #114 fills. Those keys are container-start env, not a MorphUtils image rebuild.

**Rejected:** Tell the product owner to set the embed `VITE_*` keys in this story. Acceptance asks for placeholders the wire-up story uses. Applying them now would do #114's job and contradict the current "unset on morph-utils" contracts.

**Rejected:** A Render API script that creates the services and writes the URLs back. The team rule is Forge creates nothing on Render. This repository does not call Render.

### Examples, not required hosts

The runbook lists hosts already created by the product owner and labels each as an example, not a requirement to recreate:

| Service | Example origin |
|---------|----------------|
| morph (Morph panel, `USERS_PANEL_BASE_URL`) | `https://morph-gjmb.onrender.com` |
| morph-utils | `https://morph-utils.onrender.com` |
| formx | `https://formx-vucj.onrender.com` |
| composerx | `https://composerx.onrender.com` |
| sharpreport | `https://sharpreport.onrender.com` |
| morph-engi | `https://morph-engi.onrender.com` |

The origin to record is the one the dashboard shows.

### Contract check

`deploy/check-stack-runbook.sh` reads `deploy/README.md` and `render.yaml`. It fails unless the stack section contains the ownership sentence, the service names, the placeholders, the example hosts, the rebuild-versus-restart distinction, story `#114`, and `USERS_PANEL_BASE_URL`. It fails if `render.yaml` contains `key: REACT_APP_MORPH_UTILS_URL` or any example host. `docker-image.yml` runs it beside `deploy/check-container-contract.sh`.

The check is keyword and order based: the MorphUtils record step's line must precede the rebuild step's line. That is the ceiling (a reordered paragraph that keeps those two anchors still passes). Upgrade path: parse the numbered headings if the section grows past one screen.

## Risks / Trade-offs

- [Product owner rebuilds Morph before the URL exists] → Step 4 is forbidden until step 2 has copied the origin. Empty or loopback omits the link.
- [Product owner restarts Morph and expects the header link] → The section says the image must be rebuilt, and names the `ARG` plus the production empty/loopback behavior.
- [Docker layer cache keeps the old bundle] → Clear the build cache and deploy again.
- [Blueprint create drops `USERS_PANEL_BASE_URL`] → Dashboard set after create. Health does not prove the key is present.
- [Example host treated as the only legal name] → Each host is labeled "example, do not recreate."
- [Filled matrix committed] → The section says record origins outside git. No secret values in the doc.
- [Per-service sections still say this Blueprint does not set the header URL] → They stay true. The new section is the only place that says when the dashboard value is set. `morph-utils/README.md` points at that section so the two do not read as a contradiction.

## Migration Plan

Docs only. No service change, no data migration. Rollback is reverting the doc commit. Services already created by the product owner stay as they are.

## Open Questions

None. Story #114 remains the change that sets the embed `VITE_*` keys.
