---
name: "/dorph-todo"
id: "dorph-todo"
category: "Dorph"
description: "Capture a Dorph todo: suggest kind, then create or say the todo tool is missing"
---

Leave one Dorph todo for the signed-in human. Do not click Notes & TODOs. Do not POST /api/tran/notes-todos. Never share without explicit human consent in this conversation.

**Input**: `$ARGUMENTS` is the todo text.

**Steps**

1. If dorph-mcp tools are missing, stop and report that dorph-mcp is not connected. Point at `docs/agents/14-dorph-mcp.md`.
2. If `$ARGUMENTS` is empty, ask for the todo text and do not call a tool.
3. Call `suggest_note_kind` with `intent` set to `$ARGUMENTS`. If `kind` is not `todo`, say so and ask whether to continue as that kind.
4. If `use_grill` is true and `note_grill` is available, start a grill with `kind` `todo` and show the draft. Do not save until the human confirms.
5. Today's `create_note` stores a note, not a todo. If a todo-kind create tool is not in `tools/list`, write a note whose title is the first line and whose body starts with `TODO:` plus the argument text, then tell the human it is a note standing in for a todo until unified `kind` create lands. If `create_note` with `kind` `todo` exists, call that instead.
6. Report the id and kind. Do not write `DORPH_MCP_TOKEN` or `JWT_SECRET` into the note or the chat.

The playbook is `docs/agents/16-dorph-for-agents.md`. The human reviews at http://localhost:3031/.

$ARGUMENTS
