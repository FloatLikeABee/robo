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
