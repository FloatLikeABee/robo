# Morph AI and MorphNotes

Go API + React SPA. **Morph AI** is the system chat (port **3031**). **MorphNotes** is Tasks, Timelines, Big notes, Generic data, and Settings at `/morphdata`.

This is not a Transfinder, school, or SQL Server product. Agent notes: [`docs/agents/03-morph.md`](../docs/agents/03-morph.md).

## Stack

| Layer | |
|-------|--|
| Backend | Go, Gin, Badger, SQLite (`TRAN_SQLITE_PATH`) |
| Frontend | React 18 (CRA) |
| Auth | Morph JWT + bcrypt (`plat_users` in SQLite) |

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

## Swagger

http://localhost:9090/swagger/index.html

## MCP

`cmd/morph-mcp` is a local stdio MCP server (`whoami` only in this build). It does not open Badger, so it can run beside this API. HTTP `/ai/mcp-tools` catalogs are not MCP. See [`docs/agents/14-morph-mcp.md`](../docs/agents/14-morph-mcp.md).
