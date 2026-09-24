# 03 — Morph AI and MorphNotes

## Overview

**Morph** is the Go API (`:9090`) plus React CRA UI (`:3031`). It is two products in one SPA:

- **Morph AI** — system chat and header **Skills**. The agent workspace tabs are **Notes & TODOs** and **Context & Knowledge**. Header **AI tools** opens the AI tools drawer for `bk` (Assistants, RAG, Documents, System).
- **MorphNotes** — Tasks, Timelines, Big notes, Research, Generic data

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

Drawer order in `AppDrawer.js` (base `/morphdata`):

| Label | Route |
|-------|--------|
| Tasks | `/morphdata/case-tasks` |
| Timelines | `/morphdata/timelines` |
| Big notes | `/morphdata/big-notes` |
| Research | `/morphdata/research` |
| Generic data | `/morphdata/generic-data` |

These are the drawer items, not a claim that no other `/morphdata` route exists. `/morphdata/settings` and `/morphdata/configuration` redirect to `/morphdata/generic-data`.

Case tasks require `start_at` / `end_at`. Do not add Morph AI / Notes header shortcuts on the MorphNotes chrome; chat is the Morph AI SPA.

## Research

Research is the MorphNotes module at `/morphdata/research`. A new job runs five verified online rounds. Optional reference files are `.txt`, `.pdf`, `.csv`, and `.json`.

Publish a finished job with `POST /api/tran/research/:id/publish`. That write requires a Morph JWT. The stored public path is served as HTML by `GET /api/tran/public/research/:slug`, which is on the no-session allowlist. In dev, `morph/frontend` proxies `/api` to `:9090`, so a published link opened from the Research page on `:3031` reaches that handler. The same URL on the API origin (`http://localhost:9090`) serves the HTML directly.

## Invite Signup

New accounts are provisioned in [Invite Signup](../../invite-signup/) (`invite-signup/`), not from a MorphNotes Users page. Dev UI: `http://localhost:3051`. Launcher service: `invite-signup-ui` (UI only; no API process and no `start-all.sh` alias). Start it with `./start-all.sh start invite-signup-ui` after `morph-api` is up. The Vite app proxies `/api` to `:9090`. Without the launcher: `cd invite-signup/frontend && npm run build`.

- `/redeem` is unauthenticated. It submits `POST /api/invite/redeem` and shows the new username and password once.
- `/admin` signs in with a Morph admin account (`POST /api/auth/login`, then `GET /api/auth/me`) and creates codes with `POST /api/admin/invite-codes`.

The signed-in user changes their own username or password in MorphUtils **Your account** (`PATCH /api/auth/me`). The bootstrap admin in the root README is unchanged.

## Important API groups

| Prefix | Purpose |
|--------|---------|
| `/api/auth` | Login, me, user, permissions. `PATCH /api/auth/me` updates the caller's own username and password |
| `/api/admin/users` | Admin user CRUD API. There is no MorphNotes Users page |
| `/api/admin/invite-codes` | List and create invitation codes for Invite Signup |
| `/api/invite/redeem` | Redeem one code into a new `plat_users` row |
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

`POST`, `PUT`, `PATCH`, and `DELETE` on `/api/tran/*`, `/api/forms/*`, `/api/knowledge/*`, and `/api/graph/*` require a Morph JWT. Chat (`/api/chat`) and admin (`/api/admin`) require one too. `X-User-ID` / `X-User-Role` are not a session and are not a fallback. `GET` and `HEAD` on those prefixes stay available without a session, including list and detail calls the MorphNotes UI makes before login. Published HTML is a separate allowlist (GET/HEAD of exactly `{kind}/{slug}`):

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
