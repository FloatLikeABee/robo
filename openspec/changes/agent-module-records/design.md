## Context

See proposal.md for why. Morph AI’s management tool loop already posts JSON to `/api/tran/*` and `/api/sheetx/*` as the signed-in user. The catalog lists stick notes (`case-tasks`), stories (`big-notes`), and event logs, but not timelines or research. The fast-path text never says which user-facing name maps to which route, so “make a note” lands on Notes & TODOs. morph-mcp `create_note` is a different surface and stays on Notes & TODOs.

`TranMySQL` is the local SQL store. Story, timeline, and research creates already generate content. A stick note requires `title`, `start_at`, and `end_at`. A timeline requires pasted content, `content`, or a URL. An event log requires `title` and `time`.

## Goals / Non-Goals

**Goals:**

- One short routing block in the existing tool instructions, plus catalog lines for timelines and research.
- A test that the instructions name each module’s list, get, and create route and forbid Notes & TODOs as a substitute.

**Non-Goals:**

- New tables, new MCP tools, or a plain-text twin of stories, timelines, or research.
- Delete, publish, or email unless the operator asks.
- Info Sheets, generic data, or a language control.

## Decisions

### 1. Teach the current tool loop instead of adding tools

Add a name map to `managementToolInstructions` and the missing catalog lines. The model keeps emitting one JSON call with `method`, `path`, and `body`.

**Why not new MCP tools:** The operator is in Morph AI chat. morph-mcp is the external Notes & TODOs server, and its spec says not to POST `/api/tran/notes-todos`. These five modules are already HTTP APIs behind the chat tool loop.

**Why not a new “save text only” create:** Stories, timelines, and research are generated records. A second store would split what the MorphNotes screens show.

### 2. Check is list or get

If the operator names an id, GET that record (stick notes use `/full`). Otherwise GET the list and match the title or topic they named. Answer with found or not found and the main fields. Do not invent a checked flag.

### 3. Missing times are an assumption, not a notes-todo

A stick note without times still goes to `POST /api/tran/case-tasks`. The instruction tells the agent to pick a short window, say what it assumed, and send `start_at` and `end_at`. A timeline without a source asks for paste or a URL instead of filing a personal note. A failed create is reported as the API error.

### 4. Do not call destructive routes for these five unless asked

The catalog can keep existing delete routes for other records. The new block says not to delete, publish, or email these five unless the operator asks. Create story, timeline, or research once, then GET the new id and summarize. Do not also regenerate or publish.

## Risks / Trade-offs

- [The model still files a personal note] → The instructions forbid that substitute, and a test locks the route names and the forbid line into the prompt.
- [Story, timeline, or research create is slow] → That is the existing generator. The tool loop already allows multiple rounds, so the agent can GET after create. Do not start a second job in the same turn.
- [Event Logs is down] → The agent reports the HTTP error. It does not write a note to stand in for the event.
- [A stick note needs an assignee the operator did not name] → Omit assignees. If the API rejects the body, report that error.

## Migration Plan

No data move. Rollback is reverting the instruction text. Records created while this is on stay in their modules.

## Open Questions

None. A plain-text story that skips generation, or MCP tools for these modules, would change this plan, and both are excluded.
