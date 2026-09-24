# Event Logs (`formx/`)

User-facing product: **Event Logs**. Repository folder remains `formx`; MorphUtils embed id remains `sheetx`. Events & Info only. Default embed route: `/events-info`. `/survey-bot` redirects there.

Forms/questions live in embedded SQLite; responses and document collections live in Badger. Go packages may still be named `mysql` / `mongo` — they open SQLite and Badger. No MySQL or MongoDB server is required.

Agent notes: [`docs/agents/04-formx.md`](../docs/agents/04-formx.md).

## Stack

- **Backend:** Go (Gin), Swagger, SQLite (GORM + modernc), BadgerDB
- **Frontend:** React (Vite), TypeScript, Tailwind CSS, dark theme

## Prerequisites

- Go 1.21+
- Node 18+

## Storage

On startup the backend creates (if missing):

- SQLite at `FORMSX_SQLITE_PATH` (default `./data/formsx.sqlite`) — forms, pages, questions, rules, graph outbox
- Badger at `FORMSX_BADGER_PATH` (default `./data/formsx_badger`) — responses, events-info, AI docs, survey-bot templates/results

## Environment

Backend reads:

| Variable | Default |
|----------|---------|
| `FORMSX_SQLITE_PATH` | `./data/formsx.sqlite` |
| `FORMSX_BADGER_PATH` | `./data/formsx_badger` |
| `SERVER_PORT` | `29909` |
| `SMTP_HOST` | _(empty)_ — set to send broadcast emails |
| `SMTP_PORT` | `587` |
| `SMTP_USER` | _(empty)_ |
| `SMTP_PASSWORD` | _(empty)_ |
| `SMTP_FROM` | _(empty)_ — From address for broadcasts |
| `PUBLIC_FORM_BASE_URL` | `http://localhost:19909` — base URL appended to `/f/{slug}` in broadcast emails |
| `USERS_PANEL_BASE_URL` | `http://127.0.0.1:9090` — Morph auth |

Copy and adjust at the **repo root** (nested `backend/.env` is ignored):

```bash
cp .env.example .env
```

Frontend (optional):

- `VITE_API_URL` — leave unset in dev (Vite proxies `/api` to backend); set for production (e.g. `https://api.example.com`).

## AI

Event Logs uses [`pkg/morphai`](../pkg/morphai) for optional LLM-backed APIs. The default is DashScope via `MORPH_AI_*`. Named providers are available on that same client. Platform chat is **Morph AI** (`:3031`), not a satellite drawer.

```bash
MORPH_AI_API_KEY=sk-your-dashscope-key
MORPH_AI_MODEL=qwen3-max
```

Restart: `./start-all.sh restart formx-api`

Leftover routes such as `GET|POST /api/v1/ai/mongodb-mcp*` still talk to the **Badger** document store (name is historical).

## Run

**Backend**

```bash
cd backend
go build -o bin/server ./cmd/server
./bin/server
```

API: `http://localhost:29909`  
Swagger UI: `http://localhost:29909/swagger/index.html`

**Frontend**

```bash
cd frontend
npm install
npm run dev
```

App: `http://localhost:19909`

Or from repo root: `./start-all.sh start formx`

## Regenerate Swagger docs

From `formx/backend`:

```bash
"$(go env GOPATH)/bin/swag" init -g cmd/server/main.go -o docs
```

If `go install github.com/swaggo/swag/cmd/swag@latest` cannot reach `proxy.golang.org`, try `GOPROXY=https://goproxy.cn,direct`.

## Container

One image serves the API and the UI. Build context is the repo root:

```bash
docker build -f formx/Dockerfile -t formx:local .
docker run --rm -p 127.0.0.1:29909:29909 \
  -e USERS_PANEL_BASE_URL=http://host.docker.internal:9090 \
  -v formx-data:/data \
  formx:local
```

`PORT` inside the container is `29909`. The entrypoint copies it to `SERVER_PORT`. A checkout `.env` with `PORT=9090` does not apply inside the image. `GET /health` returns `{"status":"healthy"}` and does not call Morph or MorphUtils. `GET /events-info` is the UI. `GET /api/` stays the API.

Optional compose, from `formx/`:

```bash
cp deploy/.env.production.example deploy/.env.production
docker compose up --build
```

`deploy/.env.production` is gitignored. Secret keys in the example are empty.

| Variable | Image default |
|----------|----------------|
| `PORT` / `SERVER_PORT` | `29909` |
| `FORMSX_SQLITE_PATH` | `/data/formsx.sqlite` |
| `FORMSX_BADGER_PATH` | `/data/formsx_badger` |
| `UPLOAD_DIR` | `/data/uploads` |

Named volume `formx-data` is mounted at `/data`. Run one replica. SQLite and Badger are single-writer. `start-all.sh` still uses `./data` and `./uploads` under the working directory.

`sh formx/deploy/check-container-contract.sh` checks the Dockerfile, Blueprint, and this section without a daemon.

## Deploy on Render

The product owner creates the service. This repo does not call Render. After merge, in Render project `prj-dahc33dbedkc73a1v8n0`, sync `render.yaml` on `main`. That adds web service `formx` (Singapore, starter) beside `morph` and MorphUtils. `PORT` is `29909`. `FORMSX_SQLITE_PATH`, `FORMSX_BADGER_PATH`, and `UPLOAD_DIR` are the `/data` paths above.

`USERS_PANEL_BASE_URL` is required and is not a secret: set it to `https://<morph public host>` (no path). `MORPH_AI_API_KEY` and `SMTP_PASSWORD` are optional prompts with no value in git. A disk (`formx-data` at `/data`, 1 GB) is a single instance. Do not scale it out.

`GET /health` does not call MorphUtils. After the deploy is live, copy the HTTPS origin Render shows. The placeholder is `https://<event-logs public host>`. Story #114 sets that origin as `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`) on MorphUtils. Those prompts are on `morph-utils` with no value in git. Do not set `VITE_SHEETX_URL` or `VITE_FORMSX_URL` on Morph or on the `formx` service. Event Logs does not set `REACT_APP_MORPH_UTILS_URL`. The `morph` service lists that key as a dashboard prompt with no value.

## Project layout

```
backend/
  cmd/server/       main entry, Swagger mount
  internal/
    config/         env config
    handler/        HTTP handlers (forms, questions, events-info, public)
    mail/           SMTP helper for broadcasts
    models/
    mysql/          SQLite repos (package name leftover)
    mongo/          Badger repos (package name leftover)
frontend/
```
