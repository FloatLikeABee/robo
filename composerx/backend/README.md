# Content Maker backend (`composerx/`)

Go + Gin API for **Content Maker**: compose and publish markdown/HTML. User-facing name is Content Maker; folder stays `composerx`.

Agent notes: [`docs/agents/05-composerx.md`](../../docs/agents/05-composerx.md).

## Prerequisites

- Go 1.21+
- Repo-root `.env` (no MySQL, MongoDB, or Redis)

## Storage

| Variable | Default |
|----------|---------|
| `COMPOSERX_SQLITE_PATH` | `./data/composerx.sqlite` |
| `COMPOSERX_BADGER_PATH` | `./data/composerx_badger` |
| `TRAN_FILE_STORAGE_PATH` | `./storage` |
| `PORT` / `COMPOSERX_PORT` | `8043` |
| `USERS_PANEL_BASE_URL` | `http://127.0.0.1:9090` (Morph auth) |

Column names such as `content_mongo_id` are leftover. `main` opens SQLite + Badger.

```bash
cp ../../.env.example ../../.env
```

## AI

Default DashScope / Qwen via [`pkg/morphai`](../../pkg/morphai). The same client can select a named provider; see that README.

```bash
MORPH_AI_API_KEY=sk-your-dashscope-key
MORPH_AI_MODEL=qwen3-max
```

Optional reference-library embeddings (separate key):

```bash
TRAN_OPENAI_API_KEY=sk-...
```

Optional JSON: `ai.config.example.json` → `ai.config.json` (`TRAN_AI_CONFIG_PATH`). `MORPH_AI_*` env wins over JSON.

Platform chat is **Morph AI**. Leftover endpoints: `POST /ai/assistant/chat`, `POST /ai/composer-chat`, `POST /ai/reference-docs/upload`.

## Run

```bash
go mod tidy
go run .
```

Or: `./start-all.sh restart composerx-api`

```bash
curl http://localhost:8043/health
```

UI: `http://localhost:8044` (`composerx/frontend`).

## Image

One container serves the API and the production UI on port 8043. Build context is the repo root. This repo does not call Render.

```bash
docker build -t composerx:local -f composerx/Dockerfile .
docker run --rm -p 127.0.0.1:8043:8043 composerx:local
```

`GET /health` returns HTTP 200 and `"status":"ok"` without Morph or MorphUtils. `GET /` is the UI. Other API routes need an `Authorization` bearer that Morph accepts at `USERS_PANEL_BASE_URL`. `X-User-Role` and `X-User-Permissions` are not a session. The image sets `VITE_API_BASE` empty so the UI calls that same origin. A checkout `npm run build` without that variable still targets `http://localhost:8043`.

`docker compose -f composerx/docker-compose.yml up --build` uses `composerx/deploy/.env.production` when that file exists (copy from `composerx/deploy/.env.production.example`). Data is the named volume at `/data`.

| Variable | Image default |
|----------|----------------|
| `PORT` / `COMPOSERX_PORT` | `8043` (`COMPOSERX_PORT` wins when both are set) |
| `COMPOSERX_SQLITE_PATH` | `/data/composerx.sqlite` |
| `COMPOSERX_BADGER_PATH` | `/data/composerx_badger` |
| `TRAN_FILE_STORAGE_PATH` | `/data/storage` |
| `USERS_PANEL_BASE_URL` | unset in the image; code default `http://127.0.0.1:9090` is only for a checkout. On Render set `https://<morph public host>`. It is not a secret. |
| `MORPH_AI_API_KEY` | empty. Chat stays limited until set. |
| `TRAN_OPENAI_API_KEY` | empty. Embeddings stay off until set. |

Do not pass keys as `docker build` arguments. `sh composerx/deploy/check-container-contract.sh` checks the image contract without a daemon.

## Render

The product owner syncs `render.yaml` in project `prj-dahc33dbedkc73a1v8n0`. The service is `composerx`. Secrets and the Morph origin are dashboard prompts. See [`deploy/README.md`](../../deploy/README.md). The placeholder for story #114 is `https://<composerx public host>`. This change does not set `VITE_COMPOSERX_URL`.
