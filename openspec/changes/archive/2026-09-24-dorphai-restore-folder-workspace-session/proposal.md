> **Superseded** by `dorphai-drop-files-workspace`. This change specified Dorph AI's IndexedDB Files workspace (local folder, pins, Files tab). That behavior was removed. Do not re-implement it, and do not sync these delta specs into `openspec/specs/`.
> Last-chat restore without a folder binding remains current and is restated by `dorphai-drop-files-workspace`. Folder-bound restore and folder-derived titles do not.

## Why

Opening a local folder in Dorph AI Files already binds it to the current chat in IndexedDB, but Dorph AI always starts on the `default` session. After a reload or a later visit, the operator lands on an empty chat instead of the folder workspace they were using. A folder plus its chat should be one workspace session that comes back when they enter Dorph AI.

## What Changes

- Treat an opened Files folder as that chat session’s workspace: keep the folder bound, and remember that session as the last workspace session.
- Persist the last active chat `sessionId` so Dorph AI restores that session on the next visit (same browser origin).
- On enter, if that session still has a folder binding, open the right workspace on **Files** with that folder (or the existing reconnect prompt) without picking the folder again.
- If the session title is still generic (`New chat` / `Chat`), set it to the folder name so the session rail shows the workspace.
- Embedded / `singleSession` DorphNotes chat stays without this Files workspace.

## Capabilities

### New Capabilities

- `dorphai-folder-workspace-session`: A Dorph AI chat with an opened folder is a workspace session that is restored (session, Files pane, folder) the next time the operator enters Dorph AI.

### Modified Capabilities

- (none — no archived baselines under `openspec/specs/`)

## Impact

- `dorph/frontend/src/SkoolAiChat.js` (last session restore, folder-open → workspace session)
- `dorph/frontend/src/lib/filesWorkspaceStore.js` (existing session↔folder binding)
- `PUT /api/chat/sessions/:id` (optional title = folder name)
- Browser `localStorage` + IndexedDB only. No new API or SQLite schema.
