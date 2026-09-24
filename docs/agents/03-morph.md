# 03 — Morph AI and MorphNotes

## Overview

**Morph** is the Go API (`:9090`) plus React CRA UI (`:3031`). It is two products in one SPA:

- **Morph AI** — system chat, Skills, optional AI tools workspace
- **MorphNotes** — Tasks, Timelines, Big notes, Generic data, Settings (Users)

Deep operator notes: [`morph/README.md`](../../morph/README.md).

| | |
|--|--|
| Backend | `morph/` Go + Gin |
| Frontend | `morph/frontend/` React 18 (CRA) |
| Ports | API `9090`, UI `3031` |
| Auth | Self-hosted JWT + bcrypt (see `01-auth-flow.md`) |
| Data | SQLite `TRAN_SQLITE_PATH` (`./data/tran.sqlite`) + Badger (`DB_PATH`, entity details) |

This is **not** a Transfinder / school / SQL Server product. MorphNotes UI lives at `/morphdata` (legacy `/skoolz` and similar paths may redirect).

## Startup (`morph/main.go`)

1. `pkg/repoenv` + `config.GetConfig()` (root `.env`), then the `MORPH_ENV` startup guard (local defaults still start; `production` refuses weak JWT/admin secrets before listen). See [`docs/security-hosting-checklist.md`](../security-hosting-checklist.md).
2. Badger + in-memory cache
3. MorphAI client from `MORPH_AI_API_KEY`
4. SQLite Tran store (`TRAN_SQLITE_PATH`) including `plat_users` (production also checks stored admin passwords)
5. Gin: CORS, `AuthzMiddleware`, Swagger, API, static SPA
6. Listen on `PORT` (default `9090`)

On macOS, `start-all.sh` **builds** the binary before run (Badger `LC_UUID`).

## MorphNotes nav

From `AppDrawer.js` (base `/morphdata`):

| Label | Route |
|-------|--------|
| Tasks | `/morphdata/case-tasks` |
| Timelines | `/morphdata/timelines` |
| Big notes | `/morphdata/big-notes` |
| Generic data | `/morphdata/generic-data` |
| Settings → Users | `/morphdata/configuration/users` |

Case tasks require `start_at` / `end_at`. Do not add Morph AI / Notes header shortcuts on the MorphNotes chrome; chat is the Morph AI SPA.

## Important API groups

| Prefix | Purpose |
|--------|---------|
| `/api/auth` | Login, me, user, permissions |
| `/api/admin` | User CRUD |
| `/api/chat` | Morph AI sessions and messages |
| `/api/tran/*` | MorphNotes entities (tasks, timelines, research, generic data, …) |
| `/api/forms/*`, `/api/knowledge/*`, `/api/graph/*` | Quick sheets, knowledge files, graph search/health |
| `/api/skills` | Skills catalog / markdown upload. Also embeds up to 8 **enabled** agent lessons for the caller |
| `/api/agent-lessons` | Caller-owned session lessons (list, enable/disable, delete) |

## Agent lessons

After a significant Morph AI session, distillation stores one lesson in SQLite `agent_lesson` (`morph/db/agent_lesson.go`, harvest in `morph/handlers/agent_lesson.go`). Lessons are **per user** (`owner_user_id`). Chat prompts (`buildAgentLessonsContext`) and the Skills catalog include only that user's **enabled** lessons. The Settings UI should call these routes with the same Morph bearer token as `/api/auth/me`. A client-supplied `X-User-ID` is not a lesson identity: the lesson routes return 401, and prompts and the Skills `lessons` array stay empty.

| Method | Path | Body | Result |
|--------|------|------|--------|
| `GET` | `/api/agent-lessons` | | `{ "lessons": [ ... ], "total": N }` — every lesson the caller owns, including disabled |
| `PATCH` | `/api/agent-lessons/:id` | `{ "enabled": true }` or `{ "enabled": false }` | the updated lesson |
| `DELETE` | `/api/agent-lessons/:id` | | `{ "ok": true }` |

A missing lesson and another user's lesson both return **404** `{ "error": "lesson not found" }` so callers cannot tell those cases apart. `enabled` is required on PATCH (400 if omitted). Any other JSON field on PATCH is 400 and the lesson is unchanged. Disabling or deleting a lesson changes the management-chat exact-query cache key, which is the bearer user plus a fingerprint of that user's enabled lessons. A client `X-User-ID` does not read that cache. `auth_user_id` is not a lesson identity: after #22 the middleware still copies an unverified `X-User-ID` into it when no bearer was checked.

Lesson object fields: `id`, `trigger`, `rule`, `source_session_id`, `created_at`, `enabled`, `owner_user_id`.

**Existing rows.** Lessons written before this ownership column were global. Migration (`migrateAgentLessonColumns` in `morph/db/sqlite_schema.go`, same idempotent `ADD COLUMN` path as other SQLite columns) adds `enabled INTEGER NOT NULL DEFAULT 1` so existing rows stay enabled, and `owner_user_id TEXT NOT NULL DEFAULT ''`. `ensureTranSQLiteSchema` creates `plat_users` before this backfill, including on a brand-new database. The claim is one-shot, recorded in `agent_lesson_owner_backfill`. While that table has no decision and there are zero accounts, rows stay unowned so a later start can still claim them (bootstrap admin is created after the first schema pass). The first start that sees exactly one account assigns the unowned rows that exist then. If that account already has a lesson for the same source session, the claim keeps that lesson, deletes the conflicting unowned row, and still starts. A claim error is logged and does not stop startup. The first start that sees more than one account records a decision and leaves them unowned forever, including if the database later drops to one account. If `plat_users` is missing (a partial SQLite file), the owner backfill is skipped, no decision is recorded, and startup continues. Unowned rows are not listed, injected, updated, or deleted. New lessons store the chatting user's bearer subject. Uniqueness is `(owner_user_id, source_session_id)`, so each user can have their own lesson for the shared `"default"` session id. A request with no bearer is 401 even when SQLite is unavailable; a bearer that cannot be checked because the store is down is 503.

`POST`, `PUT`, `PATCH`, and `DELETE` on `/api/tran/*`, `/api/forms/*`, `/api/knowledge/*`, and `/api/graph/*` require a Morph JWT. `GET` and `HEAD` on those prefixes stay available without a session, including list and detail calls the MorphNotes UI makes before login. Published HTML is a separate allowlist (GET/HEAD only):

- `GET /api/tran/public/big-notes/:slug`
- `GET /api/tran/public/timelines/:slug`
- `GET /api/tran/public/research/:slug`

See [`01-auth-flow.md`](01-auth-flow.md).

## Auth hub

Other apps point `USERS_PANEL_BASE_URL` at this API. Admin bootstrap: `EnsureBootstrapAdmin` in SQLite `plat_users`.

## Optional GraphRAG

Knowledge / Neo4j is optional. See [`docs/MORPH_GRAPH_OPS.md`](../MORPH_GRAPH_OPS.md).

## Local MCP

`morph-mcp` (`morph/cmd/morph-mcp`) is a stdio Model Context Protocol server. HTTP paths such as `/ai/mcp-tools` are Morph AI tool catalogs, not MCP. Build, env, and client config: [`14-morph-mcp.md`](14-morph-mcp.md).
