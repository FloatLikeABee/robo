# Proposal

## Why

AI tools (`bk/`) advertises a Model Context Protocol server that is a custom length-prefixed TCP listener (default port 8196), not JSON-RPC 2.0 MCP. Its host manager is never passed into `MCPService`, and host tools are placeholders. Shipping it markets a fake protocol surface. The real MCP server is the Morph Go binary `morph/cmd/morph-mcp` (separate work; this change does not depend on that code).

## What Changes

- **BREAKING**: Delete the TCP server (`bk/src/mcp_service.py`) and host manager (`bk/src/mcp_host_manager.py`). The BK API no longer constructs them or starts port 8196.
- **BREAKING**: Remove HTTP routes `POST /mcp/start` and `/mcp/hosts` CRUD, the MCP host Pydantic models, and the frontend callers (`MCPHosts` page, System status card, `api.js` MCP methods).
- Stop docs, startup logs, OpenAPI text, and the Help bootstrap from claiming a working BK MCP. Where a pointer helps, name `morph/cmd/morph-mcp` without linking a doc that is not on `main`.
- Mark the `mcp-tool-registry` spec removed/superseded, and retire the unfinished `mcp-tool-registry-refactor` change. Drop the assistant-module scenario that still requires `MCPService`.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `mcp-tool-registry`: Remove the registry, discovery, and invocation requirements. AI tools does not host a TCP MCP server or external MCP host manager. Operators use `morph/cmd/morph-mcp` for protocol MCP.
- `assistant-module`: The tool-removal requirement no longer tells `MCPService` to answer `tools/list`. That service is gone.

## Impact

- BK API (`bk/src/api.py`) loses MCP imports, construction, and routes. Keep the diff to that removal so a later auth change on the same file stays mergeable. RAG, assistants, and agents stay.
- BK frontend: delete the unused MCP hosts page and stop System status from calling `/mcp/hosts`.
- BK OpenSpec under `bk/openspec/` (this planning root). `docs/agents/06-ai-tools.md` gets a one-line pointer even though this change's planning root is `bk/`.
- Out of scope: Morph Go MCP implementation, `morph/frontend/`, `pkg/morphai/`, BK API authentication, ScholarForge's "MCP-style" dual-model pipeline (not this TCP server).
