## Why

Morph AI’s Files workspace (local folder picker, recents, pins, IndexedDB session binding) duplicates Context & Knowledge and never earned its keep. Operators already attach files and the knowledge library on that tab. Drop Files so the agent shell has one file surface.

## What Changes

- **BREAKING (Morph AI agent shell):** Remove the **Files** workspace tab and everything that existed only for it: Open folder, recents, reconnect, pin-from-folder, composer **Files** include chip, and IndexedDB `morphai-files-workspace` session/folder bindings.
- Keep the right workspace pane with **Notes & TODOs** and **Context & Knowledge**. Context & Knowledge stays the place to upload, paste, attach/detach session HybridContext, and manage the Knowledge Library.
- Keep last-chat restore and workspace open/closed preference. Do **not** restore a folder, retitle a session from a folder, or default the pane to Files.
- Embedded / `singleSession` MorphNotes chat is unchanged (it never had this Files workspace).

## Capabilities

### New Capabilities

- `morphai-no-files-workspace`: Morph AI agent workspace has no Files tab or local-folder workspace; Context & Knowledge is the file/knowledge surface.

### Modified Capabilities

- (none — prior Files/workspace specs live only on unarchived changes, not under `openspec/specs/`)

## Impact

- Morph AI frontend: `AgentWorkspace.js`, `AgentFilesTab.js` (remove), `filesWorkspaceStore.js` (remove), `SkoolAiChat.js` (folder/pin/include Files), `agentContext.js` pin helpers if unused, related tests and Files-only CSS
- Browser: stop using IndexedDB `morphai-files-workspace`; leftover DB MAY be ignored (no migration required)
- Chat API: stop sending local-folder `pinned_files` from the UI; backend MAY still accept empty `include_files` / `pinned_files` without a schema drop
- Docs/lessons that tell operators to persist a Files folder must be corrected on apply
- MorphNotes, MorphUtils, Event Logs, Content Maker, Data Access, Project: no change
