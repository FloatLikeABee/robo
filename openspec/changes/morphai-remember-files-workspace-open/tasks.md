## 1. Storage helpers

- [x] 1.1 Add `workspaceOpenStorageKey()` and read/write helpers for `morphai-workspace-open` in `AgentWorkspace.js` (or a tiny adjacent util) with try/catch fallbacks
- [x] 1.2 Switch `workspaceTabStorageKey` read/write from `sessionStorage` to `localStorage` in `AgentWorkspace.js` and `SkoolAiChat.js`

## 2. Restore on load

- [x] 2.1 Initialize `workspaceOpen` state from `localStorage` (default open when unset) in `SkoolAiChat.js` for `isAgentShell`
- [x] 2.2 On `sessionId` change, restore tab from `localStorage`; when `loadWorkspaceForSession` returns a `folderId` and no saved tab exists, set tab to `files`
- [x] 2.3 Persist `workspaceOpen` and tab changes to `localStorage` on toggle/tab switch

## 3. Workspace show/hide UI

- [x] 3.1 Add header toggle (agent shell only) with `aria-expanded` wired to `workspaceOpen`
- [x] 3.2 Apply `app--workspace-collapsed` when closed; update `App.css` grid so chat expands and workspace hides on desktop and stacked breakpoints

## 4. Verification

- [x] 4.1 Smoke: open Files + folder, reload — Files tab and folder visible with workspace open
- [x] 4.2 Smoke: collapse workspace, reload — workspace stays hidden; reopen toggle restores pane and last tab
- [x] 4.3 Smoke: embedded / `singleSession` chat unchanged (no toggle, no collapse class)
