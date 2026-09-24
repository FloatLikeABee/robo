# 05 — Content Maker (`composerx/`)

## Overview

User-facing product: **Content Maker**. Folder stays `composerx`. This is **not** TranMail / MergeEmailX as a live product name.

| | |
|--|--|
| Backend | `composerx/backend/` Go + Gin |
| Frontend | `composerx/frontend/` Svelte + Vite |
| Ports | API `8043`, UI `8044` |
| Auth | Morph SSO (`USERS_PANEL_BASE_URL`). API routes need that bearer. `X-User-Role` is not a session |
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

## Image

`composerx/Dockerfile` (context: repo root) runs the API and the production UI on `8043`. `GET /health` does not call Morph. `USERS_PANEL_BASE_URL` is the Morph auth base URL (`https://<morph public host>` when hosted). It is not a secret. Optional keys: `MORPH_AI_API_KEY`, `TRAN_OPENAI_API_KEY`. Local `start-all.sh` still uses `./data`, API `8043`, and Vite `8044`.

Hosted service: [`deploy/README.md`](../../deploy/README.md) (Render project `prj-dahc33dbedkc73a1v8n0`). The public URL placeholder for MorphUtils `VITE_COMPOSERX_URL` (#114) is `https://<composerx public host>`. This repo does not set that variable and does not call Render.
