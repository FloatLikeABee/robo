## 1. Session snapshot helpers

- [x] 1.1 Add `toFileSnapshot(files)` (path/size/skipped only) and `workspaceFromBinding({ bind, rec, permission })` in `morph/frontend/src/lib/filesWorkspaceStore.js` so a missing handle or unreadable handle still returns folder name + snapshot files + reconnect when a handle exists
- [x] 1.2 Extend `setSessionBinding` to store `folderName` and snapshot `files` on the session row; `loadWorkspaceForSession` uses `workspaceFromBinding` and MUST NOT return empty when the recent row is gone but the session snapshot remains
- [x] 1.3 Add `rememberFolderByName(name)` for handle-less opens; reuse an existing recent with the same name when present
- [x] 1.4 Unit-test snapshot + `workspaceFromBinding`: null handle keeps listing, missing recent keeps session snapshot, permission not granted keeps listing with reconnect

## 2. Open path

- [x] 2.1 `handleFolderOpened` in `SkoolAiChat.js`: always bind with snapshot after a successful open; call `rememberDirectory` when a handle exists, else `rememberFolderByName`; never `clearSessionBinding` on null handle
- [x] 2.2 Close folder remains the only `clearSessionBinding`; confirm `singleSession` still does not persist Files

## 3. Verification

- [x] 3.1 Unit tests in 1.4 pass
- [x] 3.2 Browser: open a folder (webkit fallback or picker), reload — Files shows the same folder name and listing, not “No folder open”
- [x] 3.3 Browser: Close folder, reload — Files is empty again
