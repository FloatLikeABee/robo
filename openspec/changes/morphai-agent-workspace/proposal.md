> **Superseded (Files tab only)** by `morphai-drop-files-workspace`. Instructions in this change to build or keep a Morph AI Files tab, local-folder picker, or pin-from-folder are not permission to restore that tab. Notes & TODOs, Context & Knowledge, the agent shell, and orchestration in this change stay.

## Why

Morph AI is still a chat box with overlay drawers. Turning it into an agent needs a durable workspace next to the conversation (files, notes, knowledge) plus room to type, and a way to control context, cache it, and fan work out to sub-agents. The session list currently eats width or sits behind an overlay; collapsing it and splitting chat from workspace makes that possible.

## What Changes

- **Collapsible sessions:** The left chat-session menu on Morph AI (not embedded Morph Data) collapses in place so the chat + workspace get the width. Expanded list still supports new / rename / delete / switch session.
- **Split agent shell:** The main panel splits roughly in half. Left is the conversation (messages + input). Right is the workspace. The composer stays in the left column, to the left of the workspace — not full-width under both panes.
- **Taller composer:** Morph AI chat input is about **three times** the current height so users can paste more context.
- **Workspace tabs:** Three tabs in the right pane:
  1. **Files** — open a folder as the session workspace and browse its files.
  2. **Notes & TODOs** — same MorphData-synced notes/todos that today live in a right drawer.
  3. **Context & Knowledge** — session HybridContext plus Knowledge Library (what today lives in the hybrid-context drawer).
- **Agent controls:** Per-session context toggles (which workspace files, notes, and knowledge are in the prompt), a context **cache** so unchanged workspace material is not re-uploaded every turn, and **multi sub-agent** runs that spawn from the user task (research vs MorphData vs files, etc.) with visible progress, then merge into one reply.
- Embedded Morph Data / `singleSession` chat stays a compact chat (no session rail, no split workspace).
- Notes and HybridContext **overlay drawers** on the main Morph AI page are replaced by workspace tabs (header icon shortcuts may still open the matching tab). **BREAKING (UI):** those overlays are no longer the primary surface on Morph AI.

## Capabilities

### New Capabilities
- `morphai-agent-shell`: Collapsible session rail, 50/50 chat | workspace split, composer in the chat column at ~3× height.
- `morphai-agent-workspace`: Files (open folder + tree), Notes & TODOs, and Context & Knowledge as three tabs in the workspace pane.
- `morphai-agent-orchestration`: Context include/exclude controls, session context cache, and task-dependent multi sub-agent execution with merged replies.

### Modified Capabilities
- (none — `openspec/specs/` has no archived baselines)

## Impact

- **Morph AI frontend:** `SkoolAiChat.js`, `App.css`; new workspace pane/tab components; reuse `NotesTodosContent`, HybridContext/knowledge UI from `HybridContextDrawer.js` and `ChatNotesTodosDrawer.jsx`.
- **Morph AI backend:** `POST /api/chat` gains agent context/cache/sub-agent fields; new or extended endpoints for workspace folder metadata, pinned context, cache invalidation, and sub-agent run status. Existing HybridContext (`/api/chat/hybrid-context*`) and Knowledge (`/api/knowledge/*`) stay the durable stores.
- **Not in scope:** Morph Utils left nav; other apps’ assistants; a local coding agent that writes to the user’s disk; changing production env files.
