---
name: morph-notes
description: Create, list, and get the signed-in user's Morph Notes and TODOs with the morph-mcp tools create_note, list_my_tasks, and get_task. Use when an agent should post a note, fetch notes, or leave text for a human to review on local Morph.
---

# Morph notes

Drive Morph notes yourself. The human reviews them at http://localhost:3031/ Notes & TODOs. You call morph-mcp. The list tool is `list_my_tasks`, not `list_notes`. The get tool is `get_task`, not `get_note`. Do not click Notes & TODOs to create the note. Do not POST /api/tran/notes-todos.

If `create_note`, `list_my_tasks`, or `get_task` is not available, stop and report that morph-mcp is not connected. The local `mcp.json` shape is in `docs/agents/14-morph-mcp.md`. Do not invent `MORPH_MCP_TOKEN`, do not read `.env`, do not run a login curl, and do not write `MORPH_MCP_TOKEN` or `JWT_SECRET` into a note, a commit, or the chat. Do not change Render or a deploy config.

`whoami` is optional. Call it when it is available and you do not know which user the server is. Do not refuse a create solely because you skipped `whoami`. These tools take no user-id argument. They act as the verified session user.

## When

- Create when the task is to leave a note, a handoff, or other text the human should see in Notes & TODOs. Do not create a note for every thought. `create_note` stores `item_type` `note` only. It does not create a todo, a MorphNotes Tasks board row, a big note, or a tool note.
- List before you claim which notes exist. Call `list_my_tasks`.
- Get when you have an id, from create or from the list. Call `get_task`. A note missing from that page is not proof the id does not exist.

Cursor commands `/morph-note`, `/morph-notes`, and `/morph-note-get` are prompts for the same three calls.

## create_note

Title or body is required. Send `title` and `body`.

- `title`: optional, at most 200 Unicode code points. Send a plain title. The stored title starts with `[morph-mcp]`. A title that already starts with `[morph-mcp]` is not prefixed twice.
- `body`: optional, at most 32000 Unicode code points, before the server adds its line. The stored body starts with `source: morph-mcp`.
- Put the note text in `body`. A later title edit in the human editor keeps the body.
- Both empty after trim: the tool errors with `title or body is required` and stores nothing.
- Over the limit: `title is too long` or `body is too long`, and nothing is stored.
- The same stored title and body for this user returns the existing id. That is the same note. Do not POST /api/tran/notes-todos to force a second copy.
- `not a notes user` (the tool says `not a notes user for this session`): stop. Do not retry, do not pass user id 1, and do not POST /api/tran/notes-todos.
- `could not store the note` or `note store is not open`: stop and report that. Do not switch to HTTP.

The result is one task: `id`, `item_type`, `title`, `status` (`open` or `done`), `body`, and optional `deadline_at`, `created_on`, `last_updated`.

## list_my_tasks

List accepts `type`, `status`, and `limit`.

- `type`: `all`, `note`, or `todo`. Omit it for all. Use `note` when you want notes.
- `status`: `all`, `open`, or `done`. Omit it for all.
- `limit`: omit it for the default 50. The cap is 100. `0` selects that default, not an empty page. A negative limit errors with `invalid limit`.
- Other type or status values error with `invalid type` or `invalid status`.
- The result is `{tasks, limit}`. Each task has the fields above. For `type=note`, newest `created_on` is first. A short page is not the full set.

## get_task

Get accepts an integer `id`. `id` of 0 or less is `invalid id`. An id that is missing or belongs to someone else is not found. The tool does not say the row is forbidden. Do not retry with a different user.

## Secrets

Do not write `MORPH_MCP_TOKEN` or `JWT_SECRET` into the note.
