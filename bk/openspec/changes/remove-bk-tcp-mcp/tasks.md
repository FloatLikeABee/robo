# Tasks

## 1. Backend removal

- [x] 1.1 Add `bk/tests/test_no_bk_tcp_mcp.py` that expects `src.mcp_service` and `src.mcp_host_manager` to raise `ModuleNotFoundError`, expects the MCP host models to be absent from `src.models`, and expects `bk/src/api.py` not to contain `/mcp/start`, `/mcp/hosts`, or `MCPService`. Run the new test and confirm it fails because those modules still import.
- [x] 1.2 Delete `bk/src/mcp_service.py` and `bk/src/mcp_host_manager.py`, delete the MCP host models in `bk/src/models.py`, and remove the MCP import, `BackgroundTasks` name, construction, route block, and MCP sentences in the OpenAPI description in `bk/src/api.py`. Leave the authentication paragraph and ScholarForge wording. Re-run `bk/tests/test_no_bk_tcp_mcp.py` and confirm it passes.

## 2. Frontend callers

- [x] 2.1 Delete `bk/frontend/src/pages/MCPHosts.js`, remove the MCP methods from `bk/frontend/src/services/api.js`, and remove the MCP hosts query and card from `bk/frontend/src/pages/SystemStatus.js` (remaining health tiles `sm={4}`). Confirm `App.js` and `Header.js` stay unchanged and a search of `bk/frontend/src` finds no `getMCPHosts`, `startMCPServer`, or `MCPHosts`.

## 3. Docs and superseded OpenSpec change

- [x] 3.1 Stop claiming a working BK MCP in `bk/project_overview.md`, `bk/src/help_service.py`, `bk/main.py`, `bk/start.sh`, and `bk/src/__init__.py`. Add a short pointer to `morph/cmd/morph-mcp` in `bk/README.md` and `docs/agents/06-ai-tools.md` without linking `docs/agents/14-morph-mcp.md`. Confirm those files no longer say the BK API starts an MCP server.
- [x] 3.2 Move `bk/openspec/changes/mcp-tool-registry-refactor/` to `bk/openspec/changes/archive/` and note in its proposal that removal superseded it. Keep `bk/openspec/config.yaml`. The live `mcp-tool-registry` headings are already named so archive can remove them. Confirm `openspec list --json` from `bk/` no longer lists `mcp-tool-registry-refactor` as an active change.

## 4. Integration check

- [ ] 4.1 From `bk/`, run the existing pytest suite and confirm it passes, import the API app, and confirm a search of the repo finds no `mcp_service`, `mcp_host_manager`, `/mcp/start`, `/mcp/hosts`, or `MCPHosts` except this change's own specs and the archived refactor note. Run `npm run build` in `bk/frontend` and confirm exit 0.
