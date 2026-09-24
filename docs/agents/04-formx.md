# 04 — Event Logs (`formx/`)

## Overview

User-facing product: **Event Logs**. Repository folder stays `formx`. MorphUtils embed id is `sheetx`. Default iframe URL is `/events-info`. Legacy `/survey-bot` redirects there.

| | |
|--|--|
| Backend | `formx/backend/` Go + Gin |
| Frontend | `formx/frontend/` React + Vite + Tailwind (dark) |
| Ports | API `29909`, UI `19909` |
| Auth | Morph SSO (`USERS_PANEL_BASE_URL`) |
| Data | SQLite `FORMSX_SQLITE_PATH` + Badger `FORMSX_BADGER_PATH` |

Inner tab: **Events & Info** only. Do not restore an Info Sheets tab. Do not label the module Survey Maker / SurveyX.

## Storage

On startup (`cmd/server/main.go`):

- `mysql.NewDB(cfg)` → **SQLite** (fatal log says `sqlite`)
- `mongo.NewStore(cfg)` → **Badger** (fatal log says `badger`)

Package directories are still named `internal/mysql` and `internal/mongo`. Do **not** install MySQL or MongoDB.

SQLite holds forms, pages, questions, rules, graph outbox. Badger holds responses, events-info, AI docs, survey-bot templates/results.

## Layout

```
formx/backend/
├── cmd/server/main.go
└── internal/
    ├── config/
    ├── handler/     forms, events-info, public, auth proxy, leftover assistant/MCP
    ├── models/
    ├── mysql/       SQLite repos
    ├── mongo/       Badger repos
    ├── mail/
    └── surveybot/
```

## MorphUtils

Embed: `VITE_SHEETX_URL` (fallback `VITE_FORMSX_URL`) + `/events-info`. Chat is Morph AI, not a satellite drawer.

## Run

```bash
./start-all.sh start formx
# UI http://localhost:19909
# Swagger http://localhost:29909/swagger/index.html
```

Env: repo-root `.env`. See [`formx/README.md`](../../formx/README.md).
