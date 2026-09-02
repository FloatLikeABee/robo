## Why

Morph AI webpack fails with `Module not found` for chat workspace helpers (`AiToolsWorkspaceDrawer`, `agentContext`, `filesWorkspaceStore`, `AgentWorkspace`, `appliedAssistantChannel`) and MorphNotes `ExtractJsonFromTextDialog`. Those files were left untracked, then dropped during the `main` rebase/stash, while `SkoolAiChat.js` and `AdminDataGrid.js` still import them.

## What Changes

- Put every module those two files import on disk in a CRA-resolvable path (prefer `.js` next to other Morph frontend sources).
- Keep `AgentFilesTab` with `AgentWorkspace` so the Files tab still loads.
- Confirm there are no leftover git conflict markers in the Morph frontend sources webpack compiles.

## Capabilities

### New Capabilities

- `morphai-webpack-workspace-modules`: Morph AI and MorphNotes admin grids compile because session-workspace, applied-assistant, and extract-JSON modules resolve from the existing imports.

### Modified Capabilities

- None. Workspace/apply-assistant behavior already lives in unarchived changes (`morphai-agent-workspace`, `morphai-files-session-workspace`, `apply-ai-tools-assistant-to-morph-chat`). This change only restores the files webpack needs.

## Impact

- `morph/frontend/src/SkoolAiChat.js`
- `morph/frontend/src/components/admin/AdminDataGrid.js`
- `morph/frontend/src/components/chat/*` (`AiToolsWorkspaceDrawer`, `AgentWorkspace`, `AgentFilesTab`)
- `morph/frontend/src/lib/{agentContext,filesWorkspaceStore,appliedAssistantChannel}.js`
- `morph/frontend/src/components/admin/ExtractJsonFromTextDialog.js`
- Morph AI CRA/webpack overlay at port 3031
