## Context

See proposal.md for why. morph-mcp is a stdio process. It verifies `MORPH_MCP_TOKEN`, reads `TRAN_SQLITE_PATH` read-only for list and get, and opens that file mode=rw for `create_note`. It must not open Badger. Stick notes (`CaseTask`), timelines, stories (`big_note`), and research are tables in that SQLite file. Event logs live in Event Logs (formx Badger), reached today through Morph `GET/POST /api/sheetx/events-info`.

## Goals / Non-Goals

**Goals:**

- Fifteen tools, three per module, registered next to the notes tools.
- Direct SQL for the four SQLite modules, scoped like `create_note`.
- Event log tools that HTTP the Morph API with the existing token.

**Non-Goals:**

- Starting story, timeline, or research generators.
- Delete, publish, email, or a new notes-todo fallback.
- Changing `create_note` or opening Badger.

## Decisions

### 1. Same bridge, separate tool names

Each module gets `create_*`, `list_*`, and `get_*`. Check is list or get. One generic tool would hide the module from the model the way the chat catalog did.

**Why not HTTP for the SQLite modules:** `create_note` writes the file the UI reads, so the record exists even when the agent is not on the Morph origin. Event logs are the exception because they are not in that file.

### 2. Store the text, do not generate

Insert the agent's title and body. Research status is `complete` so a worker does not treat the row as a job. Titles are prefixed with `[morph-mcp]` once. Stick notes have no owner column, so they stay on the shared board. Timelines, stories, and research set `user_id` from the same Tran user lookup as notes.

A stick note with no times uses the current UTC hour and the next hour, and the result includes those times.

### 3. Event logs call Morph, not Badger

`MORPH_API_BASE_URL`, default `http://127.0.0.1:9090`, plus `MORPH_MCP_TOKEN` as the bearer. Paths are `/api/sheetx/events-info` and `/api/sheetx/events-info/:id`. Errors return the status and a short body. The token is not logged and not copied into the result.

## Risks / Trade-offs

- [A story created here has no generated HTML] → The human still sees the title and text in Stories. Generation stays a MorphNotes action.
- [Event Logs or Morph API is down] → The tool errors. It does not write a note.
- [Tool-count tests assume four tools] → Update them to the new count and treat every `create_*` tool as not read-only.

## Migration Plan

No data move. Rollback is reverting the tool registration. Rows already inserted stay.

## Open Questions

None.
