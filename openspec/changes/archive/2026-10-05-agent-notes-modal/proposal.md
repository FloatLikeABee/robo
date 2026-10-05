## Why

#132 shipped Agent notes as a full page at `/agent-notes`. The chat header leaves the conversation to open it. The rest of the chat shell keeps the human in the conversation: AI tools and the notes drawer are overlays. Agent notes should open the same way.

## What Changes

- The header and more-menu "Agent notes" item opens a large overlay on the current chat. It does not assign `/agent-notes`.
- The overlay reuses the list, detail, empty, and error reading from the page, including the phone column and a 44px back control.
- `/agent-notes` stops being a second page. A signed-in visit lands in chat with the same overlay open. A logged-out visit still goes to login.
- The session-only list API stays. Ownership rules from #132 do not change. No backend change.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `morph-agent-notes`: The human view is a modal over chat, not a standalone page. Deep links and logged-out handling follow that view. List, detail, and ownership requirements stay.

## Impact

- `morph/frontend`: `SkoolAiChat.js`, `appRouter.js`, the Agent notes page and its test, `App.css`.
- Docs that name `/agent-notes` as the human list (`docs/agents/14-morph-mcp.md`).
- No Go handlers, no MorphUtils, no Render.
