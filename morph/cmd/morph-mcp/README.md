# morph-mcp

Local Model Context Protocol server for Morph. Stdio only. Read-only. `initialize` advertises tools and resources. Tools: `whoami`, `list_my_tasks`, and `get_task`. `resources/list` is empty.

`list_my_tasks` and `get_task` return the signed-in user's own Notes & TODOs (`user_note_todo`), not the shared MorphNotes Tasks board. Rows are filtered by the token subject. Optional list filters: `type` (`all`, `note`, `todo`), `status` (`all`, `open`, `done`), `limit` (default 50, max 100). `get_task` takes an integer `id`. Someone else's id is not found.

HTTP JSON catalogs (`/ai/mcp-tools`, Event Logs `/api/v1/ai/mcp-tools` and `/api/v1/ai/mongodb-mcp`) are Morph AI tool lists. They are not MCP.

Full notes: [`docs/agents/14-morph-mcp.md`](../../../docs/agents/14-morph-mcp.md).

## Build and run

```bash
cd morph && go build -o morph-mcp ./cmd/morph-mcp
```

The client launches the binary. Stdout is protocol messages only. Logs go to stderr. There is no network listener.

| Variable | Required | Purpose |
|----------|----------|---------|
| `MORPH_MCP_TOKEN` | yes | Morph session JWT from `POST /api/auth/login`. Never log or commit it. |
| `JWT_SECRET` | yes | HMAC secret used to verify that JWT. Must match the API. Empty and the development default are refused. Production also applies the API secret-strength rules. |
| `TRAN_SQLITE_PATH` | yes for task tools | SQLite file the API already uses. Opened read-only. Default `./data/tran.sqlite` if unset. |

A missing or invalid token exits non-zero before any protocol message. A raw user id is not accepted. Startup also requires the token subject to exist in `plat_users`. Every tool call checks the token again; an expired token is a tool error. A user deleted after launch can keep calling tools until the process exits. Stop the process to drop that access. MCP never exposes private data to unauthenticated callers.

This process does not open Badger, so it can run while `morph-api` holds the Badger directory lock. It does not run SQLite migrations.

## Cursor (`mcp.json`)

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

## Claude Desktop (`claude_desktop_config.json`)

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
