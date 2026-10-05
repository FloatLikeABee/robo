## Why

Morph AI's management tool loop (`chatWithManagementTools`) runs POST, PUT, PATCH, and DELETE calls as soon as the model emits them. A user can lose or change data with no chance to stop the call. Issue #34 (epic #7, E3-S3) requires those writes to pause for the owning user's approval before any approval-card UI (#55) or selectable presets (#62) exist.

## What Changes

- Classify each management tool call. Reads, including a small allowlist of read-only POSTs, still execute immediately. Writes do not.
- Persist one pending approval for the owning user and chat session, skip execution, and return `pending_approval` on the chat response with an explanation that the assistant is waiting.
- Add `POST /api/chat/approvals/:id/decide`. Approve executes that call once and resumes the loop. Deny cancels it and does not execute.
- Keep the default policy as ask-before-every-write. Classification stays method plus allowlist. A separate ask-or-allow policy hook is the extension point for a later trusted preset. Catalog `side_effect` metadata is not read in this change.
- **BREAKING:** A management write no longer finishes inside the same `POST /api/chat` response. Clients receive a wait explanation plus `pending_approval` instead of the write's tool result.

## Capabilities

### New Capabilities

- `agent-write-approvals`: Classify management tool writes, persist a user-and-session approval, return it on chat, and decide approve or deny with a single execution.

### Modified Capabilities

- (none — `openspec/specs/` has no existing requirement for tool-loop approvals)

## Impact

- Morph AI Go API: `morph/handlers/management_chat.go`, `morph/handlers/chat.go`, `morph/handlers/register_routes.go`, a new approval helper beside the tool loop, `morph/models` chat types, and a SQLite table in `morph/db`. `execManagementAPI` stays the executor; this change only chooses when it runs.
- Chat transcript: the waiting assistant message stores `pending_approval` so a later reload can show the same decision.
- Out of this change: approval cards (#55), presets (#62), helper runtime (#35/#56), authz middleware (#22/#23), and the multi-provider client (PR #71).
