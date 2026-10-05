---
name: "/morph-note-get"
id: "morph-note-get"
category: "Morph"
description: "Get one Morph note with morph-mcp get_task"
---

Read one Morph note by calling the morph-mcp tool `get_task`. Do not click Notes & TODOs. Do not POST /api/tran/notes-todos.

**Input**: `$ARGUMENTS` is the integer `id`.

**Steps**

1. If `get_task` is not available, stop and report that morph-mcp is not connected. Point at `docs/agents/14-morph-mcp.md`. Do not invent a token.
2. If `$ARGUMENTS` is missing or is not an integer, ask for the id and do not call the tool. Do not guess an id.
3. Call `get_task` with that integer `id`. There is no user-id argument.
4. Report `id`, `item_type`, `title`, `status`, and `body`. An id that is missing or belongs to someone else is not found. `id` of 0 or less is invalid. Do not retry as another user, and do not POST /api/tran/notes-todos.
5. Do not write `MORPH_MCP_TOKEN` or `JWT_SECRET` into the chat.

The human review origin is http://localhost:3031/. The rest of the contract is `.cursor/skills/morph-notes/SKILL.md`.

$ARGUMENTS
