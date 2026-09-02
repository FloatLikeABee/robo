# 05 — Content Maker (`composerx/`)

## Overview

User-facing product: **Content Maker**. Folder stays `composerx`. This is **not** TranMail / MergeEmailX as a live product name.

| | |
|--|--|
| Backend | `composerx/backend/` Go + Gin |
| Frontend | `composerx/frontend/` Svelte + Vite |
| Ports | API `8043`, UI `8044` |
| Auth | Morph SSO (`USERS_PANEL_BASE_URL`) |
| Data | SQLite `COMPOSERX_SQLITE_PATH` + Badger `COMPOSERX_BADGER_PATH` |

Compose and publish markdown/HTML for outside readers. Optional email/SMTP remains in code; do not require MySQL, MongoDB, or Redis.

## Startup (`composerx/backend/main.go`)

1. `repoenv.Load()`
2. Open SQLite (`./data/composerx.sqlite`) and Badger (`./data/composerx_badger`)
3. Gin + Morph JWT middleware
4. Listen on `PORT` / `COMPOSERX_PORT` (default `8043`)

Column names such as `content_mongo_id` may remain from the old store. Local run does not need a Mongo server.

## AI

- Chat/compose: `pkg/morphai` / `MORPH_AI_API_KEY`
- Optional reference embeddings: `TRAN_OPENAI_API_KEY`

There is no MorphUtils assistant drawer. Chat is Morph AI.

## MorphUtils

Iframe: `VITE_COMPOSERX_URL` default `http://localhost:8044`.

## Run

```bash
./start-all.sh start composerx
```

See [`composerx/backend/README.md`](../../composerx/backend/README.md).
