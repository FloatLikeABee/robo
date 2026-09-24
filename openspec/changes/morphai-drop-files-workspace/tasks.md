## 1. Workspace tabs without Files

- [x] 1.1 `AgentWorkspace.js`: tabs are Notes & TODOs and Context & Knowledge only. Remove `AgentFilesTab` import and Files pane. Default / unknown / stored `files` tab → `knowledge`. Drop Files-only props (`folderName`, `files`, pins, folder callbacks)
- [x] 1.2 Delete `morph/frontend/src/components/chat/AgentFilesTab.js`. Remove Files-only CSS if nothing else uses it
- [x] 1.3 `resolveRestoredSessionId`: last existing session, else default. Drop `boundIds` / folder-binding fallback. Delete `isGenericSessionTitle` if unused after folder retitle is gone

## 2. Chat shell: no folder workspace

- [x] 2.1 `SkoolAiChat.js`: remove Files include chip, folder/pin/reconnect/recent handlers, `filesWorkspaceStore` usage, and `pinned_files` / local-pin send path. Keep Notes and Knowledge chips. Stop retitling sessions from a folder. Keep last-chat restore and workspace open/closed
- [x] 2.2 Delete `morph/frontend/src/lib/filesWorkspaceStore.js` and `filesWorkspaceStore.test.js`. Strip `agentContext.js` pin/folder helpers that exist only for Files (and their tests). Leave HybridContext / knowledge include fingerprinting working
- [x] 2.3 Update `.cursor/skills/session-lessons/lessons.md`: Morph AI has no Files workspace; Context & Knowledge is the file surface. Remove or invert the “Files folder must survive refresh” lesson

## 3. Verify

- [x] 3.1 Jest: `AgentWorkspace.lastSession.test.js` (no Files mock; last-id restore without bound sessions), remaining `agentContext` tests, no imports of deleted Files modules
- [x] 3.2 Browser (agent shell, not embedded): workspace tabs are Notes & TODOs and Context & Knowledge only; no Open folder / recents; include bar has Notes and Knowledge, not Files; Context & Knowledge still uploads/attaches; reload does not show a Files pane; last chat still restores. Dark-only
