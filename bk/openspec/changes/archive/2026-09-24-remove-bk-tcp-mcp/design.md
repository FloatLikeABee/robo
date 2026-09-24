# Design

## Context

See proposal.md for why. Checked against the tree at `d862013` (latest `main`), not the Bridge file list alone.

What is actually wired:

- `MCPService` is constructed in `bk/src/api.py` as `MCPService(self.agent_manager, self.rag_system)`. The optional host manager argument is omitted. `MCPHostManager()` is a separate field used only by `/mcp/hosts` CRUD.
- `start_server` (length-prefixed TCP, default port 8196) runs only from `POST /mcp/start` via `BackgroundTasks`. Process boot does not bind 8196. `BackgroundTasks` is imported solely for that route.
- `bk/frontend/src/pages/MCPHosts.js` exists and calls `/mcp/hosts`, but `App.js` does not import it and `Header.js` has no nav entry. The live caller is `SystemStatus.js`, which queries `api.getMCPHosts` and renders an "MCP Hosts" card.
- `bk/src/models.py` MCP host types are used only by the host manager and those routes.
- `bk/openspec/specs/mcp-tool-registry/spec.md` still requires the registry. Its headings were blank duplicates; they are now named so archive can remove them. `assistant-module` still has a scenario where `MCPService` answers `tools/list` until this change is synced.
- `bk/openspec/changes/mcp-tool-registry-refactor/` is an active change whose tasks are checked, but it only scaffolds the same unfinished surface.
- Docs that claim a working BK MCP: `bk/project_overview.md`, the FastAPI description in `api.py`, `bk/src/help_service.py` (Help RAG bootstrap), `bk/main.py`, `bk/start.sh`, `bk/src/__init__.py`. `bk/README.md` and `docs/agents/06-ai-tools.md` do not. `env.example` and `config.py` do not mention 8196.
- `bk/chat_box_test.py` dials `ws://127.0.0.1:8196/message` with a chat payload. That is not the length-prefixed TCP server, and pytest does not collect the file (`testpaths = tests`).
- ScholarForge comments say "MCP-style" / "MCP controller" for a dual-model writing pipeline. That is not this server.

Planning root is `bk/` (`openspec` nearest root) so the removal delta archives into `bk/openspec/specs/`. The CLI action context limits edits to `/workspace/bk`. `docs/agents/06-ai-tools.md` is outside that root; the issue explicitly asks docs not to claim a working BK MCP and to name `morph/cmd/morph-mcp` where it helps, so that one file is an intentional exception.

## Goals / Non-Goals

**Goals:**

- No import of `mcp_service` or `mcp_host_manager`, no `/mcp/start` or `/mcp/hosts`, no UI caller left pointing at them.
- `api.py` diff is deletion of the MCP import, construction, route block, and the MCP sentences in the OpenAPI description. Leave the authentication paragraph untouched.
- `mcp-tool-registry` main spec no longer requires the registry. Purpose is edited in the main spec (a delta `## Purpose` is ignored for an existing capability).
- Tests prove the modules and routes are gone, written first so they fail while the code is still present.

**Non-Goals:**

- Implementing or importing `morph/cmd/morph-mcp`.
- Auth on the BK API (issue #25).
- Rewriting ScholarForge, RAG, assistants, or agents.
- Deleting `chat_box_test.py` or archive plans that mention other MCP work.
- Adding a frontend route or a 410 tombstone.

## Decisions

### 1. Delete the TCP stack, HTTP routes, models, and UI callers

Remove `mcp_service.py`, `mcp_host_manager.py`, the MCP host models, the route block, the frontend page, the `api.js` methods, and the System status query/card. After the card is gone, the three remaining health tiles use `sm={4}` so the row still fills.

Rationale: the product decision is removal. The host manager is not connected to the server, host tools are placeholders, and the page is already unrouted. A leftover System status query would call a deleted endpoint.

Alternatives rejected:

- **Feature flag defaulting off.** Dead code and a way to turn the fake server back on. The issue says remove, not finish or hide.
- **Finish this TCP server as JSON-RPC MCP.** That duplicates `morph/cmd/morph-mcp` (PR #65, not merged). This change must not depend on that code.
- **Keep routes and return 410.** OpenAPI would still advertise an MCP tag, and the models and UI would remain. Issue #25 is reducing unauthenticated surface; preserving `/mcp/*` fights that.
- **Add a `/mcp-hosts` redirect.** Nothing routes there today. A new redirect would be a new feature.

### 2. Tests assert absence, using the existing module-removal style

`tests/test_assistant_refactor.py` already expects `src.tools` to raise `ModuleNotFoundError`. Add `tests/test_no_bk_tcp_mcp.py` that expects the same for `src.mcp_service` and `src.mcp_host_manager`, asserts the MCP host models are gone, and asserts `api.py` contains neither `/mcp/start`, `/mcp/hosts`, nor `MCPService`.

Do not construct `RAGAPI` inside that test. Importing `src.api` builds Chroma and the whole app; the smoke compile test plus a separate startup check cover "the API still imports." The unit test stays fast and fails for the right reason (modules still importable) before the deletion.

### 3. Spec sync needs unique requirement names, and it cannot drop a scenario

`openspec validate --strict` refuses the removal delta while the live requirements share a blank `### Requirement:` heading (duplicate names). Give them the names the delta removes: Single tool registry, Common tool list interface, Structured tool invocation. That rename does not change the requirement text. Archive then removes those three blocks and adds "No BK TCP MCP". Edit the main spec Purpose in place to say the capability is withdrawn and name `morph/cmd/morph-mcp` (a delta `## Purpose` is ignored).

The same validator refuses a MODIFIED requirement that omits a scenario. Keep the scenario name `MCP tools/list after removal` and change its steps so AI tools does not accept tools/list. Do not delete the scenario heading.

Move `bk/openspec/changes/mcp-tool-registry-refactor/` to `bk/openspec/changes/archive/` with a superseded note in its proposal. Do not merge its placeholder delta into the main spec.

Modify `assistant-module` "Tool system removal" by dropping the `MCPService` tools/list scenario and stating AI tools does not accept tools/list on a BK MCP server. Keep the agent-without-tools scenario.

### 4. Docs and logs stop claiming BK MCP; unrelated "MCP" strings stay

Update the files listed in Context that claim a working BK MCP. Add one sentence to `bk/README.md` and `docs/agents/06-ai-tools.md` naming `morph/cmd/morph-mcp` and not linking `docs/agents/14-morph-mcp.md` (that page is not on `main`).

Leave ScholarForge "MCP-style" wording, `docs/archive/`, and `docs/agents/04-formx.md`. Leave `chat_box_test.py` and its port 8196. A verification grep for the removed modules and routes must be clean; a grep for the substring `MCP` or `8196` will still hit those leftovers on purpose.

### 5. No data migration

`mcp_hosts.json` under the data directory, if an operator created hosts, becomes an unused file. Do not delete it at startup. Rollback is reverting the commit.

## Risks / Trade-offs

- [System status still calls `/mcp/hosts`] → Remove the query and the card in the same change as the route. Confirmed it is the only routed caller.
- [OpenAPI description still markets MCP] → Delete only the MCP paragraphs and the MCP tag bullet. Do not edit the "no authentication" paragraph (issue #25).
- [Blank spec headings make archive refuse the delta] → Name the three requirements before archive. After sync the main spec contains only "No BK TCP MCP" and a rewritten Purpose. The tools/list scenario heading stays; its steps no longer call `MCPService`.
- [`BackgroundTasks` import left unused] → Drop that name from the FastAPI import. It has no other use. Do not reorder the rest of the import.
- [Grep for `8196` looks incomplete] → `chat_box_test.py` stays; the PR says why.
- [Help RAG collection already ingested the MCP Hosts bullet] → Bootstrap text changes for new ingests. Existing `system_help` chunks are not rewritten here; the source of the lie is the code string.
- [Planning-root scope vs `docs/agents/`] → One-line edit outside `bk/` is required by the issue. No other files outside `bk/` are edited.
- [`api.py` merge with auth work] → Deletions only, plus the MCP sentences in the long description. No new middleware, no route reorder.

## Migration Plan

1. Add the failing absence tests and confirm they fail because the modules still import.
2. Delete the Python modules, models, routes, and construction. Adjust imports and the OpenAPI description.
3. Delete the frontend page and API methods; fix System status.
4. Update the docs and log lines listed above. Archive the old OpenSpec change with a superseded note.
5. Confirm tests pass, the API module imports, the frontend production build succeeds, and a grep finds no remaining references to the removed modules or routes.
6. On archive of this change, sync the deltas (see Decision 3) and move this change folder to `bk/openspec/changes/archive/`.

Rollback: revert the commit. No schema change.

## Open Questions

None. The port-8196 chat script and ScholarForge wording were checked and intentionally kept.
