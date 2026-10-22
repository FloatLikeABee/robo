---
name: "/dorph-log-session"
id: "dorph-log-session"
category: "Dorph"
description: "Save this session as a Dorph note without grilling"
---

Save a session recap as a Dorph note. Do not click Notes & TODOs. Do not POST /api/tran/notes-todos. Never share without explicit human consent in this conversation.

**Input**: optional `$ARGUMENTS` title. The body is the recap you just produced in this chat.

**Steps**

1. If `create_note` is not available, stop and report that dorph-mcp is not connected. Point at `docs/agents/14-dorph-mcp.md`.
2. Call `suggest_note_kind` with `intent` set to the title or "Save a session summary" and `content_hint` set to your recap. Expect `kind` `note` and `use_grill` false. Skip `note_grill`.
3. Call `list_my_tasks` with `type` `note` (or `search_my_notes` if it exists) and do not create a second copy of the same recap.
4. Call `create_note` with `title` from `$ARGUMENTS` or "Session recap", and `body` equal to the recap. Send a plain title. The server stores `[dorph-mcp]` and `source: dorph-mcp`.
5. Report the id and kind. Do not write `DORPH_MCP_TOKEN` or `JWT_SECRET` into the note or the chat.

The playbook is `docs/agents/16-dorph-for-agents.md`. The human reviews at http://localhost:3031/.

$ARGUMENTS
