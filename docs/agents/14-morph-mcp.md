# 14 — Morph MCP (stdio)

`morph-mcp` is a local [Model Context Protocol](https://modelcontextprotocol.io/specification/2025-06-18) server. Clients such as Cursor and Claude Desktop launch it and speak JSON-RPC over stdin/stdout.

This is not the HTTP JSON tool catalogs already in the repo. Those catalogs describe Morph AI's internal tool loop. They are not MCP:

- Content Maker: `GET /ai/mcp-tools`
- Event Logs: `GET /api/v1/ai/mcp-tools`, `GET /api/v1/ai/mongodb-mcp`, `POST /api/v1/ai/mongodb-mcp/call`

`morph-mcp` speaks MCP revision **2025-06-18**. The pinned Go SDK (`github.com/modelcontextprotocol/go-sdk` v1.8.0) may negotiate a newer revision the client asks for, including 2026-07-28. The server is stdio only. It does not listen on a port. `initialize` advertises tools and resources. This build has three read-only tools: `whoami`, `list_my_tasks`, and `get_task`. `resources/list` is empty. There are no write tools, prompts, or `morph://` resources.

## Build

From the repo root, with Go 1.25:

```bash
cd morph && go build -o morph-mcp ./cmd/morph-mcp
```

Stdout is reserved for protocol messages. Logs go to stderr.

## Identity

Set `MORPH_MCP_TOKEN` to a Morph session JWT from `POST /api/auth/login` (`token` in the JSON body). The process verifies it with `morph/auth` (`auth.DecodeToken`, HS256). `JWT_SECRET` must be the same value the Morph API used to sign the token.

Missing or invalid tokens fail at startup with a message on stderr. The process exits non-zero and does not write a protocol message. Errors do not include the token. A raw user id is not accepted.

There is no separate API token yet. Do not commit the JWT. Pass it in the client config, not in git.

`JWT_SECRET` must be set to the same non-default value the API used to sign the token. An empty secret, or the built-in development default `morph-dev-jwt-secret-change-me`, refuses startup. The API substitutes that default when the variable is unset; morph-mcp does not. Startup verifies the token with `auth.DecodeToken` (signature and `exp`) before it checks `plat_users`. When `MORPH_ENV` is `production`, morph-mcp also applies the API's production JWT secret rules and the 168-hour token lifetime cap. A short local secret is refused in that mode.

On startup, after the token verifies, the process checks that the token `sub` still exists in `plat_users`. A deleted user cannot start a new process. Every tool call verifies the token again, including the production lifetime cap. An expired token is a tool error; the process stays up so the client can show that error. A user deleted after the process has started can keep calling tools until that process exits. Rotating `JWT_SECRET` only blocks new launches and new tool calls that fail verification. To drop a deleted account immediately, stop the running morph-mcp processes. There is no per-user revoke list in this build.

MCP never exposes private data to unauthenticated callers. `mcp.ExposeRecord` returns a record to a caller with no user id only when `publish.Visible` is true (the published slug is non-empty). It returns true for any verified user on any record, so it is not an owner check. `list_my_tasks` and `get_task` do not call it. They run only after the token verifies and the subject exists in `plat_users`, and `ownerClause` limits the SQL to that user's own Notes & TODOs.

On startup the binary loads the repo-root `.env` (same helper as the API) without overriding variables that are already set. A desktop client often has a clean environment, so set `JWT_SECRET` and `MORPH_MCP_TOKEN` in the MCP config.

Obtain a token (placeholders only):

```bash
curl -s -X POST http://127.0.0.1:9090/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"<username>","password":"<password>"}'
```

## Data access while morph-api is running

`morph-api` opens Badger at `DB_PATH` (default `./data/badger`) and `ENTITY_DETAILS_BADGER` (default `./data/entity_details`). Badger takes an exclusive directory lock. A second process that opens either directory fails with a directory-lock error and must not be pointed at those paths while the API is up.

`morph-mcp` does not open Badger, so it can start beside a running API and does not take those locks.

It opens `TRAN_SQLITE_PATH` (default `./data/tran.sqlite`, the same default as the API) read-only: a `file:` URI with `mode=ro` and `query_only`. Set an absolute path in the client config. The process does not call `db.NewTranSQL`, does not set the journal mode, and does not run migrations. The API already opens that file in WAL mode; the read-only connection works while the API holds it. A missing file fails startup. `db.New` and `db.NewBadgerEntityDetails` stay in the API process.

## Tools

| Tool | Purpose |
|------|---------|
| `whoami` | Verified JWT claims: id, email, username, roles. |
| `list_my_tasks` | The signed-in user's own Notes & TODOs (`user_note_todo`). Optional `type` (`all`, `note`, `todo`), `status` (`all`, `open`, `done`), and `limit` (default 50, capped at 100). |
| `get_task` | One of those rows by integer `id`. |

`list_my_tasks` and `get_task` are not the shared MorphNotes Tasks board (`CaseTask`). That board has no per-user owner. These tools scope every query by the token subject (`plat_users.id`), joined to that user's active Tran `User`, and return only that user's notes and todos. The text body is the SQLite column. Another user's id is not found. The tools do not say the row is forbidden, and they do not create a user row.

## Cursor

Project `.cursor/mcp.json`, or the user file `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "morph": {
      "command": "/absolute/path/to/morph-mcp",
      "env": {
        "MORPH_MCP_TOKEN": "<session JWT>",
        "JWT_SECRET": "<same non-default secret as the Morph API>",
        "TRAN_SQLITE_PATH": "/absolute/path/to/tran.sqlite"
      }
    }
  }
}
```

## Claude Desktop

The same `mcpServers` object goes in Claude Desktop's config file (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "morph": {
      "command": "/absolute/path/to/morph-mcp",
      "env": {
        "MORPH_MCP_TOKEN": "<session JWT>",
        "JWT_SECRET": "<same non-default secret as the Morph API>",
        "TRAN_SQLITE_PATH": "/absolute/path/to/tran.sqlite"
      }
    }
  }
}
```

## Environment

| Variable | Required | Purpose |
|----------|----------|---------|
| `MORPH_MCP_TOKEN` | yes | Morph session JWT. Never log it. |
| `JWT_SECRET` | yes | HMAC secret for that JWT. Must match morph-api. Empty and the development default are refused. |
| `TRAN_SQLITE_PATH` | yes for task tools | SQLite file the API already uses. Default `./data/tran.sqlite` if unset. Opened read-only. |

`DB_PATH` and `ENTITY_DETAILS_BADGER` are not read by this build.
