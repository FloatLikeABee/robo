## Why

External agents (Cursor, Claude Desktop) cannot attach to Morph over the Model Context Protocol. The repo's HTTP JSON catalogs (`/ai/mcp-tools` and the Event Logs mongodb-mcp routes) are Morph AI tool lists, not MCP. Story #58 needs a spec-compliant stdio server in `morph/` so later read-only MorphNotes tools have a real transport.

## What Changes

- Add `morph/cmd/morph-mcp`, a stdio-only MCP server built with the official Go SDK, pinned.
- Add a shared `morph/mcp` package (server, tool registration, identity) that a later HTTP transport in the API process can reuse.
- Complete `initialize` and `tools/list`. Advertise tools and resources capabilities. Register one read-only `whoami` tool. Do not register MorphNotes tools or `morph://` resources.
- Resolve the acting Morph user from a session JWT in the environment. Fail startup on a missing or invalid token. Do not log the token.
- Do not open Badger or SQLite, so the process can start while `morph-api` holds the Badger directory lock.
- Document build, env, and a Cursor `mcp.json` example. State that the HTTP catalogs are not MCP.

## Capabilities

### New Capabilities

- `morph-mcp-stdio`: Local stdio MCP handshake, empty resources capability, read-only whoami, and JWT identity without opening Morph stores.

### Modified Capabilities

- (none)

## Impact

- New Go package `morph/mcp` and command `morph/cmd/morph-mcp`.
- `morph/go.mod`: pin `github.com/modelcontextprotocol/go-sdk` v1.8.0 (pulls `github.com/golang-jwt/jwt/v5` v5.3.1).
- Docs: `docs/agents/14-morph-mcp.md`, short pointers from the Morph and build guides, `morph/cmd/morph-mcp/README.md`.
- No Morph frontend, no `pkg/morphai`, no network listener, no secrets, no BK TCP MCP, no Engi project tools.
- Issue: Closes #58. Epic: #3.
