## 1. Persistence

- [x] 1.1 Add IndexedDB helper: recents (id, name, directory handle, cap ~10 unique, most-recent first) and per-session map (`sessionId` → folder id + pinned paths)
- [x] 1.2 On open folder: store handle, prepend recents, write current session’s folder id; re-walk tree after `requestPermission` on restore

## 2. Session-bound Files UI

- [x] 2.1 Key Files state in `SkoolAiChat.js` by `sessionId`: restore folder + pins on session switch; Close folder clears only that session
- [x] 2.2 `AgentFilesTab`: recents list to apply a folder to the current session; remove the tutorial `hybrid-hint` paragraph; keep Open folder and a short empty state

## 3. Verify

- [x] 3.1 Open a folder, switch chat sessions and back: same folder and pins; reload restores when permission is granted; recents reopen without a full picker when the handle is valid; the old lede is gone
