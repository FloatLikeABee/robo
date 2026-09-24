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
