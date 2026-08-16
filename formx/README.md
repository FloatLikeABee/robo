# SheetX (FormsX)

Sheet builder microservice: create sheets/forms with multiple question types, publish by URL. Forms/questions live in embedded SQLite; responses and document collections live in Badger. Product brand is **SheetX**; repository folder remains `formx`.

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

No MySQL, MongoDB, or Redis is required.

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

Copy and adjust:

```bash
cp backend/.env.example backend/.env
```

Frontend (optional):

- `VITE_API_URL` — leave unset in dev (Vite proxies `/api` to backend); set for production (e.g. `https://api.example.com`).
- `VITE_USER_EMAIL` — shown in header (e.g. `ge.gao.0039@gmail.com`).

## AI provider configuration

FormsX uses the shared **MorphAI** stack ([`pkg/morphai`](../../pkg/morphai)) for the in-app assistant drawer and optional LLM-backed replies. Without a key, the assistant still runs **rule-based** create/list flows; the full LLM is enabled when an API key is set.

### 1. Get a DashScope API key

Sign up at [Alibaba Cloud DashScope](https://dashscope.aliyun.com/) and create an API key for text generation (Qwen models).

### 2. Configure the backend

Add to `backend/.env` (copy from `backend/.env.example`):

```bash
MORPH_AI_API_KEY=sk-your-dashscope-key
MORPH_AI_MODEL=qwen3-max
```

| Variable | Default | Purpose |
|----------|---------|---------|
| `MORPH_AI_API_KEY` | _(empty)_ | **Required** for LLM chat. Primary config key. |
| `MORPH_AI_MODEL` | `qwen3-max` | Chat/completion model name |
| `MORPH_AI_API_URL` | DashScope text-generation URL | Optional native endpoint override |

**Legacy fallbacks** (prefer `MORPH_AI_*`): `GEMINI_API_KEY`, `GEMINI_MODEL`, `TRAN_QWEN_API_KEY`, `TRAN_QWEN_MODEL`.

### 3. Restart the API

```bash
cd backend
go run ./cmd/server
```

Or from the repo root: `./start-all.sh restart formx-api`

### 4. Verify

1. Open the UI at `http://localhost:19909` and sign in.
2. Open the **AI assistant** drawer (header).
3. Ask a general question — you should get an LLM reply instead of the “Set `MORPH_AI_API_KEY`…” placeholder.

**Assistant API:** `POST /api/v1/assistant/chat` (JWT required).  
**Morph integration:** `GET /api/v1/ai/mcp-tools`, `GET|POST /api/v1/ai/mongodb-mcp*` for document tools.

See also: [`AI_ASSISTANT_MORPHAI_CONTRACT.md`](../../AI_ASSISTANT_MORPHAI_CONTRACT.md) in the monorepo root.

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

- **Forms** (sidebar) — searchable list with scrollable rows; create, edit, delete; Results and copy URL.
- **Events & Info** — MongoDB-only log (title, detail, reporter, time); grid + detail drawer; delete only (no edit).
- **Morph MongoDB MCP** — `GET /api/v1/ai/mongodb-mcp` + `POST /api/v1/ai/mongodb-mcp/call` for document-based AI access to synced system documents (`ai_system_documents`), `workspace_events`, and `form_responses`.
- **Broadcasts** — compose email (editable body), pick forms, recipients; history with CRUD; **Send** requires SMTP env vars.
- **Edit Form** — name, slug, questions, visibility rules.
- **Form Results** — responses table, export CSV.
- **Public form** — `http://localhost:19909/f/<slug>` to fill and submit.

## Regenerate Swagger docs

From repo root:

```bash
cd backend
go install github.com/swaggo/swag/cmd/swag@latest
```

`go install` puts `swag` in **`$(go env GOPATH)/bin`** (often `~/go/bin`). If `swag: command not found`, either:

**Option A — run without adding PATH**

```bash
"$(go env GOPATH)/bin/swag" init -g cmd/server/main.go -o docs
```

**Option B — add Go’s bin to your shell (e.g. in `~/.zshrc`)**

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
swag init -g cmd/server/main.go -o docs
```

If `go install` fails with a network error to `proxy.golang.org`, try:

```bash
export GOPROXY=https://goproxy.cn,direct   # or: GOPROXY=direct
go install github.com/swaggo/swag/cmd/swag@latest
```

Then rebuild the backend.

## Project layout

```
backend/
  cmd/server/       main entry, Swagger mount
  internal/
    config/         env config
    handler/        HTTP handlers (forms, questions, events-info, broadcasts, public)
    mail/           SMTP helper for broadcasts
    models/         Form, Question, events, broadcasts, FormResponse, DTOs
    mysql/          repos (GORM)
    mongo/          form responses + workspace_events repos
  pkg/validator/     submission validation
  docs/             generated Swagger spec
frontend/
  src/
    components/     Layout (header)
    lib/api.ts      API client
    pages/          MyForms, EditForm, FormResults, PublicForm
scripts/
  init_db.sql       MySQL schema (optional)
FORMSX_SPEC.md      Full specification
```
