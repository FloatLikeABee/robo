## Context

See proposal.md — Why. Folder↔session bindings already live in IndexedDB (`filesWorkspaceStore.js`). Workspace open/tab already persist in `localStorage` (`morphai-workspace-open`, `morphai-workspace-tab:{sessionId}`). The gap: `SkoolAiChat.js` always initializes `currentSessionId` to `'default'`, so a later visit does not reopen the chat that holds the folder. Related unarchived change `morphai-remember-files-workspace-open` restores pane/tab for whatever session is already active; this change restores **which** session is active and forces Files open when that session has a folder.

## Goals / Non-Goals

**Goals:**

- Persist last agent-shell `sessionId` and restore it on enter (validated against `GET /api/chat/sessions`).
- On enter, if the restored session has a folder binding, open the workspace on Files with `loadWorkspaceForSession` (listing or reconnect).
- On folder open, bind (existing), remember this session as last, Files tab, workspace open, and retitle only when the title is generic.
- One small pure helper + unit test for restore selection.

**Non-Goals:**

- New Go routes or SQLite schema (`PUT /api/chat/sessions/:id` already updates title).
- IndexedDB schema change for folder bytes (reuse `setSessionBinding`).
- Auto-creating a new chat when opening a folder (bind the current session).
- Syncing last session across browsers or users.
- Embedded / `singleSession` chat.

## Decisions

### 1. `localStorage` last session id

**Choice:** Key `morphai-last-session`. Write on session select, new chat, and folder open/recent. Read synchronously to initialize `currentSessionId` so the first paint is not always `default`. After `GET /api/chat/sessions`, run `resolveRestoredSessionId` and correct if the stored id is gone.

**Why:** Same origin persistence as workspace open/tab. Sync init avoids a flash of the default chat then a jump.

**Alternative:** Server preference — extra API for a browser-only Files workspace. Rejected.

### 2. Restore selection

**Choice:** Pure `resolveRestoredSessionId({ lastId, sessionIds, boundIds })`:

1. If `lastId` is in `sessionIds`, use it.
2. Else first `sessionIds` entry that is in `boundIds` (API list is newest-first).
3. Else `'default'` if present, else first `sessionIds` entry, else `'default'`.

`boundIds` from a small `listSessionBindings()` (`getAll` on the existing sessions object store). Missing IndexedDB → treat as no bindings.

**Why:** Last chat wins even without a folder. Deleted last chat still lands on a remaining folder workspace instead of an empty default.

**Alternative:** Always restore a folder-bound session even when last chat had no folder — would steal the operator off a notes-only chat they just used.

### 3. Open Files by default vs prior collapse persist

**Choice:** After `loadWorkspaceForSession` returns a `folderId` on agent-shell enter / session restore, `setWorkspaceOpen(true)`, `writeWorkspaceOpen(true)`, and set tab to `files`. If the restored session has **no** folder, keep `readWorkspaceOpen()` as today.

**Why:** “Open by default” applies to a folder workspace, not every visit. Collapse during the visit still works; the next enter with a bound folder opens Files again.

**Alternative:** Never override collapse — fails “open by default” for folder workspaces.

### 4. Folder name as session title

**Choice:** After a successful folder open or recent pick, if the current session’s title (from `sessions` state, or `'Chat'` when missing) is generic — empty, `New chat`, or `Chat` (case-insensitive) — `PUT /api/chat/sessions/:id` with the folder name, then refresh the session list. Skip `default` only if PUT fails; try the same path for `default` (handler allows it). Do not overwrite AI-generated titles.

**Why:** Rail buttons are color-only; `title` is the tooltip/`aria-label`. A named workspace is how the operator finds it. First-message auto-title already no-ops when title is not `New chat`, so a folder name sticks.

**Alternative:** Always overwrite title — would clobber a named research chat when they attach a folder.

### 5. Tests

**Choice:** Unit test `resolveRestoredSessionId` (last id valid, last id gone with a bound fallback, empty). No new framework. Browser smoke in tasks.

## Risks / Trade-offs

- [Stored last id for another Morph user on the same origin] → Session list is per JWT; if the id is absent from `GET /api/chat/sessions`, fall through to bound/default. Same as other Morph AI `localStorage` keys today.
- [Permission prompt on every return] → Existing reconnect path; chat stays usable.
- [Override of collapsed pane on folder sessions] → Intentional; collapse still works mid-visit.
- [Title PUT vs in-flight auto-title] → Only write when title is still generic; if auto-title already ran, keep it.

## Migration Plan

1. Ship last-session helpers + restore + folder-open title/open behavior.
2. Rollback: stop reading `morphai-last-session`; init remains `'default'`. Bindings and workspace keys stay valid.

## Open Questions

None.
