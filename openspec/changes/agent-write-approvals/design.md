## Context

See proposal.md for why this change exists. Behavior is specified in `specs/agent-write-approvals/spec.md`.

Today `chatWithManagementTools` (`morph/handlers/management_chat.go`) loops up to 10 rounds. Each round calls `ChatCompletion`, parses the first JSON object (`parseManagementCall` / `ExtractJSONObject`), and `execManagementAPI` runs it immediately. The in-memory `[]ai.DashScopeMessage` slice and the gin context die when the HTTP handler returns. Only the final prose is stored, via `persistChatExchange`, in Badger. A one-line tool log lives in an in-process cache (`managementSessionSnapshot`). `POST /api/chat` returns one JSON `models.ChatResponse`. There is no chat SSE handler. `morph/frontend/src/lib/aiProgress.js` is a client ticker while that POST is in flight.

Chat sessions are Badger keys `chat_sess:` / `chat_msg:`. Relational data, including `morph_agent_context_cache`, is SQLite through `TranSQL` (`morph/db/sqlite_schema.go`), one writer connection. `pending_form.go` is an in-memory map keyed only by user.

`execManagementAPI` forwards the live request's `Authorization` and `X-User-*` headers. `AuthzMiddleware` sets `auth_user_id` and copies the resolved user onto `X-User-ID`. `ChatHandler` still reads `X-User-ID` and falls back to `"admin"`. This change does not edit `authz_middleware.go`.

Issue #57 and PR #71 are not sources of tool metadata. PR #71 changes `pkg/morphai` providers and does not touch the tool loop. PR #65's stdio MCP server exposes `whoami` only. HTTP catalogs such as `/ai/mcp-tools` are prompt-facing lists, not a `side_effect` registry.

## Goals / Non-Goals

**Goals:**

- Stop the management loop before a classified write, persist the call on the user and session, and resume only after that user approves.
- Leave a policy function the trusted preset (#62) can replace, without shipping presets.
- Keep resume independent of one goroutine and of `ai.DashScopeMessage`, so a later helper (#56) can attach a write to the parent session.

**Non-Goals:**

- Approval cards, composer blocking in the UI (#55), and preset storage (#62).
- Helper runtime. The row may record an empty `helper_id`; nothing sets it here.
- Reading or trusting `side_effect` (#57). The loop does not grow a catalog client.
- Authz middleware, header-impersonation removal, or changes to `execManagementAPI` credential forwarding (#22, #23).
- `pkg/morphai` provider adapters (PR #71).
- Sheet-draft, registration, and image-generator branches in `ChatHandler`. They return before the management loop and already have their own confirmation cards.

## Decisions

### 1. Classify with method plus a path allowlist

**Choice:** `isWriteToolCall(method, path string) bool` in a new `morph/handlers/agent_approval.go`. Normalize the method. GET, HEAD, and OPTIONS are reads. PUT, PATCH, and DELETE are writes. POST is a write unless the path matches the allowlist in the spec. Match exact paths, the `/api/formsx/` alias of each `/api/sheetx/` entry, and exactly one pattern: `/api/tran/big-notes/:id/analyze`. Run the same trim and `?query` split as `execManagementAPI` before classifying. The model JSON is `method`, `path`, `query`, and `body` only. Ignore any `side_effect` field on that JSON.

**Why this matches the code:** The catalog the model sees is a prompt string (`managementToolCatalog`), not structured metadata. Handlers disagree with a keyword guess. `POST /api/tran/big-notes/:id/analyze` returns markdown and does not update a row. `POST /api/tran/big-notes/:id/responses/:responseId/analyze` and `POST /api/tran/generic-data/:id/analyze` both `UPDATE`. `POST /api/tran/tool-notes/:id/read` sets `ReadAt`. `POST /api/skills/improve`, `POST /api/tran/extract-json`, `POST /api/tran/case-tasks/ai-draft`, `POST /api/tran/generic-data/extract`, graph search, and the SheetX / Content Maker draft and web-search routes do not persist. A substring rule would mis-classify those.

**Rejected:**

- Catalog metadata as the source of truth. No HTTP catalog in this repo carries `side_effect`. PR #65 does not list management routes. Blocking #34 on that catalog couples this story to a later MCP/Bridge story (#57).
- A `side_effect` flag the model may set. The model would be able to label `DELETE` as a read and skip approval. Server metadata may override the heuristic only when #57 passes it in from the catalog, not from the tool JSON. This story's function does not take that argument, so a caller cannot accidentally thread the model flag through.
- Treat every POST as a write. That would ask before graph search and skill drafts, which the product still wants to run immediately.

### 2. Policy hook is ask-or-allow, separate from classification

**Choice:** A `WriteApprovalPolicy` with one method, `Allow(userID string, call managementCall) bool`. The loop calls it only after `isWriteToolCall` is true. The production value is `askEveryWritePolicy`, which always returns false. `Handlers` holds the policy so a test can substitute an allow-all policy and prove the branch executes. The constructor used by the server keeps the ask-every-write policy. #62 replaces the implementation and reads a per-user preset. It does not rewrite classification.

**Rejected:** Folding trusted mode into `isWriteToolCall`. Trusted mode still classifies a call as a write; it only skips the ask for a low-risk subset. Mixing them forces #62 to fork the allowlist. A single hook that also accepts `side_effect` was rejected for the same reason as decision 1: nothing trustworthy sets that field yet.

### 3. Persist a SQLite row keyed by user and session, not by the loop

**Choice:** Table `agent_write_approval` on the embedded SQLite store (`sqlite_schema.go`). Columns: `id`, `user_id`, `session_id`, `helper_id` (empty string in this story), `status`, `method`, `path`, `query`, `body` (raw JSON bytes, not re-encoded), `resume_json`, `http_status`, `result_body`, `decision`, `created_at`, `updated_at`, `expires_at`, `decided_at`. Status is `pending`, `executing`, `executed`, `denied`, `expired`, or `failed`. A partial unique index on `(user_id, session_id)` where status is `pending` or `executing` enforces one blocker per session. `expires_at` is `created_at` plus 24 hours.

`resume_json` holds the interrupted turn: user prompt, agent instructions, skill ids, round index, and truncated tool results already executed. Cap the blob (about 200KB) by dropping the oldest tool bodies first. It does not hold `Authorization`, cookies, or `X-User-*` headers. The approval row and the Badger transcript for that turn MUST use the same resolved user id (`auth_user_id` when middleware set it, otherwise the header `ChatHandler` already uses).

If two chats insert a blocker for the same user and session, the unique index rejects the second insert. That request MUST NOT execute the write. It returns HTTP 409 with the blocker that won.

If `TranMySQL` is nil, a classified write returns an error and is not executed. Reads still run.

**Rejected:**

- In-memory map like `pending_form.go`. Restart drops the approval, and the map is per user, so two sessions share one slot and a helper cannot target the parent session after the goroutine is gone.
- Badger beside chat messages. Transcript rows are append-only. Approval needs a conditional status update. SQLite is already the relational store and runs with one writer, which makes the conditional update the lock.
- Storing the bearer token so approve can run later with the original credential. That copies a secret into `./data`. The decide request's live gin context is passed to `execManagementAPI` instead.
- Keying the row by a loop id or goroutine. Helpers (#56) must enqueue on the parent user and session. `helper_id` is stored so that story can label the row without a second table; this story never sets it.

### 4. Resume with a fresh model turn plus the saved transcript

**Choice:** Approving does not restore the `messages` slice. It marks the row `executing`, calls `execManagementAPI` with the decide request's context, stores the HTTP status and truncated body, marks `executed`, then starts another `ChatCompletion` loop. The first prompt is the normal management instructions plus `resume_json` and the approved `TOOL_RESULT`. Rounds already spent count toward `managementToolMaxRounds` (10). The approved call itself still runs even when the round budget is already exhausted; in that case there is no further model call, and the response reports the tool status. If a later round classifies a write, the loop pauses again and the decide response carries the new `pending_approval`.

The wait text on the original chat response is a fixed sentence that includes method and path and says nothing has changed. Deny does not call the model: a fixed sentence says the change was cancelled and nothing was modified.

Do not call `persistManagementCachesAsync` for a turn that paused or that executed a write. A cached final answer would skip a later identical ask.

**Rejected:**

- Restore the DashScope message slice and continue that loop object. The slice is process-local, so a later decide request cannot see it. It also freezes the catalog prompt and ties resume to `ai.DashScopeMessage` while PR #71 is moving the client. The gin context from the original chat is already finished, so it cannot forward a current JWT (#22 will require that JWT on mutating `/api/tran` calls).
- A fresh turn that only sees the short session tool log. That log is one line per call. The model would re-fetch or repeat the write. The saved transcript is what makes a fresh turn safe.
- Let deny re-enter the model. The model can answer with another write JSON, and the user just refused. A fixed cancellation cannot do that.

### 5. Decide is idempotent on the approval id

**Choice:** `POST /api/chat/approvals/:id/decide` with JSON `{"decision":"approve"|"deny"}`.

- Approve of `pending`, not expired: one conditional `UPDATE` from `pending` to `executing`. Only the winner executes. The loser, while status is still `executing`, gets HTTP 409.
- Approve of `executed` or `failed` when a result body is already stored: HTTP 200 with that stored outcome. No second `execManagementAPI`.
- Approve of `denied`: HTTP 409. No execution.
- Deny of `pending`: conditional `UPDATE` to `denied`. No execution and no model call.
- Deny of `denied`: HTTP 200 with the same cancellation text.
- Deny of `executed` or `executing`: HTTP 409.
- Expired `pending`: conditional `UPDATE` to `expired`, HTTP 410, no execution. Loading a session or starting a chat also expires a row whose `expires_at` has passed, so it stops blocking.
- Missing id, or a row whose `user_id` is not the caller: HTTP 404 and the same error text. Do not use 403.
- Process start: `UPDATE` leftover `executing` rows to `failed` with an "outcome unknown" result and do not call the API. The session is then free.

There is no automatic retry of `executing`. A retry can double-create when the call succeeded and the process died before the result was saved.

The caller is `auth_user_id` from the middleware context when it is set, otherwise the `X-User-ID` header `ChatHandler` already reads. Do not add a new header trust path. Do not change `resolveUserScope`.

`GetChatSessionHandler` overlays the row's current `status` onto a stored `pending_approval` so a decided card does not still look pending after reload. Badger messages stay append-only.

### 6. `pending_approval` rides the existing JSON chat body

**Choice:** Add `PendingApproval` to `models.ChatResponse` and `models.StoredChatMessage` as `pending_approval`. `chatWithManagementTools` returns the reply string plus an optional payload (only `ChatHandler` calls it). `persistChatExchange` copies the payload onto the assistant message. `POST /api/chat` stays `c.JSON`. Do not add an SSE or streaming response. The client ticker is not a carrier; if a stream is added later, its last event must be this same JSON object and it must end before any write runs.

A session that already has a blocker returns HTTP 409 and the existing payload, and does not enter the tool loop.

**Rejected:** A side channel (websocket, polling-only resource, or the progress ticker). #55 renders the card from the chat payload. A second channel would be out of date with the transcript. Polling can be added later; the id is on the message.

### 7. Several writes are several approvals, one at a time

**Choice:** Keep the current parser: one JSON object per model reply. A write ends that HTTP turn. After approve, the next write is a new row. Do not add a batch parser or a multi-call card. An array or a second object is not executed (a failed parse returns the prose, which is today's behavior; a second object after a valid first object is ignored).

**Rejected:** Execute every read in the reply, then ask for the writes as a batch. The parser does not return a list, and #55's card is one method/path/body. Batching would approve calls the user has not seen as separate steps.

## Risks / Trade-offs

- [Allowlist drifts from handlers] → The spec lists the paths that were checked in code, including the two analyze routes that differ. Tests table-drive those edges. A new POST is a write until someone adds it.
- [Approved POST is not idempotent, and a crash lands between success and the result save] → No retry of `executing`. Startup marks it failed and tells the user the outcome is unknown.
- [Fresh resume repeats the write] → The transcript includes the `TOOL_RESULT` and says the call already ran. A repeat is a new approval, not a silent second execution.
- [24h body is stale] → The user approves the stored bytes, not a re-read of the record. Expiry bounds the window.
- [409 blocks ordinary chat on that session] → Matches the composer block #55 will add, including for clients that skip the UI. Deny or expiry clears it. Other sessions still chat.
- [Identity is only as strong as the current middleware] → Until #23, `X-User-ID` can still authenticate. This story scopes rows to whatever user id that middleware resolved. It does not close the header hole and does not edit those files.
- [`execManagementAPI` still defaults a missing `X-User-ID` to `admin`] → Left untouched on purpose. Approve forwards the decide request after middleware has set the header. Changing the default is #23's area.
- [Read-only POSTs still pass through authz] → The allowlist skips approval only. After #22, `POST /api/graph/search` and other mutating methods need a session. `execManagementAPI` forwards the live `Authorization` header. Do not add an auth bypass for allowlisted paths.
- [SQLite single connection] → Conditional updates queue. That is the idempotency lock. No extra lock service.
- [Resume model fails after the write] → The row is already `executed`. The response reports the HTTP status in prose instead of returning 500 with no record of the write.

## Migration Plan

1. Ship the SQLite table with schema init. Existing databases gain it on next open (`CREATE TABLE IF NOT EXISTS`).
2. Deploy the API. New writes pause. In-flight loops from the previous process are not resumed; they already executed, which is the old behavior.
3. Rollback: stop calling the gate. Leftover `pending` rows are inert. Dropping the table is optional and only after no client depends on `pending_approval`.

No frontend release is required for the API contract. #55 consumes it later.

## Open Questions

These do not change the state machine, the resume shape, or the task breakdown. They are constants the PO can override before implementation:

- Expiry is 24 hours. A shorter window is a different `expires_at` offset.
- Deny uses a fixed sentence and does not ask the model for a follow-up. A model sentence is acceptable only if that call cannot emit a tool JSON.
- `POST /api/sheetx/events-info/ai-draft` (and the formsx alias) is draft-only in Event Logs but is not in `managementToolCatalog`, so it stays a write. Add it to the allowlist if the PO wants it to run without a card.
