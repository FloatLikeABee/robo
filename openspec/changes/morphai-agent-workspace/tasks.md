## 1. Agent shell layout

- [x] 1.1 Convert main Morph AI (`!singleSession`) to a three-region grid: sessions | chat | workspace; keep embedded Morph Data chat unchanged
- [x] 1.2 Make the session list collapse in document flow (slim rail or 0 width), persist in `sessionStorage`, and keep new/switch/rename/delete when expanded
- [x] 1.3 Put messages + composer only in the left column (~50% width); stack workspace under chat below ~900px
- [x] 1.4 Raise composer default height to ~3× (`min-height` ~72px, ~3 rows, larger `max-height`); do not stretch the input under the workspace

## 2. Workspace tabs

- [x] 2.1 Add Files / Notes & TODOs / Context & Knowledge tabs in the right pane; persist last tab per session
- [x] 2.2 Mount `NotesTodosContent` in the Notes & TODOs tab (same MorphData store as today’s drawer)
- [x] 2.3 Mount HybridContext + Knowledge Library UI in the Context & Knowledge tab
- [x] 2.4 Retarget header note/knowledge controls to open the matching tab; stop using overlay drawers as the primary Morph AI surface

## 3. Files folder workspace

- [x] 3.1 Open-folder via File System Access API with `webkitdirectory` fallback; empty state if cancelled/unsupported
- [x] 3.2 Render a file tree (skip `.git` / `node_modules` and oversize files); pin/unpin for agent context; never write to the user’s folder

## 4. Context controls and cache

- [x] 4.1 Add include chips on the chat column for pinned files, notes, and knowledge/hybrid; excluded sources must not go on the next send
- [x] 4.2 Add per-user/session context cache (content hashes + include-set fingerprint); reuse on unchanged pins; invalidate on pin/knowledge/notes change
- [x] 4.3 Wire `POST /api/chat` to honor include flags and cache keys; add tests for cache hit vs pin-change miss

## 5. Multi sub-agent

- [x] 5.1 Add a task router that may spawn up to a small cap of workers (files, MorphData, knowledge/notes) and skip fan-out for simple questions
- [x] 5.2 Show sub-agent progress in the existing AI progress UI; merge worker results into one assistant message
- [x] 5.3 Guard that sub-agents never write the local workspace folder; only Morph stores (chat, notes, hybrid, knowledge)

## 6. Verify

- [x] 6.1 Exercise main Morph AI in the browser: collapse sessions, split panes, tall input, all three tabs, pin + include chips
- [x] 6.2 Confirm embedded Morph Data chat is still compact (no session rail, no workspace split)
