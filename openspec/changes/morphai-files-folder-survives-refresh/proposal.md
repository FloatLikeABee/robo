## Why

Morph AI Files still shows an empty workspace after refresh. Last-session restore does not help: opening a folder often never binds (webkit picker sends no directory handle, then the handler **clears** the session binding), and even a bound folder disappears when recents are capped or the live handle needs permission again. The last opened folder must still be there after reload.

## What Changes

- Persist the opened folder **name and file listing** on the current chat session, including when there is no File System Access directory handle (Safari / Firefox / Chrome fallback).
- Stop clearing the session’s folder binding on a successful open that has no handle.
- On refresh, restore that folder workspace (name + last listing). Show Reconnect only when a stored handle needs permission again—not an empty “No folder open.”
- Keep Close folder as the way to forget. Embedded / `singleSession` unchanged.

## Capabilities

### New Capabilities

- `morphai-files-folder-persist`: The Files folder opened for a Morph AI chat session survives page refresh with its name and listing, without requiring a live directory handle.

### Modified Capabilities

- (none — `morphai-folder-workspace-session` lives only under `openspec/changes/`, not `openspec/specs/`)

## Impact

- `morph/frontend/src/SkoolAiChat.js` (`handleFolderOpened` must not clear on null handle)
- `morph/frontend/src/components/chat/AgentFilesTab.js` (webkit open still counts as a workspace)
- `morph/frontend/src/lib/filesWorkspaceStore.js` (session snapshot of name + files; load uses snapshot when handle is missing or not granted)
- Browser IndexedDB only. No Go API or SQLite change.
