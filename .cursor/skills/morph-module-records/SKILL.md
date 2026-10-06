---
name: morph-module-records
description: Create, list, and get Morph stick notes, timelines, stories, research, and event logs with morph-mcp. Use when an outside agent should write or check those modules. Notes and TODOs stay on create_note, list_my_tasks, and get_task.
---

# Morph module records

Drive these Morph modules yourself through morph-mcp. Do not click MorphNotes. Do not POST `/api/tran/case-tasks`, `/api/tran/timelines`, `/api/tran/big-notes`, `/api/tran/research`, or `/api/sheetx/events-info` yourself. Notes and TODOs still use `create_note`, `list_my_tasks`, and `get_task`.

If the tool you need is not available, stop and report that morph-mcp is not connected. Do not invent `MORPH_MCP_TOKEN`, do not read `.env`, and do not write the token into a record or the chat.

The tools take no user-id argument. They act as the verified session user. Stored titles for stick notes, timelines, stories, and research start with `[morph-mcp]`.

## When

- Stick note: `create_stick_note`, `list_stick_notes`, `get_stick_note`. Shared board. Title is required. `start_at` and `end_at` are optional RFC3339; empty means the current UTC hour and the next hour.
- Timeline: `create_timeline`, `list_timelines`, `get_timeline`. Title and content are required. This stores the text. It does not run the timeline generator.
- Story: `create_story`, `list_stories`, `get_story`. Title or idea is required. This stores the text. It does not run the story generator.
- Research: `create_research`, `list_research`, `get_research`. Prompt is required. This stores the text with status complete. It does not run the research generator.
- Event log: `create_event_log`, `list_event_logs`, `get_event_log`. Title and time are required. These call the Morph events API. They do not open Badger. If the API is down, report the error. Do not create a note instead.
- Check means list, then get by id. A missing id, or another user's timeline, story, or research, is not found.

List `limit` defaults to 50 and is capped at 100. A short page is not the full set.
