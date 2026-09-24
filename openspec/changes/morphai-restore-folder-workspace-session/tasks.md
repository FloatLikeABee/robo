## 1. Last-session helpers

- [x] 1.1 Add `morphai-last-session` read/write helpers (try/catch) and pure `resolveRestoredSessionId({ lastId, sessionIds, boundIds })` next to the existing workspace storage helpers in `morph/frontend/src/components/chat/AgentWorkspace.js`
- [x] 1.2 Add `listSessionBindings()` in `morph/frontend/src/lib/filesWorkspaceStore.js` (`getAll` on the sessions store; empty array on failure)
- [x] 1.3 Unit-test `resolveRestoredSessionId`: valid last id, missing last id with a bound fallback, empty inputs → `default`

## 2. Restore on enter

- [x] 2.1 On agent shell, initialize `currentSessionId` from last-session storage (fallback `default`); skip for `singleSession` / embedded
- [x] 2.2 After `GET /api/chat/sessions`, resolve against session ids + `listSessionBindings()` and `setCurrentSessionId` if different; write the chosen id back to last-session storage
- [x] 2.3 Write last-session id on select session, new chat, and successful folder open / recent pick (agent shell only)

## 3. Folder workspace session

- [x] 3.1 After `loadWorkspaceForSession` returns a `folderId` on the agent shell, open the workspace pane, persist open, and select the Files tab
- [x] 3.2 After folder open or recent pick, if the current session title is generic (`New chat`, `Chat`, empty), `PUT /api/chat/sessions/:id` with the folder name and refresh the session list; do not overwrite a custom title
- [x] 3.3 Confirm embedded / `singleSession` does not read/write last session, force-open Files, or retitle from a folder

## 4. Verification

- [x] 4.1 Unit: `resolveRestoredSessionId` tests pass
- [x] 4.2 Browser: open a folder in a non-default chat, reload Morph AI — that chat is active, Files is open, folder listing (or reconnect) is shown
- [x] 4.3 Browser: new chat titled New chat, open a folder — session title becomes the folder name; a chat with a custom title keeps it
- [x] 4.4 Browser: session with no folder — reload does not force the workspace pane open
