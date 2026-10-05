## Why

Agents already store notes on the signed-in Morph user's `user_note_todo` rows (`create_note`, #129). A human still has no view that shows only those agent-authored notes under their own login, and the existing Notes & TODOs list resolves the owner through `tranUserIDFromContext`, which trusts `?user_id=` and can fall back to user id 1.

## What Changes

- Add a signed-in Agent notes page at `/agent-notes` that lists the caller's agent-authored notes (title, time, status) and a readable detail view of the body, with an empty state that explains how agents post notes.
- Add a fail-closed read API for that page. It uses the session subject only. It does not reuse `ListUserNotesTodos`.
- Logged-out visitors are sent to login. The API returns 401 and no note text.
- Phone and desktop layouts stay inside the existing Morph AI shell (safe areas, dark theme, wrapping text). No schema change.

## Capabilities

### New Capabilities

- `morph-agent-notes`: Human browse of agent-authored Morph notes, scoped to the signed-in user, with login required and a readable phone and desktop layout.

### Modified Capabilities

- None. `create_note` and the Notes & TODOs editor stay as they are. The phone-shell specs are reused, not rewritten.

## Impact

- New Go handlers and tests under `morph/handlers`, registered on `/api/tran/agent-notes`.
- New SPA route, page, and styles under `morph/frontend`, plus a header entry on the Morph AI shell.
- Local only: `http://localhost:3031/` proxying `/api` to `:9090`. No Render, MorphUtils, or new agent write tools.
