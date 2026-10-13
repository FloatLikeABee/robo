# 14 — Morph MCP (stdio)

`morph-mcp` is a local [Model Context Protocol](https://modelcontextprotocol.io/specification/2025-06-18) server. Clients such as Cursor and Claude Desktop launch it and speak JSON-RPC over stdin/stdout.

This is not the HTTP JSON tool catalogs already in the repo. Those catalogs describe Morph AI's internal tool loop. They are not MCP:

- Content Maker: `GET /ai/mcp-tools`
- Event Logs: `GET /api/v1/ai/mcp-tools`, `GET /api/v1/ai/mongodb-mcp`, `POST /api/v1/ai/mongodb-mcp/call`

`morph-mcp` speaks MCP revision **2025-06-18**. The pinned Go SDK (`github.com/modelcontextprotocol/go-sdk` v1.8.0) may negotiate a newer revision the client asks for, including 2026-07-28. The server is stdio only. It does not listen on a port. `initialize` advertises tools, resources, and the MCP Apps extension `io.modelcontextprotocol/ui` (MIME type `text/html;profile=mcp-app`). The tool list includes `whoami`, unified own-notes tools (`create_note`, `get_note`, `list_notes`, `update_note`, `append_to_note`, `archive_note`, `search_my_notes`), `note_grill`, the previous module names as aliases, and event-log tools. `resources/list` returns `ui://morph/notes` and `ui://morph/grill`. There are no prompts or `morph://` resources. `resources.listChanged` is not advertised.

## Build

From the repo root, with Go 1.25:

```bash
cd morph && go build -o morph-mcp ./cmd/morph-mcp
```

Stdout is reserved for protocol messages. Logs go to stderr.

## Identity

Set `MORPH_MCP_TOKEN` to either a Morph session JWT from `POST /api/auth/login` or a personal access token from `POST /api/agent-tokens`. A PAT starts with `morph_pat_`. Prefer a PAT in client configs so the machine does not need `JWT_SECRET`.

Missing or invalid tokens fail at startup with a message on stderr. The process exits non-zero and does not write a protocol message. Errors do not include the token. A raw user id is not accepted. Do not commit the token.

### Personal access tokens

Mint, list, and revoke tokens with a session JWT. There is no Settings page.

```bash
curl -s -X POST http://127.0.0.1:9090/api/agent-tokens \
  -H "Authorization: Bearer <session JWT>" \
  -H 'Content-Type: application/json' \
  -d '{"name":"cursor"}'
```

```bash
curl -s http://127.0.0.1:9090/api/agent-tokens \
  -H "Authorization: Bearer <session JWT>"
```

```bash
curl -s -X DELETE http://127.0.0.1:9090/api/agent-tokens/<id> \
  -H "Authorization: Bearer <session JWT>"
```

The create response includes the plaintext token once. The store keeps only a SHA-256 hash. Default scopes are `notes:read`, `notes:write`, and `pool:read`. Default expiry is 90 days. Ask for `pool:publish` or `admin:pool` only when needed. List is owner-scoped. Revoking another user's id is not found.

`morph-mcp` accepts `MORPH_MCP_TOKEN=morph_pat_…` without `JWT_SECRET`. It resolves the subject from the hashed row and confirms that user still exists in `plat_users`. A revoked, expired, or deleted-user token fails the next tool call. A PAT without `notes:read` cannot run a read notes tool. A PAT without `notes:write` cannot run a write notes tool (`forbidden_scope`).

### Session JWT

`JWT_SECRET` must be set to the same non-default value the API used to sign the token. An empty secret, or the built-in development default `morph-dev-jwt-secret-change-me`, refuses startup. The API substitutes that default when the variable is unset; morph-mcp does not. Startup verifies the token with `auth.DecodeToken` (signature and `exp`) before it checks `plat_users`. When `MORPH_ENV` is `production`, morph-mcp also applies the API's production JWT secret rules and the 168-hour token lifetime cap. A short local secret is refused in that mode. A session JWT is treated as `notes:read` and `notes:write`.

On startup, after the token verifies, the process checks that the token subject still exists in `plat_users`. A deleted user cannot start a new process. Every tool call verifies the token again (JWT expiry/signature, or PAT hash/revoke/expiry, plus `plat_users`). An expired or revoked token is a tool error; the process stays up so the client can show that error. A user deleted after launch fails the next tool call.

MCP never exposes private data to unauthenticated callers. `mcp.ExposeRecord` returns a record to a caller with no user id only when `publish.Visible` is true (the published slug is non-empty). It returns true for any verified user on any record, so it is not an owner check. `list_my_tasks`, `get_task`, `create_note`, and `note_grill` do not call it. The notes tools run only after the token verifies and the subject exists in `plat_users`, and `ownerClause` limits every read and the insert to that user's own Notes & TODOs. A missing or invalid token fails `create_note` before any insert. An expired token is a tool error and writes nothing. `note_grill` keeps its interview in memory for that token subject and does not insert a row.

On startup the binary loads the repo-root `.env` (same helper as the API) without overriding variables that are already set. A desktop client often has a clean environment, so set `MORPH_MCP_TOKEN` (and `JWT_SECRET` only for a session JWT) in the MCP config. How to obtain a JWT for local Morph at `http://localhost:3031/` is in [Local Morph](#local-morph).

## Data access while morph-api is running

`morph-api` opens Badger at `DB_PATH` (default `./data/badger`) and `ENTITY_DETAILS_BADGER` (default `./data/entity_details`). Badger takes an exclusive directory lock. A second process that opens either directory fails with a directory-lock error and must not be pointed at those paths while the API is up.

`morph-mcp` does not open Badger, so it can start beside a running API and does not take those locks.

It opens `TRAN_SQLITE_PATH` (default `./data/tran.sqlite`, the same default as the API) read-only for list and get: a `file:` URI with `mode=ro` and `query_only`. Set an absolute path in the client config. `create_note` opens a second connection on that same existing file with `mode=rw` and a busy timeout. `mode=rw` does not create a missing file. The process does not call `db.NewTranSQL`, does not set the journal mode, and does not run migrations. The API already opens that file in WAL mode; the read-only connection and the short insert both work while the API holds it. A missing file fails startup. `db.New` and `db.NewBadgerEntityDetails` stay in the API process.

`create_note` does not POST to the HTTP API. `POST /api/tran/notes-todos` resolves the Tran user through `tranUserIDFromContext`, which falls back to user id 1 when the session has no active Tran user. The MCP insert uses the same owner subquery as list and get, and it writes nothing in that case.

## Tools

| Tool | Purpose |
|------|---------|
| `whoami` | Verified identity: id, email, username, roles, scopes, and PAT expiry when present. Read-only. |
| `create_note` | Create one record. `kind` is `note`, `todo`, `big_note`, `timeline`, `research`, `stick`, or `task`. Omitted kind stays a Notes & TODOs `note` and returns an integer `id`. Optional `idempotency_key` and `tags`. |
| `get_note` | One of the caller's records by `kind:rowid`. Read-only. |
| `list_notes` | The caller's records. Optional `kind` and `limit`. Read-only. |
| `update_note` | Patch title, body, or tags. `expected_version` mismatch is `conflict`. |
| `append_to_note` | Append text to the body. |
| `archive_note` | Soft-delete. Later get, list, and search treat the id as not found. |
| `search_my_notes` | Owner-only FTS5 search. A junk query returns an empty page. Read-only. |
| `list_my_tasks` / `get_task` | Alias for Notes & TODOs list/get by integer id. Read-only. |
| `create_stick_note` / `list_stick_notes` / `get_stick_note` | Aliases for kind `stick`. Owner-scoped. |
| `create_timeline` / `list_timelines` / `get_timeline` | Aliases for kind `timeline`. |
| `create_story` / `list_stories` / `get_story` | Aliases for kind `big_note`. |
| `create_research` / `list_research` / `get_research` | Aliases for kind `research`. |
| `note_grill` | Interview that user, then return a draft body. `start` accepts every kind. Does not insert a row. |
| `list_event_logs` / `get_event_log` / `create_event_log` | Event Logs HTTP API. Does not open Badger. |

Every note kind is scoped to the token subject. Another user's id is not found and does not say forbidden. Stick and task rows live on `CaseTask` with `owner_user_id` and `board_kind`. The admin CaseTask HTTP list is still unfiltered.

`create_note` with omitted kind still takes `title` and `body`. At least one is required. The title limit is 200 characters and the body limit is 32000 characters. The stored title starts with `[morph-mcp]` and the stored body starts with `source: morph-mcp`. A second call with the same title and body, or the same `idempotency_key`, returns the existing id. Empty input, an over-long field, a missing Tran user, or a bad token inserts nothing.

`note_grill` takes `action` (`start`, `answer`, `skip`, or `finish`). `start` needs `kind` and `title`. Later steps need the `id` from `start`. The question count depends on the title, with a cap of four. Skip moves on. Finish, or an answer that is only a stop phrase such as "enough" or "write it now", drafts the body. The draft is the tool result. Another user's id is not found, and the error does not include the draft. Saving still requires `create_note` after the draft is shown.

An agent that should create, list, or get those notes follows [15 — Morph notes, self-use](15-morph-notes-self-use.md). The human still reviews them at `http://localhost:3031/`. The agent calls the tools in this chapter.

## Local Morph

The human UI is [http://localhost:3031/](http://localhost:3031/). Agent notes in that chat is the list a person reviews for agent-authored notes. [http://localhost:3031/agent-notes](http://localhost:3031/agent-notes) opens the same list over the chat. The API listens on `http://127.0.0.1:9090`. The UI dev server proxies `/api` to that port. `morph-mcp` is not an HTTP server on either port. The stdio process below talks to the same SQLite file the API on `:9090` uses, and the note shows up at `:3031` after Agent notes reloads.

Obtain a token (placeholders only). Either URL works while the UI dev server is up; the API port does not need the UI:

```bash
curl -s -X POST http://127.0.0.1:9090/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"<username>","password":"<password>"}'
```

```bash
curl -s -X POST http://localhost:3031/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"<username>","password":"<password>"}'
```

Use the JSON `token` value as `MORPH_MCP_TOKEN`. Do not commit it.

## MCP Apps

`list_my_tasks`, `get_task`, and `create_note` set `_meta.ui.resourceUri` to `ui://morph/notes`. `note_grill` sets the same keys to `ui://morph/grill`. The URI is also on the older flat key `_meta["ui/resourceUri"]`. `whoami` does not declare a UI. A host that supports [MCP Apps](https://modelcontextprotocol.org/extensions/apps/overview) (SEP-1865) reads that resource and renders the HTML in a sandboxed iframe. The notes page lists notes (`list_my_tasks` with type `note`), opens one (`get_task`), and creates one (`create_note`). The grill page asks the questions, shows the draft, and calls `create_note` only from its save button after that draft. Those calls go through the host to this process, so the page never sees `MORPH_MCP_TOKEN`. Owner checks stay on the server. Both pages are dark, use Arial, and have no external URL. Notes & TODOs at `http://localhost:3031/` embeds the same grill HTML.

The MCP Apps overview names these renderers: Claude, Claude Desktop, VS Code GitHub Copilot, Microsoft 365 Copilot, Goose, Postman, MCPJam, and Archestra.AI. They launch this same stdio binary. The steps below are for Claude Desktop against local Morph at `http://localhost:3031/`.

### Claude Desktop

Obtain the token with the login request in [Local Morph](#local-morph). Put it in `claude_desktop_config.json` with the other two variables. Placeholders only. Do not commit the token.

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

Restart Claude Desktop after saving the file. Ask it to open Morph notes. The panel lists that user's notes, opens one, and creates one. The stored title still starts with `[morph-mcp]`, and the human reads it in Notes & TODOs at `http://localhost:3031/`.

### Plain tools

A client that does not render MCP Apps, including Cursor, uses the same binary and the same three environment variables. It ignores the UI metadata and receives the tool results for `whoami`, `list_my_tasks`, `get_task`, `create_note`, and `note_grill`. `note_grill` still returns the interview JSON and still does not insert a note. The Cursor snippet is in [Cursor](#cursor).

## Cursor

Project `.cursor/mcp.json`, or the user file `~/.cursor/mcp.json`. PAT form (no `JWT_SECRET`):

```json
{
  "mcpServers": {
    "morph": {
      "command": "/absolute/path/to/morph-mcp",
      "env": {
        "MORPH_MCP_TOKEN": "<morph_pat_…>",
        "TRAN_SQLITE_PATH": "/absolute/path/to/tran.sqlite"
      }
    }
  }
}
```

A session JWT still works if you also set `JWT_SECRET` to the same non-default secret as the Morph API.

## Environment

| Variable | Required | Purpose |
|----------|----------|---------|
| `MORPH_MCP_TOKEN` | yes | Session JWT or `morph_pat_…`. Never log it. |
| `JWT_SECRET` | JWT only | HMAC secret for a session JWT. Omit it when the token is a PAT. Empty and the development default are refused for JWTs. |
| `TRAN_SQLITE_PATH` | yes for notes and PATs | SQLite file the API already uses. Default `./data/tran.sqlite` if unset. List and get open it read-only. `create_note` needs the file to be writable. |

`DB_PATH` and `ENTITY_DETAILS_BADGER` are not read by this build.
