## 1. Pin content cache (store)

- [x] 1.1 Add `pinnedContents` to session binding in `filesWorkspaceStore.js` (path → `{ hash, content }`, capped)
- [x] 1.2 Add `readFileEntryText(file)` helper shared by pin and send paths
- [x] 1.3 On pin: read text, write cache; on unpin: remove entry; after reconnect: refresh cache for active pins

## 2. Send path

- [x] 2.1 Update `readPinnedFileBodies` to prefer `pinnedContents` cache, then live handle/`File`
- [x] 2.2 In `SkoolAiChat.js` `handleSend`: if `includeFiles` and pins exist but all bodies empty, block send with reconnect message
- [x] 2.3 Clear context fingerprint when pin cache changes

## 3. UI feedback

- [x] 3.1 `AgentFilesTab`: show warning on pinned rows without readable cache when `reconnectNeeded` or snapshot-only
- [x] 3.2 Files `(N)` chip warning style when any pin unreadable
- [x] 3.3 Optional: clicking unreadable pin offers Reconnect

## 4. Tests & verification

- [x] 4.1 Unit test: snapshot-only `files` + `pinnedContents` → `readPinnedFileBodies` returns text
- [x] 4.2 Unit test: pins on, empty bodies → send guard returns error (extract guard fn if needed)
- [x] 4.3 Manual: pin `maindocing/southpole sci-fi story.txt`, ask for summary — assistant uses file text; reload + ask again still works after pin cache
