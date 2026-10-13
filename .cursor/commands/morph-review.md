---
name: "/morph-review"
id: "morph-review"
category: "Morph"
description: "Review the signed-in user's open Morph notes and todos"
---

Review the signed-in user's Morph notes. Do not click Notes & TODOs. Do not POST /api/tran/notes-todos.

**Input**: optional `$ARGUMENTS` such as `type=todo` `status=open`.

**Steps**

1. If `list_my_tasks` is not available, stop and report that morph-mcp is not connected. Point at `docs/agents/14-morph-mcp.md`.
2. Default to `type` `todo` and `status` `open`. Apply `type=`, `status=`, and `limit=` from `$ARGUMENTS` when present. Cap is 100. Default limit is 50.
3. Call `list_my_tasks`. If `search_my_notes` exists and `$ARGUMENTS` is a free-text query, call that instead.
4. Summarize `id`, `title`, and `status`. A short page is not the full set. Call `get_task` (or `get_note`) for any id you discuss.
5. Do not create, share, or archive unless the human asks. Never share without explicit human consent in this conversation. Do not write `MORPH_MCP_TOKEN` or `JWT_SECRET` into the chat.

The playbook is `docs/agents/16-morph-for-agents.md`. The human review origin is http://localhost:3031/.

$ARGUMENTS
