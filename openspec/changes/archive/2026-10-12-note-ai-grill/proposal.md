## Why

Notes & TODOs AI assist sends the title to the model in one shot. A title that reads like a request ("make this morphAI a platform for agents, actually agent notebook") comes back as a chatty assistant reply ("Great! I'm Morph AI...") instead of note content. Small local Ollama models follow that cue. The operator never gets to correct the model before the body is filled.

## What Changes

- AI assist on a new or existing note or TODO opens a modal grill. The grill asks on-topic questions about the title and any existing text, then shows a draft and inserts it only after the operator confirms.
- Question count follows topic complexity and stops at a cap. The operator can skip a question or finish early ("enough, write it now").
- The same conversation is an MCP Apps page (`ui://morph/grill`) with a session-scoped `note_grill` tool, and that page is what the Morph SPA modal renders.
- The draft is note or TODO body text. A greeting, a self-introduction, or an offer of more help is not inserted.
- Big notes, task-chain steps, and Render/deploy config stay as they are.

## Capabilities

### New Capabilities

- `note-ai-grill`: Notes and TODOs AI assist grills from the title, then inserts confirmed body text. The same flow is the MCP grill app.

### Modified Capabilities

- `morph-mcp-stdio`: Adds the grill app resource and the `note_grill` tool, scoped to the verified session. The notes app resource stays.
- `morph-chat-reply-shape`: Chat still answers in plain language. Note and TODO AI assist no longer treats the title as a question to answer; the grill owns that path.

## Impact

- `morph/grill` state machine and prompts, used by the API and morph-mcp.
- `POST /api/tran/grill` and `GET /api/tran/grill/ui` on the Morph API (session required, same private `/api/tran/` rule).
- Morph SPA Notes & TODOs modal (`morph/frontend`).
- morph-mcp tools, `ui://morph/grill`, and `docs/agents/14-morph-mcp.md`.
- `POST /api/tran/text-assist` remains for task-chain steps. The notes/TODO button stops calling it.
- No Render, image, or deploy changes.
