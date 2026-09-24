## 1. Read-only store and identity

- [x] 1.1 Add a failing test that seeds two users with `db.NewTranSQL` (real schema), opens the MCP connection read-only while the writer stays in WAL mode, and shows: list/get return only the caller's `user_note_todo` rows; another user's id is not-found without that title; a `CaseTask` title is absent; insert on the read-only connection fails; a missing database file is not created.
- [x] 1.2 Add failing tests for startup: empty `JWT_SECRET` and `config.DefaultJWTSecret` are refused without echoing the secret; a token subject missing from `plat_users` fails; `DecodeToken` on an expired token is the per-call failure and the error omits the token.
- [x] 1.3 Implement the read-only opener, the email → `User.UserID` lookup (no user-id `1`, no insert), list/get queries, secret refusal, and `plat_users` existence check. Re-run the tests from 1.1 and 1.2 and confirm they pass.

## 2. MCP tools

- [x] 2.1 Add a failing in-memory client test that `tools/list` is exactly `whoami`, `list_my_tasks`, and `get_task`, each with `readOnlyHint`, and that an expired token on a later call makes `whoami` and `list_my_tasks` return a tool error.
- [x] 2.2 Register the two tools with the field set and limit cap from the spec. Update the existing handshake tests that assumed `whoami` was the only tool. Confirm `morph/mcp` still does not import `idongivaflyinfa/db` or Badger.
- [x] 2.3 Point `morph/cmd/morph-mcp` at startup checks plus the read-only database. Extend the stdio child test so a seeded database, a non-default secret, and a real JWT complete initialize, tools/list, and `list_my_tasks`. A deleted `plat_users` row exits non-zero with empty stdout.

## 3. Line reader and bundled text

- [x] 3.1 Add a failing test that `startLineReader` returns after its context is canceled while a line is unread. Make sends selectable on `ctx.Done()` and pass `t.Context()` from the stdio tests.
- [x] 3.2 In `bk/openspec/changes/remove-tools-merge-adviser-customization/proposal.md`, stop saying BK answers MCP `tools/list` / `tools/call`. In `bk/openspec/changes/archive/2026-09-24-remove-bk-tcp-mcp/tasks.md`, record pytest 9 passed and 1 pre-existing Gemini-key failure, and the pre-existing `bk/frontend` build failure on a missing module.

## 4. Docs and a real client

- [x] 4.1 Update `docs/agents/14-morph-mcp.md` and `morph/cmd/morph-mcp/README.md` with the three tools, read-only `TRAN_SQLITE_PATH`, the JWT secret rule, and a Cursor `mcp.json` example that includes no real secret.
- [x] 4.2 Build `morph-mcp` and run a go-sdk stdio client (or MCP Inspector) against a seeded database and a real signed JWT. Record the redacted tools/list and tools/call transcript.
