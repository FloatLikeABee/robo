# dorph-mcp

Local Model Context Protocol server for Dorph. Stdio only. `initialize` advertises tools, resources, and the MCP Apps extension `io.modelcontextprotocol/ui`. Tools include `whoami`, unified notes (`create_note`, `get_note`, `list_notes`, `update_note`, `append_to_note`, `archive_note`, `search_my_notes`), `note_grill`, the previous module names as aliases, event logs, and read-only `suggest_note_kind` and `dorph_guide`. `resources/list` returns `ui://dorph/notes`, `ui://dorph/grill` (`text/html;profile=mcp-app`), and `dorph://guide`. `resources.listChanged` is not advertised. `note_grill` does not insert a note.

`DORPH_MCP_TOKEN` is a session JWT or a `dorph_pat_` personal access token. Mint a PAT with `POST /api/agent-tokens` (session JWT required). A PAT does not need `JWT_SECRET` in the client env. Revoked, expired, and deleted-user tokens fail the next call.

`create_note` takes optional `kind` (`note`, `todo`, `big_note`, `timeline`, `research`, `stick`, `task`). Omitted kind stays a Notes & TODOs note and returns an integer `id`. Optional `idempotency_key` and `tags`. Unified get/update/append/archive take `kind:rowid`. Every kind is owner-scoped. The admin CaseTask HTTP list stays unfiltered.

HTTP JSON catalogs (`/ai/mcp-tools`, Event Logs `/api/v1/ai/mcp-tools` and `/api/v1/ai/mongodb-mcp`) are Dorph AI tool lists. They are not MCP.

Full notes: [`docs/agents/14-dorph-mcp.md`](../../../docs/agents/14-dorph-mcp.md).

## Build and run

```bash
cd dorph && go build -o dorph-mcp ./cmd/dorph-mcp
```

The client launches the binary. Stdout is protocol messages only. Logs go to stderr. There is no network listener. This process is not served on port 3031 or 9090.

| Variable | Required | Purpose |
|----------|----------|---------|
| `DORPH_MCP_TOKEN` | yes | Session JWT or `dorph_pat_…`. Never log or commit it. |
| `JWT_SECRET` | JWT only | HMAC secret used to verify a session JWT. Omit it for a PAT. Empty and the development default are refused for JWTs. |
| `TRAN_SQLITE_PATH` | yes for notes and PATs | SQLite file the API already uses. List and get open it read-only. `create_note` needs it writable. Default `./data/tran.sqlite` if unset. |

A missing or invalid token exits non-zero before any protocol message and does not insert a note. A raw user id is not accepted. Startup also requires the token subject to exist in `plat_users`. Every tool call checks the token again; an expired or revoked token is a tool error and `create_note` writes nothing. A user deleted after launch fails the next call. MCP never exposes private data to unauthenticated callers.

This process does not open Badger, so it can run while `dorph-api` holds the Badger directory lock. It does not run SQLite migrations or change the journal mode.

## Token for local Dorph

The UI is http://localhost:3031/. The API is http://127.0.0.1:9090. The UI dev server proxies `/api` to the API, so either login URL works while that server is up:

```bash
curl -s -X POST http://127.0.0.1:9090/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"<username>","password":"<password>"}'
```

Put the JSON `token` in `DORPH_MCP_TOKEN`. Placeholders only. Do not commit a real token.

## Cursor (`mcp.json`)

PAT (no `JWT_SECRET`):

```json
{
  "mcpServers": {
    "dorph": {
      "command": "/absolute/path/to/dorph-mcp",
      "env": {
        "DORPH_MCP_TOKEN": "<dorph_pat_…>",
        "TRAN_SQLITE_PATH": "/absolute/path/to/tran.sqlite"
      }
    }
  }
}
```

A session JWT still works if you also set `JWT_SECRET`. Mint a PAT with:

```bash
curl -s -X POST http://127.0.0.1:9090/api/agent-tokens \
  -H "Authorization: Bearer <session JWT>" \
  -H 'Content-Type: application/json' \
  -d '{"name":"cursor"}'
```

## MCP Apps (Claude Desktop)

`list_my_tasks`, `get_task`, and `create_note` point at `ui://dorph/notes`. `note_grill` points at `ui://dorph/grill`. Claude Desktop renders those pages in a sandboxed iframe. The notes page lists notes, opens one, and creates one by calling those tools through the host. The grill page drafts a body and calls `create_note` only after that draft is shown. The pages do not receive the token. The MCP Apps overview also names Claude, VS Code GitHub Copilot, Microsoft 365 Copilot, Goose, Postman, MCPJam, and Archestra.AI as renderers of this same stdio server.

Local Dorph is http://localhost:3031/. The API is http://127.0.0.1:9090. Get the token with the login request in [Token for local Dorph](#token-for-local-dorph), then put these three variables in `claude_desktop_config.json`. Placeholders only.

```json
{
  "mcpServers": {
    "dorph": {
      "command": "/absolute/path/to/dorph-mcp",
      "env": {
        "DORPH_MCP_TOKEN": "<session JWT>",
        "JWT_SECRET": "<same non-default secret as the Dorph API>",
        "TRAN_SQLITE_PATH": "/absolute/path/to/tran.sqlite"
      }
    }
  }
}
```

Restart Claude Desktop and ask it to open Dorph notes. A client that does not render MCP Apps, including Cursor with the snippet above, still receives the plain tool results.
