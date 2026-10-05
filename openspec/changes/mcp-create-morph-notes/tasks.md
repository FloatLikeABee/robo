# Tasks

## 1. Owner-scoped note insert

- [x] 1.1 Add a failing test that `create_note`'s store inserts one `user_note_todo` note for the token subject's Tran user, prefixes the title with `[morph-mcp]`, starts the body with `source: morph-mcp`, rejects an empty note and over-long title or body with no insert, returns the same id on a second identical call, writes nothing when no active Tran user (user id 1 unchanged), and leaves the read-only pool unable to insert. Verify the new test fails before the insert exists.
- [x] 1.2 Implement the `mode=rw` opener and the insert that uses the `ownerClause` user subquery, with the limits and duplicate check from design.md. Verify the test from 1.1 passes, a write on the read-only connection still fails, and the journal mode stays `wal` beside an open `db.NewTranSQL` writer.

## 2. create_note tool

- [x] 2.1 Add a failing MCP client test that `tools/list` includes `whoami`, `list_my_tasks`, and `get_task` marked read-only and `create_note` not read-only, that create then `get_task` and `list_my_tasks` return the id and content, and that an expired or invalid token makes `create_note` a tool error that does not contain the token and does not insert. Verify the test fails first.
- [x] 2.2 Register `create_note` on the stdio server, recheck the token before any write, and keep `whoami`. Verify the test from 2.1 passes and `resources/list` is still empty.

## 3. Stdio client

- [x] 3.1 Extend the go-sdk stdio test so a real child process creates a note, lists it, and gets it; a second user's `get_task` is not-found and omits the first user's text; a missing token and an invalid token leave the note table unchanged. Verify the test passes and logs a redacted transcript.

## 4. Local connection docs

- [x] 4.1 Update `docs/agents/14-morph-mcp.md`, `morph/cmd/morph-mcp/README.md`, and the MCP paragraph in `morph/README.md` with the Cursor and Claude `mcp.json` entry, `MORPH_MCP_TOKEN`, `JWT_SECRET`, `TRAN_SQLITE_PATH`, the login command, and `http://localhost:3031/` as the human UI. Verify those docs name `create_note` and contain no real token.
- [x] 4.2 Update the Purpose in `openspec/specs/morph-mcp-stdio/spec.md` so it describes create as well as list and get. Verify the purpose no longer says the server only reads.

## 5. Module check

- [ ] 5.1 Run `cd morph && go build ./... && go vet ./... && go test -race ./...` and verify the command exits 0.
