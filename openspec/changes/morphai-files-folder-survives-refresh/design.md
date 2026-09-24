## Context

See proposal.md — Why. `morphai-restore-folder-workspace-session` restores which chat is active. Folder bytes still depend on a cloneable `FileSystemDirectoryHandle` in IndexedDB. Safari/Firefox (and Chrome when `showDirectoryPicker` fails) open via `<input webkitdirectory>` and pass `directoryHandle: null`. `handleFolderOpened` then **clears** the session binding. Recents are capped at 10, so a session’s `folderId` can point at a deleted recent and `loadWorkspaceForSession` returns an empty workspace. After refresh the operator sees “No folder open.”

## Goals / Non-Goals

**Goals:**

- Persist folder name + file metadata on the **session** row so refresh restores Files without a live handle.
- Successful open never calls `clearSessionBinding` (only Close folder does).
- `loadWorkspaceForSession` uses the session snapshot when the recent/handle is missing or not granted; Reconnect only if a handle exists but is unreadable.
- Pure snapshot helpers + unit tests (no real directory handle required).

**Non-Goals:**

- Persisting file *contents* or webkit `File` blobs (they do not survive refresh). Pins still need a live read or reconnect to send bodies to the agent.
- New Go/SQLite APIs or IndexedDB version bump if extra fields on the existing sessions store suffice.
- Changing last-session id restore (already shipped).
- Embedded / `singleSession` Files workspace.

## Decisions

### 1. Snapshot on the session binding, not only recents

**Choice:** Extend the existing sessions object-store row with `folderName` and `files: [{ path, size, skipped }]`. Keep `folderId` + recents handle when available. `loadWorkspaceForSession` prefers a live walk when permission is granted; otherwise returns snapshot name/files. If the recent row is gone, still use the session snapshot.

**Why:** Recents cap and missing handles are why refresh looks empty. The session row is the workspace source of truth.

**Alternative:** localStorage JSON of the listing — quota/size worse for large trees; IDB already used.

### 2. Open without handle still binds

**Choice:** `rememberFolderByName(name)` (or `rememberDirectory` accepting a name-only open) creates/updates a recent `{ id, name, openedAt }` without a handle. `handleFolderOpened` always `setSessionBinding` with snapshot after a successful list. **Do not** `clearSessionBinding` when `directoryHandle` is null.

**Why:** Webkit path is a successful open, not a close.

**Alternative:** Force Chromium-only `showDirectoryPicker` — fails this user on refresh in browsers without it, and still fails when the picker throws and falls back.

### 3. Snapshot shape

**Choice:** Metadata only (`path`, `size`, `skipped`). Strip `file` / `handle` before IDB put so structured clone cannot fail. Cap listing at a high ceiling if needed (`ponytail:` comment) rather than dropping persist.

**Why:** File objects are not cloneable into IDB; listing is what “the folder is still there” means.

### 4. Tests

**Choice:** Unit-test `toFileSnapshot` / `workspaceFromBinding` (null handle, missing recent, permission prompt uses snapshot + reconnect, close-equivalent empty bind). Browser: open via fallback or mocked picker, reload, Files still shows name + listing.

## Risks / Trade-offs

- [Stale listing after files change on disk] → Snapshot is last open; Reconnect / Open folder refreshes when a handle is available.
- [Pins without readable bodies after refresh] → Listing and pin marks remain; agent file read still needs reconnect/live handle (same as today).
- [IndexedDB blocked] → try/catch; workspace stays in memory for the visit only.

## Migration Plan

1. Ship snapshot fields + stop clearing on null handle; old rows without `files` still load from recents when present.
2. Rollback: ignore extra fields; restore clear-on-null only if rolling back the whole change (not recommended).

## Open Questions

None.
