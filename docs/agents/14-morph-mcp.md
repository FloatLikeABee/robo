# 14 — Morph MCP (stdio)

`morph-mcp` is a local [Model Context Protocol](https://modelcontextprotocol.io/specification/2025-06-18) server. Clients such as Cursor and Claude Desktop launch it and speak JSON-RPC over stdin/stdout.

This is not the HTTP JSON tool catalogs already in the repo. Those catalogs describe Morph AI's internal tool loop. They are not MCP:

- Content Maker: `GET /ai/mcp-tools`
- Event Logs: `GET /api/v1/ai/mcp-tools`, `GET /api/v1/ai/mongodb-mcp`, `POST /api/v1/ai/mongodb-mcp/call`

`morph-mcp` speaks MCP revision **2025-06-18**. The pinned Go SDK (`github.com/modelcontextprotocol/go-sdk` v1.8.0) may negotiate a newer revision the client asks for, including 2026-07-28. The server is stdio only. It does not listen on a port. `initialize` advertises tools and resources. The only tool in this build is `whoami` (read-only). `resources/list` is empty. MorphNotes tools and `morph://` resources are a later story.

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

The server trusts the verified claims (`sub`, email, username, roles). It does not look up `plat_users`. A later story should confirm the subject still exists.

On startup the binary loads the repo-root `.env` (same helper as the API) without overriding variables that are already set. A desktop client often has a clean environment, so set `JWT_SECRET` and `MORPH_MCP_TOKEN` in the MCP config.

Obtain a token (placeholders only):

```bash
curl -s -X POST http://127.0.0.1:9090/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"<username>","password":"<password>"}'
```

## Data access while morph-api is running

`morph-api` opens Badger at `DB_PATH` (default `./data/badger`) and `ENTITY_DETAILS_BADGER` (default `./data/entity_details`). Badger takes an exclusive directory lock. A second process that opens either directory fails with a directory-lock error and must not be pointed at those paths while the API is up.

`morph-mcp` does not open Badger or SQLite, so it can start beside a running API and does not take those locks.

The API opens `TRAN_SQLITE_PATH` (default `./data/tran.sqlite`) in WAL mode. Later read-only tools should open that file with `mode=ro` and `query_only=1`. They must not call `db.NewTranSQL`: that helper sets the journal mode and runs schema writes. `db.New` and `db.NewBadgerEntityDetails` stay in the API process.

## Cursor

Project `.cursor/mcp.json`, or the user file `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "morph": {
      "command": "/absolute/path/to/morph-mcp",
      "env": {
        "MORPH_MCP_TOKEN": "<session JWT>",
        "JWT_SECRET": "<same secret as the Morph API>"
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
        "JWT_SECRET": "<same secret as the Morph API>"
      }
    }
  }
}
```

## Environment

| Variable | Required | Purpose |
|----------|----------|---------|
| `MORPH_MCP_TOKEN` | yes | Morph session JWT. Never log it. |
| `JWT_SECRET` | yes, unless the dev default matches the API | HMAC secret for that JWT. Must match morph-api. |

`DB_PATH`, `ENTITY_DETAILS_BADGER`, and `TRAN_SQLITE_PATH` are not read by this build.
