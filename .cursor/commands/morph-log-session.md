---
name: "/morph-log-session"
id: "morph-log-session"
category: "Morph"
description: "Save this session as a Morph note without grilling"
---

Save a session recap as a Morph note. Do not click Notes & TODOs. Do not POST /api/tran/notes-todos. Never share without explicit human consent in this conversation.

**Input**: optional `$ARGUMENTS` title. The body is the recap you just produced in this chat.

**Steps**

1. If `create_note` is not available, stop and report that morph-mcp is not connected. Point at `docs/agents/14-morph-mcp.md`.
2. Call `suggest_note_kind` with `intent` set to the title or "Save a session summary" and `content_hint` set to your recap. Expect `kind` `note` and `use_grill` false. Skip `note_grill`.
3. Call `list_my_tasks` with `type` `note` (or `search_my_notes` if it exists) and do not create a second copy of the same recap.
4. Call `create_note` with `title` from `$ARGUMENTS` or "Session recap", and `body` equal to the recap. Send a plain title. The server stores `[morph-mcp]` and `source: morph-mcp`.
5. Report the id and kind. Do not write `MORPH_MCP_TOKEN` or `JWT_SECRET` into the note or the chat.

The playbook is `docs/agents/16-morph-for-agents.md`. The human reviews at http://localhost:3031/.

$ARGUMENTS
