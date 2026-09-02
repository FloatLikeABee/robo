## Why

Morph AI Files treats the opened folder as throwaway UI state: refresh loses it, switching chat sessions does not restore that session’s folder and pins, and a long hint explains a workflow the user already understands. The conversation and the files should stay one workspace.

## What Changes

- Bind the Files folder (and pins) to the **current chat session**. Switching sessions restores that session’s workspace folder and pins; opening a folder in a session keeps it as that session’s workspace until the user closes it or opens another folder there.
- Persist the workspace across browser reloads when the browser allows (directory handle + permission), so the chosen folder remains the workspace without picking it again.
- Keep a **recent folders** list so the user can reopen a previously used folder without hunting through the system picker.
- **Remove** the Files hint: “Open a local folder to browse files. Pin files to include them in the agent. Morph never writes to this folder.” Empty state stays short (actions + recents, not a tutorial).
- Morph still MUST NOT write to the local folder. Embedded MorphNotes chat (`singleSession`) stays without this Files workspace.

## Capabilities

### New Capabilities

- `morphai-files-session-workspace`: Per-session Files folder that persists as the workspace, with a recent-folder list for quick reopen, and no instructional Files lede.

### Modified Capabilities

- (none — `morphai-agent-workspace` lives only on a prior change, not under `openspec/specs/`)

## Impact

- Morph AI frontend: `AgentFilesTab.jsx`, `AgentWorkspace.jsx`, `SkoolAiChat.js` (folder/pin state is currently global, not keyed by `sessionId`)
- Browser storage for directory handles and recents (IndexedDB / File System Access); no Morph API change required for folder bytes
