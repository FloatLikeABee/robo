## Why

Outside agents still attach to Morph with a session JWT and `JWT_SECRET` in the client env, and there is no revoke list (#134). Stick notes and CaseTask rows have no owner, so they cannot be part of a per-user note surface. The 20 existing MCP tools sprawl by module instead of by kind.

## What Changes

- Personal access tokens: create, list, and revoke via authenticated HTTP API (no Settings UI). Tokens are stored hashed, carry scopes, and expire. `morph-mcp` accepts `MORPH_MCP_TOKEN=morph_pat_…` without `JWT_SECRET` in the client env. A revoked, expired, or deleted-user token fails on the next tool call. Legacy session JWTs still work when `JWT_SECRET` is set.
- Unified own-notes tools `create_note`, `get_note`, `list_notes`, `update_note`, `append_to_note`, `archive_note` taking `kind` (`note`, `todo`, `big_note`, `timeline`, `research`, `stick`, `task`). Old tool names stay as aliases for one phase. Creates accept an idempotency key. Updates accept `expected_version` and return `conflict` on mismatch.
- `owner_user_id` on `CaseTask` (the stick board and the MorphNotes task board are the same table). Migration backfills existing rows to the first Admin plat user. MCP access is owner-scoped. The admin UI keeps listing every row and does not break.
- Tags plus `search_my_notes` (SQLite FTS5, owner-only).
- `note_grill` accepts every kind. Draft text matches the kind.
- Operator doc `docs/agents/14-morph-mcp.md` and the morph-mcp README: PAT curl, scopes, unified tools. Do not touch `docs/agents/16-morph-for-agents.md` or the skills pack.

## Capabilities

### New Capabilities

- `agent-tokens`: Session users create, list, and revoke hashed personal access tokens with scopes. morph-mcp verifies a PAT from SQLite without a client-side JWT secret.

### Modified Capabilities

- `morph-mcp-stdio`: Accept PATs; unified kind tools with aliases; owner-scoped stick/task; tags and owner-only FTS search; whoami reports scopes and expiry; a deleted or revoked identity fails the next call.
- `note-ai-grill`: `note_grill` start accepts every unified kind, not only note and todo.

## Impact

- Schema in `morph/db` (`agent_token`, `note_tag`, `note_idempotency`, `my_notes_fts`, CaseTask owner/version/archive columns). morph-mcp still does not run migrations.
- HTTP: `POST/GET/DELETE /api/agent-tokens` (session JWT). CaseTask create writes `owner_user_id` when the session is present.
- `morph/mcp` identity, tools, and store. `morph/grill` kind list and prompts.
- Docs: `docs/agents/14-morph-mcp.md`, `morph/cmd/morph-mcp/README.md`.
- Tests: owner-scope for every kind, PAT hash/revoke/expiry/scope denial, FTS isolation, CaseTask UI list still works.
- Out of scope: pool, share, OAuth, HTTP MCP, Settings UI, `morph_guide`, `suggest_note_kind`, Echo's agent guide and skills pack.
