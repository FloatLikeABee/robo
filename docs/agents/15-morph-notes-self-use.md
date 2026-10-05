# 15 — Morph notes, self-use

You are a bot told to use Morph notes. Follow this file. A human does not click Notes & TODOs for you.

Local Morph for the human is http://localhost:3031/. Notes & TODOs on that origin is where they review. `morph-mcp` is a stdio server, not an HTTP server on port 3031 or 9090. Do not change Render or a deploy config.

## Tools

Call the morph-mcp tools. Their names are `create_note`, `list_my_tasks`, and `get_task`. There is no `list_notes` tool and no `get_note` tool. `list_my_tasks` with `type` `note` is how you list notes. `get_task` is how you get one note.

If those tools are not available, stop and report that morph-mcp is not connected. Tell the human the local `mcp.json` shape is in [14-morph-mcp.md](14-morph-mcp.md). Do not invent `MORPH_MCP_TOKEN`, do not read `.env`, do not run a login curl, and do not write `MORPH_MCP_TOKEN` or `JWT_SECRET` into a note, a commit, or the chat.

`whoami` is optional. Call it when it is available and you do not know which user the server is. Do not refuse a create solely because you skipped `whoami`. The note tools take no user-id argument. They act as the verified session user.

Cursor slash commands `/morph-note`, `/morph-notes`, and `/morph-note-get` are prompts for the same three calls. The procedure below is the same whether or not you have those commands.

## When

- Create when the task is to leave a note, a handoff, or other text the human should see in Notes & TODOs. Do not create a note for every thought. `create_note` stores `item_type` `note` only. It does not create a todo, a MorphNotes Tasks board row, a big note, or a tool note.
- List before you claim which notes exist.
- Get when you have an id. A note missing from that page is not proof the id does not exist.

## Create

Call `create_note`. Title or body is required. Send `title` and `body`.

- `title`: optional, at most 200 Unicode code points. Send a plain title. The stored title starts with `[morph-mcp]`. A title that already starts with `[morph-mcp]` is not prefixed twice.
- `body`: optional, at most 32000 Unicode code points, before the server adds its line. The stored body starts with `source: morph-mcp`.
- Put the note text in `body`. A later title edit in the human editor keeps the body.
- Both empty after trim: the tool errors with `title or body is required` and stores nothing.
- Over the limit: `title is too long` or `body is too long`, and nothing is stored.
- The same stored title and body for this user returns the existing id. That is the same note. Do not POST /api/tran/notes-todos to force a second copy.
- `not a notes user` (the tool says `not a notes user for this session`): stop. Do not retry, do not pass user id 1, and do not POST /api/tran/notes-todos. That HTTP route can store a note as the wrong user. The MCP insert does not.
- `could not store the note` or `note store is not open`: stop and report that. Do not switch to HTTP.

The result is one task: `id`, `item_type`, `title`, `status` (`open` or `done`), `body`, and optional `deadline_at`, `created_on`, `last_updated`.

## List

Call `list_my_tasks`. List accepts `type`, `status`, and `limit`.

- `type`: `all`, `note`, or `todo`. Omit it for all. Use `note` when you want notes.
- `status`: `all`, `open`, or `done`. Omit it for all.
- `limit`: omit it for the default 50. The cap is 100. `0` selects that default, not an empty page. A negative limit errors with `invalid limit`.
- Other type or status values error with `invalid type` or `invalid status`.
- The result is `{tasks, limit}`. For `type=note`, newest `created_on` is first. A short page is not the full set. Call `get_task` when you have an id.

## Get

Call `get_task` with an integer `id`. `id` of 0 or less is `invalid id`. An id that is missing or belongs to someone else is not found. The tool does not say the row is forbidden. Do not retry with a different user. Do not click Notes & TODOs to open it instead.

## Do not

- Do not click Notes & TODOs.
- Do not POST /api/tran/notes-todos.
- Do not write `MORPH_MCP_TOKEN` or `JWT_SECRET` into the note.
