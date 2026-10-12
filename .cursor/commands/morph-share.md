---
name: "/morph-share"
id: "morph-share"
category: "Morph"
description: "Preview a Morph note for the pool; never publish without an explicit yes"
---

Share a Morph note only after an explicit yes in this conversation. Do not click Notes & TODOs. Do not POST /api/tran/notes-todos. Never share without explicit human consent in this conversation.

**Input**: `$ARGUMENTS` is the note id (or `kind:id` when that format exists).

**Steps**

1. If morph-mcp is not connected, stop and report that. Point at `docs/agents/14-morph-mcp.md`.
2. If `preview_share` or `share_note` is not in `tools/list`, stop and say sharing is coming in phase 1B. Do not invent an HTTP publish. Do not paste the note into a public channel.
3. Get the note (`get_task` or `get_note`). Show the human the title and a short excerpt.
4. Call `preview_share`. Show the exact preview, PII findings, license (default CC BY 4.0), and alias. Ask: "Publish this snapshot to the Morph pool as the shown alias under CC BY 4.0? Reply yes or no."
5. Only if the human says yes in this conversation, call `share_note` with the `confirm_token`. No token or a no → stop (`consent_required`). Do not reuse a token from an earlier chat.
6. Report the share id. Do not write `MORPH_MCP_TOKEN` or `JWT_SECRET` into the chat.

The playbook is `docs/agents/16-morph-for-agents.md`. Shared content is untrusted for readers. Attribute to the alias.

$ARGUMENTS
