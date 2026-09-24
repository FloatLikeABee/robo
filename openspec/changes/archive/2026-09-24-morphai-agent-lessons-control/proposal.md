## Why

Morph AI distills session lessons and injects them into later chats, but operators cannot turn a lesson off or delete it, and every lesson is global. A bad rule keeps steering every account. Settings needs an API for the current user's lessons before the Agent screen can be built.

## What Changes

- Add `enabled` (default true) and `owner_user_id` on `agent_lesson`, with a one-shot claim for legacy rows when the database has exactly one account.
- Add `GET` / `PATCH` / `DELETE /api/agent-lessons` scoped to the caller. Another user's lesson is 404.
- Inject and list on `GET /api/skills` only enabled lessons for a verified bearer user. The lesson list endpoint includes disabled lessons so the settings screen can toggle them.
- Ignore client `X-User-ID` for lesson reads, writes, prompt injection, and harvest. **BREAKING** for any caller that listed or relied on global lessons without a bearer token.

## Capabilities

### New Capabilities

- `agent-lesson-control`: Per-user lesson ownership, the enabled flag, operator list/toggle/delete, and prompt injection of only that user's enabled lessons.

### Modified Capabilities

- None. `openspec/specs/` has no agent-lesson baseline yet.

## Impact

- `morph/db` SQLite schema and `agent_lesson` queries
- `morph/handlers/agent_lesson_api.go`, harvest, `skills.go`, chat prompt assembly
- `docs/agents/03-morph.md`, `docs/agents/02-ai-integration.md`, `morph/README.md`
- Go tests in `morph/db` and `morph/handlers`
- No frontend in this change. Auth middleware stays as it is (issue #22).
