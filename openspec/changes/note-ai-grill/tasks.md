## 1. Grill state and prompts

- [x] 1.1 Add `morph/grill` with question cap, plan and write prompts, question parsing, stop phrases, and chatty-draft replacement. Cap uses the title only: 1 for a short single intent and at most 4. A long body does not raise it. A stop phrase is the whole answer. Prompts do not include reply-shape "answer the question" instructions.
- [x] 1.2 Add an in-memory owner-keyed session store (start, answer, skip, finish) that returns not-found for another owner and does not write SQLite. Hold at most 32 sessions per owner.
- [x] 1.3 Test the cap, the parser, the stop phrase, the chatty replacement, and the owner check with a fake completer. No network.

## 2. HTTP API

- [x] 2.1 Add `POST /api/tran/grill` for start, answer, skip, and finish. Require `auth_user_id`. Missing session is 401 and does not call the model. Another user's id is 404 without the draft.
- [x] 2.2 Add `GET /api/tran/grill/ui` that returns the same grill HTML the MCP resource serves. Session required.
- [x] 2.3 Test the HTTP contract with a fake completer, including the chatty draft and the cross-session miss.

## 3. MCP app

- [x] 3.1 Add `ui://morph/grill` and `note_grill` on morph-mcp. Point `note_grill` at that resource. Leave the notes tools on `ui://morph/notes`. Do not write a note from `note_grill`.
- [x] 3.2 Serve the closed dark Arial HTML page. It calls `note_grill` through the host bridge, has no external URL and no HTML sink, and calls `create_note` only after the draft is shown.
- [x] 3.3 Extend the MCP tests for the tool, the resource, the closed HTML, and a fake-completer interview that does not insert a row.

## 4. Morph SPA modal

- [x] 4.1 Open the grill HTML in the Notes & TODOs AI assist modal for notes and TODOs. Bridge `note_grill` to `POST /api/tran/grill`. Pass kind, title, and current body as context.
- [x] 4.2 On confirm, replace the editor body with the draft from the last grill result and do not save. On cancel, or when the model call fails, leave the body unchanged. Do not call `text-assist` from this button.
- [x] 4.3 Add a frontend test that confirm inserts the draft and cancel does not.

## 5. Docs

- [x] 5.1 Document `note_grill` and `ui://morph/grill` in `docs/agents/14-morph-mcp.md`, including session scope and that the tool does not insert a note.
