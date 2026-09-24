# MorphUtils

Integrated shell for **Event Logs**, **Content Maker**, **Data Access**, and **Project**. Module frontends stay in their own folders; MorphUtils embeds them in one dark Morph-themed workspace.

## Dev

```bash
cd morph-utils/frontend
npm install
npm run dev
```

Open http://localhost:3040

Prefer the repo launcher: `./start-all.sh start morph-utils` starts the MorphUtils shell plus Data Access UI and API (`sharpreport-ui`, `sharpreport-api`).

Event Logs, Content Maker, and Project UIs must still be running on their default ports (or set `VITE_*_URL` in the **repo-root** `.env`). If Data Access is down, MorphUtils shows a start hint instead of a refused iframe.

Sign in with Morph auth so the shared session cookie is present. Chat is **Morph AI**, not a drawer in this shell.

## Environment

Dev (`npm run dev`) reads these from the **repo-root** `.env`. Unset embed variables use the localhost defaults below. Unset `VITE_MORPH_API_URL` / `VITE_USERS_PANEL_API_URL` keeps auth on the same origin (`/api`, proxied to Morph on port 9090).

| Variable | Dev default when unset |
|----------|------------------------|
| `VITE_MORPH_API_URL` | same-origin `/api` (alias: `VITE_USERS_PANEL_API_URL`) |
| `VITE_USERS_PANEL_API_URL` | used only when `VITE_MORPH_API_URL` is unset or blank |
| `VITE_SHEETX_URL` | `http://localhost:19909` (alias: `VITE_FORMSX_URL`) |
| `VITE_FORMSX_URL` | `http://localhost:19909` (legacy alias) |
| `VITE_COMPOSERX_URL` | `http://localhost:8044` |
| `VITE_DATAX_URL` | `http://localhost:5178` |
| `VITE_PROJECTS_URL` | `http://localhost:5179` (alias: `VITE_MORPH_ENGI_URL`) |
| `VITE_MORPH_ENGI_URL` | `http://localhost:5179` (legacy alias) |
| `VITE_MORPH_AI_URL` | `http://localhost:3031` |

The first non-empty value wins: runtime primary, runtime alias, value pinned at image build, build-time alias. A blank value does not block the next one.

Embed ids stay `sheetx`, `composerx`, `datax`, `projects`. Event Logs iframe defaults to `/events-info`.

## Deploy

The image is the MorphUtils shell only. It does not include Morph, Event Logs, Content Maker, Data Access, or Project. The local commands below do not use the root `Dockerfile` or `render.yaml`.

From the repo root:

```bash
docker build -t morph-utils:local -f morph-utils/Dockerfile morph-utils
docker run --rm -p 127.0.0.1:3040:3040 \
  -e VITE_MORPH_API_URL=https://morph.example \
  morph-utils:local
```

Or from `morph-utils/`:

```bash
VITE_MORPH_API_URL=https://morph.example docker compose up --build
```

The process listens on `PORT` inside the container (default 3040) on all interfaces. Compose keeps that at 3040 and publishes host port `MORPH_UTILS_PUBLISH_PORT` (default 3040) on `127.0.0.1`. `GET /health` and `GET /ready` return 200 without the other apps. `GET /` is the shell.

Set the public URL variables when the container starts. A non-empty value is written to `/config.js` and overrides a build pin. The same names may be passed as `docker build --build-arg` when the URL is already known; they are not secrets. Do not pass JWTs, passwords, or API keys as build arguments. The image does not copy `.env` files.

With the embed variables unset, the production shell does not point those modules at localhost. Set `VITE_SHEETX_URL`, `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL`, and `VITE_MORPH_AI_URL` (or the aliases `VITE_FORMSX_URL` and `VITE_MORPH_ENGI_URL`) to the origins you want to embed. Health does not call them.

`sh morph-utils/deploy/check-container-contract.sh` checks the Dockerfile, entrypoint, and this section without a daemon.

## Hosted on Render

The product owner creates the shell. This repo does not call Render. After merge, sync `render.yaml` on `main` in Render project `prj-dahc33dbedkc73a1v8n0`. That adds the Docker web service `morph-utils`. The same steps are in `deploy/README.md`.

`PORT` is `3040`. `VITE_MORPH_API_URL` is required and is not a secret: set it to `https://<morph public host>` (no path), then restart so `/config.js` is rewritten. `VITE_USERS_PANEL_API_URL` is the optional alias, used only when the primary is unset or blank. Do not set `VITE_SHEETX_URL`, `VITE_FORMSX_URL`, or `VITE_COMPOSERX_URL` on this service. Event Logs is the `formx` service. Content Maker is the `composerx` service. Data Access is the `sharpreport` service. This change does not set `VITE_DATAX_URL`. The placeholder for story #114 is `https://<sharpreport public host>`. Project is the `morph-engi` service. Leave `VITE_PROJECTS_URL` and `VITE_MORPH_ENGI_URL` unset. The placeholder for those is `https://<morph-engi public host>`. This Blueprint does not set those variables:

| Variable | Module |
|----------|--------|
| `VITE_SHEETX_URL` | Event Logs (alias `VITE_FORMSX_URL`) |
| `VITE_FORMSX_URL` | legacy alias of `VITE_SHEETX_URL` |
| `VITE_COMPOSERX_URL` | Content Maker |
| `VITE_DATAX_URL` | Data Access |
| `VITE_PROJECTS_URL` | Project (alias `VITE_MORPH_ENGI_URL`) |
| `VITE_MORPH_ENGI_URL` | legacy alias of `VITE_PROJECTS_URL` |
| `VITE_MORPH_AI_URL` | Morph AI link |

Morph already allows cross-origin `Authorization` for non-loopback origins. This change does not edit CORS. There is no disk and no secret on this service.

When the deploy is live, copy the HTTPS URL Render shows. `GET /health` returns `ok`. The placeholder for story #106 is `https://<morph-utils public host>`. #106 sets that origin as `REACT_APP_MORPH_UTILS_URL` on Morph. This change does not set `REACT_APP_MORPH_UTILS_URL`. The placeholder for story #114 is `https://<sharpreport public host>`. This change does not set `VITE_DATAX_URL`.
