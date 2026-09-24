# OpenSpec project notes

## Morph AI Files workspace

Do not restore Morph AI’s IndexedDB Files tab (local folder picker, recents, pins, `AgentFilesTab`, `filesWorkspaceStore`, database `morphai-files-workspace`) unless the product owner explicitly says yes.

Current Files policy is the active change `morphai-drop-files-workspace`: the agent shell has Notes & TODOs and Context & Knowledge only, and a stored `files` tab means Context & Knowledge.

These archived changes are history, not work to re-implement, and their delta specs must not be merged into `openspec/specs/`:

- `2026-09-24-morphai-files-session-workspace`
- `2026-09-24-morphai-restore-folder-workspace-session`
- `2026-09-24-morphai-files-folder-survives-refresh`
- `2026-09-24-morphai-remember-files-workspace-open`
- `2026-09-24-morphai-pinned-files-readable-context`

Active changes that still mention a Files tab are not permission to bring it back: `morphai-agent-workspace`, `morphai-restore-missing-webpack-modules`, and the “Files-workspace tabs MUST still exist” sentence in `platform-trim-readme-chat-skills`.
