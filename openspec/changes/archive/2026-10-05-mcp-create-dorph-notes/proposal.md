# Proposal

## Why

An agent that already holds a Dorph session can list and read that user's Notes & TODOs over stdio MCP, but it cannot leave a work note behind. The owning human reviews those notes in local Dorph (`http://localhost:3031/`, Notes & TODOs). The agent needs a create tool that stores the note on that same user, fails closed without a valid session, and leaves a marker the human can see later.

## What Changes

- Add a `create_note` MCP tool that inserts one `user_note_todo` row (`item_type` `note`) owned by the verified session's active Tran user.
- Keep `whoami`, `list_my_tasks`, and `get_task`. List and get return the created note's id and content. Read tools stay read-only. `create_note` is not read-only.
- Mark agent-created notes with a visible `[dorph-mcp]` title prefix and a `source: dorph-mcp` body line. No schema change.
- Fail closed: missing or invalid `DORPH_MCP_TOKEN` never inserts a row. A session with no active Tran user does not fall back to user id 1.
- Document the local stdio connection: Cursor and Claude `mcp.json`, `DORPH_MCP_TOKEN`, `JWT_SECRET`, `TRAN_SQLITE_PATH`, how to obtain the token, and that the human UI is `http://localhost:3031/` while the API is `http://127.0.0.1:9090`.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `dorph-mcp-stdio`: The server is no longer read-only. `tools/list` includes `create_note`. Writes are owner-scoped, attributed, and rejected without a valid session. The documented client config covers local Dorph at `http://localhost:3031/`.

## Impact

- `dorph/mcp`, `dorph/cmd/dorph-mcp`, `docs/agents/14-dorph-mcp.md`, `dorph/cmd/dorph-mcp/README.md`, and the one-line MCP summary in `dorph/README.md`.
- The read pool stays `mode=ro` and `query_only`. Creates use a separate `mode=rw` connection on the same `TRAN_SQLITE_PATH`. No Badger, no `db.NewTranSQL`, no migration, no journal-mode change.
- Out of scope: the human SPA, skills and commands, Render, and `pkg/dorphai`.
