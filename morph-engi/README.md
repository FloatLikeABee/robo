# Morph Engi — Projects

AI **project documents** from files or paste, plus a simple **files library**. Rust API + Svelte UI. Embedded as **Project** in MorphUtils.

## Quick start

```bash
# From the repo root:
cp .env.example .env

# From repo root (with morph-api running):
./start-all.sh start morph-api morph-engi-api morph-engi-ui

# Or manually:
cd backend && cargo run
cd frontend && npm install && npm run dev
```

| Service | URL |
|---------|-----|
| API | http://127.0.0.1:9096/health |
| UI | http://localhost:5179 |

## What it covers

- **Projects** — upload requirements/specs and/or paste content; AI organizes into markdown + HTML you can publish
- **Files** — retained uploads and paste-origin content (list, open, delete)

Primary navigation is Projects and Files only.

## AI assistant

`POST /api/v1/assistant/chat` — MorphAI tool loop for project documents and the files library. Set `MORPH_AI_API_KEY` in `.env`.

## AI project documents

| Endpoint | Purpose |
|----------|---------|
| `POST /api/v1/projects/generate-document` | Multipart: `files` and/or `paste`; optional `title`. Requires at least one source |
| `POST /api/v1/projects/:id/publish` | Publish HTML to a public path |
| `GET /api/v1/public/projects/:slug` | Unauthenticated published HTML |
| `DELETE /api/v1/projects/:id` | Delete a project document |
| `GET/POST /api/v1/resource-files` | Files library |
| `POST /api/v1/resource-files/upload` | Store an uploaded file |
| `DELETE /api/v1/resource-files/:id` | Remove a files-library entry |

Limits: up to **5 files** per generate, each up to 8 MB. Accepted types: **PDF, TXT, CSV, MD**. Paste-only and file-only both work; an empty request is rejected before any AI call. Paste content is also saved into Files. Without `MORPH_AI_API_KEY`, generate returns HTTP 503 with an "AI not configured" message.

## Vercel (static preview)

The repo root `vercel.json` builds the Projects UI as a static site. Data lives in **browser localStorage** (no Rust API, no Morph login, no Morph AI).

On [vercel.com/new](https://vercel.com/new), import this GitHub repo and use:

| Field | Value |
|-------|--------|
| Framework Preset | **Other** (leave as-is) |
| Root Directory | `./` |
| Build Command | **ON** → `npm run vercel-build` |
| Output Directory | **ON** → `morph-engi/frontend/dist` |
| Install Command | **OFF** (root `vercel.json` already runs `npm install`) |
| Environment Variables | none |

Deploy the branch that contains `vercel.json` (not `main` until this is merged). After import, Vercel will reuse these settings from `vercel.json` even if the dashboard toggles stay off.

```bash
# Same build Vercel runs
npm run vercel-build
```

## Environment

| Variable | Default |
|----------|---------|
| `MORPH_ENGI_DATABASE_URL` | `sqlite://morph_engi.db` |
| `MORPH_ENGI_PORT` | `9096` (falls back to `PORT`) |
| `MORPH_ENGI_CORS_ORIGIN` | `http://localhost:5179` |
| `USERS_PANEL_BASE_URL` | `http://127.0.0.1:9090` (Morph auth) |
| `MORPH_AI_API_KEY` | _(empty — deterministic help only)_ |
| `MORPH_ENGI_UPLOAD_DIR` | `uploads` (relative to the process working directory) |

`MORPH_ENGI_PORT` wins over `PORT` when both are set. The container image sets only `PORT`.

## Container

One image serves the API and the built UI. From the repo root:

```bash
DOCKER_BUILDKIT=1 docker build -f morph-engi/Dockerfile -t morph-engi:local .
```

That command does not use Render or `render.yaml`. `GET /health` is the healthcheck (`wget` on `http://127.0.0.1:${PORT}/health`). The image default `PORT` is `9096`. `sh morph-engi/deploy/check-container-contract.sh` checks the Dockerfile and Blueprint without a daemon.

Copy `morph-engi/deploy/.env.production.example` to a gitignored file and pass it with `--env-file`. Leave secrets empty in the example.

| Variable | Production value |
|----------|------------------|
| `APP_ENV` | `production` |
| `PORT` | `9096` |
| `STATIC_DIR` | `/app/frontend/dist` |
| `MORPH_ENGI_DATABASE_URL` | `sqlite:///data/morph_engi.db` |
| `MORPH_ENGI_UPLOAD_DIR` | `/data/uploads` |
| `USERS_PANEL_BASE_URL` | `https://<morph public host>` (no path). It is not a secret. |
| `JWT_SECRET` | Same value as Morph, at least 32 characters. Not committed. |
| `MORPH_AI_API_KEY` | Optional for `/health`. Not committed. |

`APP_ENV=production` refuses a development `JWT_SECRET` and a loopback Morph API base before the process listens.

## Hosted on Render

The product owner creates the service. This repo does not call Render. After merge, sync `render.yaml` on `main` in Render project `prj-dahc33dbedkc73a1v8n0`. That adds the Docker web service `morph-engi`. The same steps are in `deploy/README.md`.

Disk `morph-engi-data` is mounted at `/data` (1 GB) and is a single instance. `GET /health` must return HTTP 200.

The MorphUtils module id is `projects`. The placeholder is `https://<morph-engi public host>`. Story #114 sets that origin as `VITE_PROJECTS_URL` and `VITE_MORPH_ENGI_URL` on `morph-utils`. Those prompts have no value in git. Do not set them on `morph-engi`. `GET /health` is the health check.
