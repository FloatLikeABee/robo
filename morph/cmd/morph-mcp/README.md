# morph-mcp

Local Model Context Protocol server for Morph. Stdio only. Read-only. The only tool today is `whoami`, which returns the Morph user from a verified session JWT.

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
| `JWT_SECRET` | must match the API | HMAC secret used to verify that JWT. |

A missing or invalid token exits non-zero before any protocol message. A raw user id is not accepted.

This process does not open Badger or SQLite, so it can run while `morph-api` holds the Badger directory lock.

## Cursor (`mcp.json`)

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

## Claude Desktop (`claude_desktop_config.json`)

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
