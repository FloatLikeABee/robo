## Why

Agent platforms that speak MCP Apps attach a server by loading a declared UI resource, not by hand-wiring each tool. Dorph already exposes `create_note`, `list_my_tasks`, and `get_task` on the local stdio server. Without a `ui://` notes app, those platforms cannot offer create, list, and get through an MCP app surface.

## What Changes

- Serve one self-contained Dorph notes HTML resource (`ui://dorph/notes`, MIME `text/html;profile=mcp-app`) from `dorph-mcp`.
- Link `list_my_tasks`, `get_task`, and `create_note` to that resource with `_meta.ui.resourceUri` so an MCP-Apps host can render the notes panel. The panel lists, opens, and creates notes by calling those existing tools through the host bridge.
- Advertise the `io.modelcontextprotocol/ui` extension. Keep `resources.listChanged` false. Tool results stay plain JSON for clients that do not render apps.
- Document localhost attach steps for one MCP-Apps-capable client, plus the existing plain-tools config.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `dorph-mcp-stdio`: `resources/list` is no longer empty; it includes the notes app resource. The three note tools declare that UI. Docs add MCP Apps attach steps for local Dorph.

## Impact

- `dorph/mcp` server registration, one embedded HTML file, and tests. `dorph/cmd/dorph-mcp` docs and the agent doc `docs/agents/14-dorph-mcp.md`.
- No change to note storage, `ownerClause`, JWT checks, or tool argument limits. No new dependency. No Dorph SPA, Render, skills, or `pkg/dorphai` work.
