# Morph AI and MorphNotes

Go API + React SPA. **Morph AI** is the system chat (port **3031**). **MorphNotes** at `/morphdata` is Tasks, Timelines, Big notes, Research, and Generic data.

New accounts are created in Invite Signup at `http://localhost:3051` (`invite-signup-ui`). Nav, Research publish, and redeem/admin flows: [`docs/agents/03-morph.md`](../docs/agents/03-morph.md).

This is not a Transfinder, school, or SQL Server product. Agent notes: [`docs/agents/03-morph.md`](../docs/agents/03-morph.md).

## Stack

| Layer | |
|-------|--|
| Backend | Go, Gin, Badger, SQLite (`TRAN_SQLITE_PATH`) |
| Frontend | React 18 (CRA) |
| Auth | Morph JWT + bcrypt (`plat_users` in SQLite). Chat, admin, and writes on `/api/tran`, `/api/forms`, `/api/knowledge`, and `/api/graph` require that JWT. `X-User-*` headers are not a session. List reads and published HTML GETs stay reachable without it (see below). |

Ports: API **9090**, UI **3031**.

## Run

From the repo root (Morph API must be up for login):

```bash
cp .env.example .env
./start-all.sh start morph
```

Or manually: build/run `morph/` then `cd frontend && npm start`. On macOS, prefer `go build` over `go run` (Badger `LC_UUID`). `start-all.sh` already builds.

Default login: see the [root README](../README.md).

## Config

Repo-root `.env` (nested leftover `.env` is ignored):

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `9090` | API |
| `TRAN_SQLITE_PATH` | `./data/tran.sqlite` | MorphNotes + `plat_users` |
| `DB_PATH` | app Badger | Chat / forms KV |
| `MORPH_AI_API_KEY` | _(empty)_ | DashScope / Qwen |
| `MORPH_ENV` | unset (local) | `production` refuses development JWT/admin secrets. See [`docs/security-hosting-checklist.md`](../docs/security-hosting-checklist.md) |
| `JWT_SECRET` | development default in `.env.example` | Signing key. Production requires 32+ unique characters |
| `ADMIN_PASSWORD` | development default in `.env.example` | Bootstrap password. Production requires 12+ unique characters |
| `JWT_EXPIRY_HOURS` | `876000` local | Production default 24; allowed 1–168 |
| `USERS_PANEL_BASE_URL` | n/a on Morph itself | Other apps point this **at** Morph `:9090` |

Relative `./data/...` paths are cwd-relative.

## API auth

`POST`, `PUT`, `PATCH`, and `DELETE` on `/api/tran/*`, `/api/forms/*`, `/api/knowledge/*`, and `/api/graph/*` need a Morph JWT. Research create and publish are in that set. Chat and admin need a Morph JWT as well. `X-User-ID`, `X-User-Role`, and `X-User-Email` do not authenticate and do not override the token. `GET` and `HEAD` on those prefixes still work without a session (MorphNotes lists are not behind login).

Published pages are an allowlist, not the whole `/api/tran/public/` prefix. With no session, only `GET` and `HEAD` of:

- `/api/tran/public/big-notes/:slug`
- `/api/tran/public/timelines/:slug`
- `/api/tran/public/research/:slug`

The SPA sends `Authorization: Bearer` from the Morph cookie when the user is signed in. Details: [`docs/agents/01-auth-flow.md`](../docs/agents/01-auth-flow.md).

## Swagger

http://localhost:9090/swagger/index.html

## MCP

`cmd/morph-mcp` is a local stdio MCP server (`whoami`, `list_my_tasks`, `get_task`). It does not open Badger, so it can run beside this API. It reads `TRAN_SQLITE_PATH` read-only. HTTP `/ai/mcp-tools` catalogs are not MCP. See [`docs/agents/14-morph-mcp.md`](../docs/agents/14-morph-mcp.md).
