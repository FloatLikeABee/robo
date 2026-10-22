---
name: "/dorph-notes"
id: "dorph-notes"
category: "Dorph"
description: "List Dorph notes with dorph-mcp list_my_tasks"
---

List the signed-in user's Dorph notes by calling the dorph-mcp tool `list_my_tasks`. Do not click Notes & TODOs. Do not POST /api/tran/notes-todos.

**Input**: optional `$ARGUMENTS` of `type=`, `status=`, and `limit=` tokens.

**Steps**

1. If `list_my_tasks` is not available, stop and report that dorph-mcp is not connected. Point at `docs/agents/14-dorph-mcp.md`. Do not invent a token.
2. Default the call to `type` `note`. Apply `type`, `status`, and `limit` from `$ARGUMENTS` when those tokens are present. `type` is `all`, `note`, or `todo`. `status` is `all`, `open`, or `done`. Omit `limit` for the default 50. The cap is 100. Do not pass `0` to mean none.
3. Call `list_my_tasks`. Summarize `id`, `title`, and `status` for each returned task. A note missing from that page is not proof the id does not exist. Call `get_task` when you have an id.
4. Do not write `DORPH_MCP_TOKEN` or `JWT_SECRET` into the chat.

The human review origin is http://localhost:3031/. The rest of the contract is `.cursor/skills/dorph-notes/SKILL.md`.

$ARGUMENTS
