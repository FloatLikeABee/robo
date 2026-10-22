> **Superseded (Files modules only)** by `dorphai-drop-files-workspace`. Do not restore `AgentFilesTab` or `filesWorkspaceStore`. The other modules this change restores (`AgentWorkspace`, `agentContext`, `AiToolsWorkspaceDrawer`, `appliedAssistantChannel`, `ExtractJsonFromTextDialog`) are still required.

## Why

Dorph AI webpack fails with `Module not found` for chat workspace helpers (`AiToolsWorkspaceDrawer`, `agentContext`, `filesWorkspaceStore`, `AgentWorkspace`, `appliedAssistantChannel`) and DorphNotes `ExtractJsonFromTextDialog`. Those files were left untracked, then dropped during the `main` rebase/stash, while `SkoolAiChat.js` and `AdminDataGrid.js` still import them.

## What Changes

- Put every module those two files import on disk in a CRA-resolvable path (prefer `.js` next to other Dorph frontend sources).
- Keep `AgentFilesTab` with `AgentWorkspace` so the Files tab still loads.
- Confirm there are no leftover git conflict markers in the Dorph frontend sources webpack compiles.

## Capabilities

### New Capabilities

- `dorphai-webpack-workspace-modules`: Dorph AI and DorphNotes admin grids compile because session-workspace, applied-assistant, and extract-JSON modules resolve from the existing imports.

### Modified Capabilities

- None. Workspace/apply-assistant behavior already lives in unarchived changes (`dorphai-agent-workspace`, `dorphai-files-session-workspace`, `apply-ai-tools-assistant-to-dorph-chat`). This change only restores the files webpack needs.

## Impact

- `dorph/frontend/src/SkoolAiChat.js`
- `dorph/frontend/src/components/admin/AdminDataGrid.js`
- `dorph/frontend/src/components/chat/*` (`AiToolsWorkspaceDrawer`, `AgentWorkspace`, `AgentFilesTab`)
- `dorph/frontend/src/lib/{agentContext,filesWorkspaceStore,appliedAssistantChannel}.js`
- `dorph/frontend/src/components/admin/ExtractJsonFromTextDialog.js`
- Dorph AI CRA/webpack overlay at port 3031
