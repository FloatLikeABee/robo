## 1. Persistence

- [ ] 1.1 Add SQLite table `agent_write_approval` in `morph/db/sqlite_schema.go` with the columns, 24-hour expiry, and partial unique blocker index from design.md
- [ ] 1.2 Add store methods for insert, conditional status transition, get by id, and get the blocker for a user and session
- [ ] 1.3 On store open, mark leftover `executing` rows `failed` with an outcome-unknown result and do not call any HTTP API
- [ ] 1.4 Unit-test insert, the one-blocker constraint, expiry, and the startup sweep against a temporary SQLite file

## 2. Classification and policy

- [ ] 2.1 Write table tests for `isWriteToolCall` covering GET/HEAD, PUT/PATCH/DELETE, the spec allowlist (including the `/api/formsx/` aliases and `POST /api/tran/big-notes/:id/analyze`), and the write POSTs called out in the spec (per-response analyze, generic-data analyze, tool-note read, regenerate, publish, send-email)
- [ ] 2.2 Implement `isWriteToolCall` with the same path trim and query split as `execManagementAPI`. Ignore a `side_effect` field on the model JSON
- [ ] 2.3 Add `WriteApprovalPolicy` and ship `askEveryWritePolicy` as the server default. Add a test that a substitute allow-all policy executes a write, and that the default policy does not

## 3. Tool loop pause

- [ ] 3.1 Change `chatWithManagementTools` so a classified write, when the policy does not allow it, persists the approval and returns without calling `execManagementAPI`. Reads, including allowlisted POSTs, still call it
- [ ] 3.2 Save `resume_json` (prompt, agent instructions, skill ids, round, truncated tool results) and the raw body bytes. Do not store credentials
- [ ] 3.3 Return a fixed waiting sentence plus `pending_approval` on `models.ChatResponse`, and copy that object onto `StoredChatMessage` in `persistChatExchange`
- [ ] 3.4 Skip the exact-query cache when a turn pauses or completes a write. Keep earlier reads in the same turn
- [ ] 3.5 If a blocker already exists for that user and session, return HTTP 409 with it and do not enter the tool loop. A unique-index conflict does the same and does not execute
- [ ] 3.6 If SQLite is unavailable, fail the write closed (no execution) and still allow reads
- [ ] 3.7 Resolve the owner from `auth_user_id` when middleware set it, otherwise the header `ChatHandler` already uses, and use that same id for the transcript

## 4. Decide and resume

- [ ] 4.1 Register `POST /api/chat/approvals/:id/decide` in `register_routes.go`. Do not modify `authz_middleware.go` or `authz_middleware_test.go`
- [ ] 4.2 Approve: conditional `pending` → `executing`, one `execManagementAPI` using the decide request context, store the result, then resume a fresh `ChatCompletion` loop from `resume_json` counting prior rounds toward the cap of 10
- [ ] 4.3 If resume proposes another write, persist a new approval and return it on the decide response without executing it
- [ ] 4.4 If the follow-up model call fails after the write ran, still return the tool HTTP status in the assistant text
- [ ] 4.5 Deny: conditional `pending` → `denied`, no execution, no model call, fixed cancellation sentence, HTTP 200
- [ ] 4.6 Map replay, mismatch, expiry, and in-flight status to the spec: idempotent 200, 404 for missing or other user, 410 for expired, 409 for a contradictory or in-flight decision
- [ ] 4.7 Overlay the current approval status onto `pending_approval` in `GetChatSessionHandler` without rewriting Badger messages

## 5. Verification

- [ ] 5.1 Handler tests: write pauses and does not execute; allowlisted POST executes; approve executes once and a second approve does not; deny does not execute; other user gets 404; expired gets 410; two JSON objects do not both run; a second chat on that session gets 409. Drive those tests through a local router or a stub executor so they do not depend on editing authz middleware
- [ ] 5.2 Run `cd morph && go test ./...` and `cd morph && go vet ./...`. Do not change `pkg/morphai`, the Morph frontend, or Rust crates in this change
