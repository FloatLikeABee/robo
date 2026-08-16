# Academi Ledger

Account booking, lightweight accounting, warehouse inventory, fixed assets, and CSV/JSON/HTTP imports—**Go** API with **React** UI. Fits desktop and phone (responsive layout + mobile bottom navigation).

## Prerequisites

- [Go](https://go.dev/dl/) 1.22+
- [Node.js](https://nodejs.org/) 20+ (for the frontend)
- [MySQL](https://dev.mysql.com/doc/) 8.x
- [Redis](https://redis.io/) (optional but recommended for refresh tokens)

## Quick start

1. **Clone and configure environment**

   ```bash
   cp .env.example .env
   ```

   Edit `.env` with your MySQL credentials and a strong `JWT_SECRET`. Passwords may contain special characters (for example `@`); the server builds the DSN with `mysql.Config`, so no manual URL-encoding is required in `.env`.

2. **Create the database** (empty schema is fine—the API applies DDL on startup):

   ```sql
   CREATE DATABASE tran CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

3. **Start Redis** (if you use `/api/v1/auth/refresh`):

   ```bash
   redis-server
   ```

4. **Run the API** (from repo root, `.env` is picked up when the working directory is `backend/` or the parent folder):

   ```bash
   cd backend
   go run ./cmd/server
   ```

   Default listen address: `http://127.0.0.1:9095` (override with `APP_PORT`). Health check: `http://127.0.0.1:9095/health`

5. **Run the frontend**:

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

   Open [http://localhost:5174](http://localhost:5174) (or [http://127.0.0.1:5174](http://127.0.0.1:5174)).

   In **development**, the UI calls the API at `http://127.0.0.1:9095` by default so login works even when the page is **not** served by Vite (for example Cursor’s Simple Browser / preview), where relative `/api` would hit the wrong host and return HTML. Vite still proxies `/api` if you rely on same-origin requests. Override with `VITE_API_BASE_URL` in `frontend/.env` when needed.

6. **First use**: register at `/register` to create an organization. A default **chart of accounts** is seeded automatically. Use **Settings** to set country, currency, tax percent, fiscal year start, and accounting method (accrual vs cash).

### Development-only auth (until OAuth)

With `APP_ENV=development`, the API seeds **admin@example.com** / password **`AdminExample2026!`** (same as UsersPanel’s default bootstrap admin) and registers **`POST /api/v1/auth/dev-login`** (no body), which returns JWTs like a normal login. The Vite app in dev calls this automatically on load so you usually skip the login screen. This route is **not** registered when `APP_ENV=production`.

## Default ports

| Service    | Default |
|-----------|---------|
| Frontend (Vite) | `5174` |
| Backend API     | `9095` |

Ensure `CORS_ORIGIN` in `.env` matches the URL you use for the UI when `APP_ENV=production`. With `APP_ENV=development`, the API reflects the request `Origin` header so local tools (Cursor preview, alternate hosts) still work.

### Frontend (`frontend/.env`)

| Variable | Purpose |
|----------|---------|
| `VITE_API_BASE_URL` | API origin for production builds or custom dev URLs (no trailing slash). If unset in dev, the client uses `http://127.0.0.1:9095`. |

## Environment variables (backend / repo `.env`)

| Variable | Purpose |
|----------|---------|
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | MySQL connection |
| `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB` | Redis for refresh-token storage |
| `JWT_SECRET` | HMAC key for access tokens (set a long random value in production) |
| `JWT_ACCESS_EXPIRY_MIN` | Access token lifetime in minutes |
| `JWT_REFRESH_EXPIRY_DAYS` | Refresh token lifetime stored in Redis |
| `APP_ENV` | `development` or `production` (affects dev-only JWT fallback warning) |
| `APP_PORT` | HTTP port for the API (default in examples: `9095`; match `VITE_API_BASE_URL` / dev client default) |
| `CORS_ORIGIN` | Allowed browser origin for the SPA |

See [.env.example](.env.example) for a full template.

## AI provider configuration

Booki uses the shared **MorphAI** client ([`pkg/morphai`](../../pkg/morphai)) for the platform assistant drawer (P&L, customers, bookings, calculations). Without a key, **deterministic** assistant flows still work; the LLM enables natural-language answers.

### 1. Get a DashScope API key

Create an API key at [Alibaba Cloud DashScope](https://dashscope.aliyun.com/) (Qwen / text generation).

### 2. Configure the backend

Add to the repo **`.env`** (same file as MySQL/JWT settings — see [.env.example](.env.example)):

```bash
MORPH_AI_API_KEY=sk-your-dashscope-key
MORPH_AI_MODEL=qwen3-max
```

| Variable | Default | Purpose |
|----------|---------|---------|
| `MORPH_AI_API_KEY` | _(empty)_ | **Required** for LLM chat |
| `MORPH_AI_MODEL` | `qwen3-max` | Chat model |
| `MORPH_AI_API_URL` | DashScope text-generation URL | Optional endpoint override |

**Legacy fallbacks:** `GEMINI_API_KEY`, `GEMINI_MODEL`, `TRAN_QWEN_*`.

The API loads `.env` when run from `backend/` or the repo root.

### 3. Restart the API

```bash
cd backend && go run ./cmd/server
```

Or: `./start-all.sh restart booki-api`

### 4. Verify

1. Open `http://localhost:5174`.
2. Open the **Booki AI assistant** drawer.
3. Try `P&L this month` or a free-form accounting question.

**Assistant API:** `POST /api/v1/assistant/chat` (Bearer JWT).

See also: [`AI_ASSISTANT_MORPHAI_CONTRACT.md`](../../AI_ASSISTANT_MORPHAI_CONTRACT.md).

## Operations cheat sheet

- **Build API binary**

  ```bash
  cd backend && go build -o bin/server ./cmd/server
  ./bin/server
  ```

- **Production frontend build** (static files in `frontend/dist/`)

  ```bash
  cd frontend && npm run build
  ```

  Serve `dist/` behind any static host or reverse proxy, and point API calls to the same origin or configure your proxy so `/api` reaches the Go service.

- **Schema / migrations**  
  SQL is embedded in `backend/internal/database/schema.sql` and executed when the API connects. If a previous partial schema causes errors, fix or drop conflicting tables (or recreate the database) and restart the API.

- **Refresh tokens**  
  If Redis is unavailable, login still returns a refresh token, but `POST /api/v1/auth/refresh` responds with `501` until Redis is reachable.

- **Imports (API only)** — no Imports or Broadcast UI modules in Booki; use Data AI for reference file import, Morph Utils for broadcast. Remaining programmatic endpoints:
  - `POST /api/v1/imports/csv?entity=products|assets` — multipart field `file`  
  - `POST /api/v1/imports/json?entity=products|assets` — JSON array body  
  - `POST /api/v1/imports/http?entity=products|assets` — JSON body `{ "url", "method", "headers" }` fetching a JSON array or `{ "data": [...] }`

  Use a valid `Authorization: Bearer <access_token>` header on all authenticated routes.

## Project layout

```text
backend/
  cmd/server/          # API entrypoint
  internal/            # config, auth, handlers, database schema embed
  migrations/          # copy of schema for reference
frontend/
  src/                 # React app, routes, pages, Zustand store
design.md              # Product / UX specification (reference)
design-db-ui.md        # ERD / UI tokens (reference)
```

## Production notes

- Set `GIN_MODE=release` and a strong `JWT_SECRET`.
- Restrict `CORS_ORIGIN` to your real front-end URL.
- Run MySQL and Redis with TLS and secrets management appropriate to your environment.
- Do not commit `.env`; use `.env.example` as documentation only.
