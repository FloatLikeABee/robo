---
name: "/dorph-note"
id: "dorph-note"
category: "Dorph"
description: "Create a Dorph note with dorph-mcp create_note"
---

Create one Dorph note by calling the dorph-mcp tool `create_note`. Do not click Notes & TODOs. Do not POST /api/tran/notes-todos.

**Input**: `$ARGUMENTS` is the note text.

**Steps**

1. If `create_note` is not available, stop and report that dorph-mcp is not connected. Point at `docs/agents/14-dorph-mcp.md`. Do not invent a token, and do not write `DORPH_MCP_TOKEN` or `JWT_SECRET` into the note or the chat.
2. If `$ARGUMENTS` is empty, ask for the note text and do not call the tool.
3. `title` is the first non-empty line, trimmed to 200 Unicode code points. `body` is the full argument text. If the body is over 32000 Unicode code points, do not call.
4. Call `create_note` with that title and body. Send a plain title. The server stores a title that starts with `[dorph-mcp]` and a body that starts with `source: dorph-mcp`.
5. Report the returned `id` and stored `title`. The same stored title and body returns the existing id.
6. If the error says the session is not a notes user, stop. Do not retry with `POST /api/tran/notes-todos`.

The human reviews the note at http://localhost:3031/. The rest of the contract is `.cursor/skills/dorph-notes/SKILL.md`.

$ARGUMENTS
