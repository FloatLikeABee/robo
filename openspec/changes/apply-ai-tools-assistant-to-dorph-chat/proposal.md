## Why

AI Tools assistants still have a **Run** control that opens a one-shot query dialog inside the AI Tools app. Dorph chat already knows how to use those assistants (`agent_id` with a `bk:` prefix), but nothing in the UI applies or dismisses one. Users need a single path: pick an assistant in AI Tools, apply it to Dorph chat, and dismiss it when they want the default chat back.

## What Changes

- Replace the Assistants **Run** button and Run dialog with an **Apply** control.
- Clicking **Apply** applies that assistant to the current Dorph chat so subsequent messages use its prompt, model, and RAG.
- Add a **Dismiss** control to stop using the applied assistant and return Dorph chat to unscoped (no assistant) behavior.
- Only one assistant is applied at a time; applying another replaces the current one.
- Dorph chat shows which assistant is applied and can dismiss it without reopening AI Tools.
- **BREAKING (UI):** the Assistants Run dialog and `Run` label are removed. One-shot `POST /assistants/{id}/run` from this page is no longer offered.

## Capabilities

### New Capabilities

- `dorphai-applied-assistant`: Apply an AI Tools assistant onto Dorph chat from the Assistants list, show applied state, and dismiss it so chat runs without that assistant.

### Modified Capabilities

- (none — no main specs under `openspec/specs/`)

## Impact

- **AI Tools frontend** (`bk/frontend/src/pages/AssistantManager.js`): Run icon/dialog → Apply / Dismiss; signal Dorph when embedded.
- **Dorph frontend** (`dorph/frontend/src/SkoolAiChat.js`, `AiToolsWorkspaceDrawer.jsx`): listen for apply/dismiss, show applied assistant chrome, send `agent_id` on `/api/chat`.
- **Dorph backend**: no new endpoints; existing `POST /api/chat` `agent_id` (`bk:<id>`) and `GET /api/ai-agents` stay as the apply mechanism.
- **AI Tools backend**: keep `POST /assistants/{id}/run` for other callers; this page no longer uses it.
