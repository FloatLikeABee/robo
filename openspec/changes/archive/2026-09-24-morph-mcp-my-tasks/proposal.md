## Why

`morph-mcp` can name the signed-in user (`whoami`) but cannot read any of that user's data, so a Cursor or Claude Desktop session still cannot see the caller's own work. This is the one remaining v1 MCP data story (epic #3, issue #83): two read-only tools, nothing else.

## What Changes

- Add `list_my_tasks` and `get_task` on the existing stdio server. Both are read-only. `whoami` stays. No resources, prompts, HTTP transport, API tokens, or write tools.
- Open `TRAN_SQLITE_PATH` read-only (`mode=ro` and `query_only`) while `morph-api` holds the same file in WAL mode. Do not open Badger and do not call `db.NewTranSQL`.
- "Mine" is the signed-in user's `user_note_todo` rows (Notes & TODOs). MorphNotes Tasks (`CaseTask`) have no per-user owner; exposing that board would return other people's rows. See design.md.
- Fail startup when `JWT_SECRET` is empty or the built-in development default, and when the JWT `sub` is missing from `plat_users`. Re-check token expiry on every tool call.
- Document the tools, `TRAN_SQLITE_PATH`, and a Cursor `mcp.json` example.
- Bundled text fixes: the stdio test line reader must be able to exit if the consumer stops; two BK OpenSpec files still describe the removed BK TCP `tools/list` / a fully green pytest and frontend build.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `morph-mcp-stdio`: The server lists three read-only tools, reads the caller's own notes and todos from SQLite without taking the API locks, and fails closed on a development JWT secret, a deleted user, or an expired token at call time.

## Impact

- `morph/mcp`, `morph/cmd/morph-mcp`, `docs/agents/14-morph-mcp.md`, `morph/cmd/morph-mcp/README.md`.
- `openspec/specs/morph-mcp-stdio/spec.md` (via this delta).
- Text-only: `morph/cmd/morph-mcp/stdio_test.go`, `bk/openspec/changes/remove-tools-merge-adviser-customization/proposal.md`, `bk/openspec/changes/archive/2026-09-24-remove-bk-tcp-mcp/tasks.md`.
- No `morph/frontend/`, no `pkg/morphai/`, no BK product code, no secrets.
- Issues: Closes #83. Part of #3.
