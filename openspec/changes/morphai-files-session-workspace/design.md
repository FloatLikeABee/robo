## Context

See proposal.md — Why. Today `folderName` / `workspaceFiles` / `pinnedPaths` live in `SkoolAiChat.js` as one global React state. Session switch reloads messages and HybridContext but not Files. Directory handles from `showDirectoryPicker` are not stored, so reload empties Files. `AgentFilesTab.jsx` still shows the tutorial `hybrid-hint`. File System Access directory handles can be stored in IndexedDB; after reload, `queryPermission` / `requestPermission` is required.

## Goals / Non-Goals

**Goals:**

- Per-`sessionId` workspace: folder handle, listing, and pin set.
- Persist handles + recents in IndexedDB; re-walk the tree on restore (do not cache file bodies).
- Recents UI on Files (name list); Open folder still available.
- Delete the tutorial paragraph.

**Non-Goals:**

- Writing to the local folder (still forbidden).
- Syncing folder bytes to Morph servers.
- Embedded MorphNotes (`singleSession`) Files pane.
- Teaching Safari/webkitdirectory to persist handles (no API); recents there may be names only and reopen via picker if a handle cannot be stored.

## Decisions

### 1. IndexedDB for handles, not localStorage

- **Choice:** Store `FileSystemDirectoryHandle` objects in IndexedDB (structured clone). Session map: `sessionId → { folderId, pinnedPaths }`. Recents: ordered unique `{ id, name, handle }` (cap ~10).
- **Why:** Handles cannot live in JSON localStorage. IDB is the File System Access persistence path.
- **Alternative:** Remember folder *name* only — user would re-pick every reload. Rejected.

### 2. Bind workspace to chat session, not the browser tab

- **Choice:** Restore Files from the session map when `sessionId` changes (same effect that already loads messages). Opening a folder updates that session’s map and prepends recents.
- **Why:** “Session history and files comprehending” means each conversation owns its folder/pins.
- **Alternative:** One global sticky folder for all sessions. Rejected — mixing chats with the wrong tree.

### 3. Re-walk on restore; keep pins by path

- **Choice:** On restore, `requestPermission` then walk like Open folder. Pins are path strings; missing files drop from the pin set silently.
- **Why:** File contents change on disk; stale File objects from a previous walk are wrong.

### 4. Recents are global, workspace assignment is per session

- **Choice:** Recents list is shared across sessions (quick choose). Applying a recent sets the *current* session’s workspace.
- **Why:** Same repo is often reused across chats; the assignment still stays per session after pick.

## Risks / Trade-offs

- [Permission prompt after reload] → Expected; if denied, keep the folder name and recents, show a short retry, do not pretend the tree is loaded.
- [Chrome-only durable handles] → Fallback picker unchanged; recents without a handle still call Open folder / webkit input.
- [IndexedDB origin isolation] → Workspace is per origin (localhost vs production); acceptable.

## Migration Plan

- No server migration. First visit: empty recents. Rollback: revert Files UI and stop reading IDB (stale IDB can stay).

## Open Questions

- None.
