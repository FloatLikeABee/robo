## Why

Morph AI operators who work with the right-hand **Files** workspace lose that layout when they come back. Folder bindings already survive in IndexedDB, but the workspace tab is only stored in `sessionStorage` (gone when the browser closes), and there is no persisted preference for whether the right workspace pane is open. Returning users should land on the same Files workspace they left—not an empty default.

## What Changes

- Persist the **workspace pane open/closed** preference in `localStorage` so a user who opened the right workspace sees it open on the next visit.
- Persist the **active workspace tab** (Files / Notes & TODOs / Context & Knowledge) in `localStorage` per chat `sessionId`, replacing the current `sessionStorage`-only write.
- On Morph AI load, **restore** the saved tab and open state before paint where possible; when the current session has a folder binding, ensure **Files** is shown with that folder (or reconnect prompt) without requiring the user to reopen the panel or tab.
- Add a compact **show/hide workspace** control on the agent shell (chat header or workspace edge) so operators can collapse the right pane for more chat width; the open state feeds the same persistence.
- Embedded / `singleSession` Morph AI stays unchanged (no workspace pane).

## Capabilities

### New Capabilities

- `morphai-workspace-open-persist`: Cross-visit persistence for Morph AI right workspace visibility and active tab, with folder workspace restored on return when IndexedDB still has a session binding.

### Modified Capabilities

- (none — prior `morphai-files-session-workspace` and `morphai-agent-workspace` live only under `openspec/changes/`, not `openspec/specs/`)

## Impact

- `morph/frontend/src/SkoolAiChat.js` (workspace open state, restore on mount, toggle control)
- `morph/frontend/src/components/chat/AgentWorkspace.js` (`workspaceTabStorageKey` read/write → `localStorage`)
- `morph/frontend/src/App.css` (collapsed workspace grid column)
- Browser-only. No Go API or SQLite changes.
