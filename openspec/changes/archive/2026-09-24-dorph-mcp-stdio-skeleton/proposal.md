## Why

External agents (Cursor, Claude Desktop) cannot attach to Dorph over the Model Context Protocol. The repo's HTTP JSON catalogs (`/ai/mcp-tools` and the Event Logs mongodb-mcp routes) are Dorph AI tool lists, not MCP. Story #58 needs a spec-compliant stdio server in `dorph/` so later read-only DorphNotes tools have a real transport.

## What Changes

- Add `dorph/cmd/dorph-mcp`, a stdio-only MCP server built with the official Go SDK, pinned.
- Add a shared `dorph/mcp` package (server, tool registration, identity) that a later HTTP transport in the API process can reuse.
- Complete `initialize` and `tools/list`. Advertise tools and resources capabilities. Register one read-only `whoami` tool. Do not register DorphNotes tools or `dorph://` resources.
- Resolve the acting Dorph user from a session JWT in the environment. Fail startup on a missing or invalid token. Do not log the token.
- Do not open Badger or SQLite, so the process can start while `dorph-api` holds the Badger directory lock.
- Document build, env, and a Cursor `mcp.json` example. State that the HTTP catalogs are not MCP.

## Capabilities

### New Capabilities

- `dorph-mcp-stdio`: Local stdio MCP handshake, empty resources capability, read-only whoami, and JWT identity without opening Dorph stores.

### Modified Capabilities

- (none)

## Impact

- New Go package `dorph/mcp` and command `dorph/cmd/dorph-mcp`.
- `dorph/go.mod`: pin `github.com/modelcontextprotocol/go-sdk` v1.8.0 (pulls `github.com/golang-jwt/jwt/v5` v5.3.1).
- Docs: `docs/agents/14-dorph-mcp.md`, short pointers from the Dorph and build guides, `dorph/cmd/dorph-mcp/README.md`.
- No Dorph frontend, no `pkg/dorphai`, no network listener, no secrets, no BK TCP MCP, no Engi project tools.
- Issue: Closes #58. Epic: #3.
