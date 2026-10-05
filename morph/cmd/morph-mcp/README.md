# morph-mcp

Local Model Context Protocol server for Morph. Stdio only. `initialize` advertises tools and resources. Tools: read-only `whoami`, `list_my_tasks`, and `get_task`, plus `create_note` (not read-only). `resources/list` is empty.

`list_my_tasks` and `get_task` return the signed-in user's own Notes & TODOs (`user_note_todo`), not the shared MorphNotes Tasks board. Rows are filtered by the token subject. Optional list filters: `type` (`all`, `note`, `todo`), `status` (`all`, `open`, `done`), `limit` (default 50, max 100). `get_task` takes an integer `id`. Someone else's id is not found.

`create_note` stores one note for that same user. The title starts with `[morph-mcp]` and the body starts with `source: morph-mcp`. Title or body is required (200 and 32000 characters). A bad token writes nothing. The human reads the note in Notes & TODOs at http://localhost:3031/.

HTTP JSON catalogs (`/ai/mcp-tools`, Event Logs `/api/v1/ai/mcp-tools` and `/api/v1/ai/mongodb-mcp`) are Morph AI tool lists. They are not MCP.

Full notes: [`docs/agents/14-morph-mcp.md`](../../../docs/agents/14-morph-mcp.md).

## Build and run

```bash
cd morph && go build -o morph-mcp ./cmd/morph-mcp
```

The client launches the binary. Stdout is protocol messages only. Logs go to stderr. There is no network listener. This process is not served on port 3031 or 9090.

| Variable | Required | Purpose |
|----------|----------|---------|
| `MORPH_MCP_TOKEN` | yes | Morph session JWT from `POST /api/auth/login`. Never log or commit it. |
| `JWT_SECRET` | yes | HMAC secret used to verify that JWT. Must match the API. Empty and the development default are refused. Production also applies the API secret-strength rules. |
| `TRAN_SQLITE_PATH` | yes for notes | SQLite file the API already uses. List and get open it read-only. `create_note` needs it writable. Default `./data/tran.sqlite` if unset. |

A missing or invalid token exits non-zero before any protocol message and does not insert a note. A raw user id is not accepted. Startup also requires the token subject to exist in `plat_users`. Every tool call checks the token again; an expired token is a tool error and `create_note` writes nothing. A user deleted after launch can keep calling tools until the process exits. Stop the process to drop that access. MCP never exposes private data to unauthenticated callers.

This process does not open Badger, so it can run while `morph-api` holds the Badger directory lock. It does not run SQLite migrations or change the journal mode.

## Token for local Morph

The UI is http://localhost:3031/. The API is http://127.0.0.1:9090. The UI dev server proxies `/api` to the API, so either login URL works while that server is up:

```bash
curl -s -X POST http://127.0.0.1:9090/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"<username>","password":"<password>"}'
```

Put the JSON `token` in `MORPH_MCP_TOKEN`. Placeholders only. Do not commit a real token.

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
