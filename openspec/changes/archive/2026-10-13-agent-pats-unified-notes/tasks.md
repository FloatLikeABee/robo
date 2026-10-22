## 1. Schema

- [x] 1.1 Add `agent_token`, `note_tag`, `note_idempotency`, and `my_notes_fts` to the API SQLite ensure path, plus `owner_user_id`, `board_kind`, `version`, and `archived_at` on `CaseTask` and version/archive columns on the other note tables. Backfill CaseTask owners once to the first Admin (else oldest plat user).
- [x] 1.2 Add a schema test that a legacy CaseTask row gets an owner and that `GET /api/tran/case-tasks` still lists title and dates.

## 2. Personal access tokens

- [x] 2.1 Store helpers: create (hash, default scopes, 90-day expiry), list, revoke, lookup by id. Never persist the plaintext. Constant-time compare.
- [x] 2.2 HTTP `POST/GET/DELETE /api/agent-tokens` on a session JWT. Tests: create returns the secret once, list is owner-scoped, revoke of another user's id is not found, missing bearer is 401.
- [x] 2.3 dorph-mcp PAT path: `IdentityFromEnv` / `CheckStartup` / `RecheckFromEnv` accept `dorph_pat_` without `JWT_SECRET`. Revoked, expired, and deleted-user tokens fail the next call. JWT path still refuses an empty or default secret. `whoami` includes scopes. Tests cover PAT start, revoke, expiry, and `forbidden_scope` on create.

## 3. Owner-scoped stick and task

- [x] 3.1 Change stick create/list/get to take the token subject, write `owner_user_id` and `board_kind`, and hide other users. HTTP create sets owner when the session is present. Admin list stays unfiltered.
- [x] 3.2 Update existing stick tests and add an other-user not-found case.

## 4. Unified notes

- [x] 4.1 Implement create/get/list/update/append/archive for every kind, with idempotency key, `expected_version` conflict, tags, archive, and FTS upsert. `create_note` without kind stays a note with an integer `id`. Register the new tools; keep aliases.
- [x] 4.2 Owner-scope tests for every kind. Idempotency, conflict, archive, and `search_my_notes` isolation (including a junk FTS query).

## 5. Grill

- [x] 5.1 Accept all seven kinds on start. Kind-specific write prompt. Test timeline start and unknown kind.

## 6. Docs and module check

- [x] 6.1 Update `docs/agents/14-dorph-mcp.md` and `dorph/cmd/dorph-mcp/README.md` with PAT curl, scopes, unified tools, and a PAT `mcp.json` that has no `JWT_SECRET`. Do not edit `16-dorph-for-agents.md` or the skills pack.
- [x] 6.2 `cd dorph && go vet ./... && go test ./...` exits 0. Dorph frontend tests still pass (no SPA change expected).
