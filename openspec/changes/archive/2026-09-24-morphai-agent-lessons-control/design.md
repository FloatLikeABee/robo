## Context

See proposal.md. Spec: `agent-lesson-control`.

Lessons live in SQLite `agent_lesson` and were global: `id`, `trigger`, `rule`, `source_session_id`, `created_at`, unique on `source_session_id`. Distillation already receives the chat user id and does not store it. Chat sessions live in Badger as `chat_sess:{userID}:{sessionID}`. `models.DefaultChatSessionID` is the shared string `"default"`.

`AuthzMiddleware` (`morph/main.go`, before `RegisterAPIRoutes`) rejects `/api/*` with no credentials, except login, invite redeem, `/api/tran/public/`, `/api/auth/`, and the open Morph Data prefixes (`/api/tran/`, `/api/forms/`, `/api/knowledge/`, `/api/graph/`). Lesson routes are not in those lists. `resolveUserScope` still accepts a client `X-User-ID` when there is no bearer token (and when `TranMySQL` is nil) and `applyAuthScope` copies that value into `auth_user_id` and back onto the header. A valid bearer token overwrites the header with the token subject. `ChatHandler` reads `X-User-ID` and, if it is empty, uses `"admin"`. Issue #22 owns middleware and that chat identity. This change does not edit the middleware.

`ensureTranSQLiteSchema` creates `plat_users` before lesson migration on a normal or brand-new database. `EnsureBootstrapAdmin` runs after that, so the first schema pass of a brand-new process can see zero accounts even though an admin appears later in the same process. The next process start sees that account.

## Goals / Non-Goals

**Goals:**

- Operators list, disable, and delete only their lessons.
- Prompts and the skills catalog follow only that user's enabled lessons.
- Legacy rows stay enabled, and a single-operator database keeps them, without later leaking them onto whoever remains.

**Non-Goals:**

- Settings UI (later story).
- Changing `AuthzMiddleware` or `ChatHandler`'s header fallback.
- Making skills themselves per-user.

## Decisions

### 1. One-shot claim for unowned legacy rows

**Choice:** Add `enabled INTEGER NOT NULL DEFAULT 1` and `owner_user_id TEXT NOT NULL DEFAULT ''` with the existing idempotent `sqliteAddColumnIfMissing` path. If `plat_users` is missing, add the columns and skip the claim so startup still succeeds. If the table exists and the claim has not been decided yet: zero accounts → leave rows unowned and decide later; exactly one account → assign every currently unowned row to that account and record the decision; more than one account → record the decision and leave rows unowned. After the decision is recorded, startup MUST NOT assign again.

**Collision policy:** Keep the account's existing lesson and delete the conflicting unowned row, inside the claim transaction, before the bulk assign. The harvested row is the lesson stored after the account existed. Leaving the legacy row unowned would keep a dead row on the empty-owner side of the unique index. A claim error is logged and does not fail startup.

**Why:** There is no reliable owner on the old rows. Badger keys include the user, but `"default"` is shared, so a session-id scan cannot tell whose lesson it was. The common install is one operator, and `EnsureBootstrapAdmin` can create that account only after the first schema pass, so the claim has to be willing to wait. It must not run again after that, or deleting users until one remains would dump historical unowned rules into the survivor's prompts.

**Alternatives rejected:**

- Leave `owner_user_id` blank forever. Safe, and a single-operator database loses every existing lesson until someone edits SQLite.
- Copy each legacy lesson onto every account. Leaks one operator's rules into every login.
- Re-run "if exactly one account, assign all unowned rows" on every startup. Covers the late bootstrap admin, and fails when a multi-user database later drops to one account.
- On a session collision, leave the legacy row unowned. It is never listed, and it occupies `( '', session )` so a later unowned insert for that session cannot land.
- On a session collision, overwrite the harvested lesson with the legacy rule. Throws away the lesson the account just stored.

### 2. 404 for missing and foreign lessons

**Choice:** `PATCH` and `DELETE` return 404 `{ "error": "lesson not found" }` both when the id is absent and when another user owns it. The same string is used.

**Why:** The settings screen must not learn which ids exist on other accounts.

**Alternatives rejected:**

- 403 when the row exists and is owned by someone else. Confirms the id.
- 200 with no change. The client cannot tell a failed toggle from success.

### 3. Bearer subject only, not the header

**Choice:** Lesson list, patch, delete, skills-catalog lessons, prompt injection, harvest, and the management-chat exact cache resolve the user by decoding the bearer token and loading that subject from `plat_users`. `X-User-ID` and `auth_user_id` are ignored. No bearer token → lesson routes return 401, even if the SQLite handle is nil (a header-only caller must not get 503). A bearer token with a nil store → 503 when the handler runs. Through `AuthzMiddleware`, a bearer the middleware cannot accept (no header fallback) is 401 before the handler. An invalid token or unknown subject → 401. Harvest stores that bearer user id and does not run for a header-only chat.

**Why:** #22 merged and still does not set a trusted-user flag. `resolveUserScope` verifies a bearer when the store is up, and otherwise copies a client `X-User-ID` into `auth_user_id`. Lesson code cannot prefer that context value. Re-parsing the bearer is what fails closed. The middleware file is not edited here.

**Alternatives rejected:**

- Trust `auth_user_id` after middleware. The middleware copies the spoofed header into that value.
- Trust the user id `ChatHandler` already computed. Same header, plus `"admin"` when the header is empty.

### 4. Unique on owner and source session

**Choice:** Create `idx_agent_lesson_owner_session` on `(owner_user_id, source_session_id)`, then drop `idx_agent_lesson_session`. The new index is created first so a failure does not leave the table with neither unique key.

**Why:** Every user's default chat uses session id `"default"`. A global unique session id lets only one account ever keep a lesson for it. Harvest looks up by owner and session, so the second user would be blocked or would collide.

**Alternatives rejected:**

- Keep the global unique index. The second user's default-session harvest cannot insert.
- Drop uniqueness. One user can accumulate duplicate lessons for one session.

### 5. Skills catalog shows enabled lessons only

**Choice:** `GET /api/agent-lessons` returns the caller's lessons including disabled ones (cap 500). `GET /api/skills` embeds at most 8 enabled lessons for the bearer user, and embeds none when there is no bearer user. Prompts use the same enabled-and-owned filter, cap 8.

**Why:** The catalog and the prompt answer "what the agent will follow." The dedicated list answers "what can I turn off." Disabled rows in the catalog look active, and they can crowd the 8-slot cap.

**Alternatives rejected:**

- Return disabled lessons from `GET /api/skills` with `enabled: false`. Mixes the control list into the catalog the agent context is built from.
- Remove the `lessons` field from `GET /api/skills`. The catalog already returns it; the settings screen uses `/api/agent-lessons` instead, and the field stays for enabled lessons only.

### 6. Exact-query cache fingerprints enabled lessons

**Choice:** `mgmt:exact:` is keyed by `trustedLessonUserID` plus a hash of that user's enabled lesson ids and rules. No trusted user: the cache is not read and not written. `PATCH` rejects unknown JSON fields.

**Why:** The cache is checked before the lesson block is rebuilt, with a one-hour TTL, and the write is asynchronous. Deleting `mgmt:exact:` rows on PATCH/DELETE loses to an in-flight response that still holds the old key and writes it back after the delete. A fingerprint taken when the cache is read cannot match an entry built for a different enabled set, including that late write, and it also changes when harvest adds a lesson. Invalidation would have to find every writer.

**Alternatives rejected:**

- Delete the owner's `mgmt:exact:` entries on PATCH and DELETE. The async writer can put the stale reply back, and harvest would not invalidate.
- Keep the chat header user id in the key. A header-only caller who sets `X-User-ID` to a bearer user reads that user's cached reply.

## Risks / Trade-offs

- [First decision sees several accounts] → Historical lessons stay hidden. Documented. They are not copied out later.
- [Claim waits until an account exists] → A process that creates the admin and exits before the next schema pass leaves lessons unowned until the following start. Acceptable; the following start claims them once.
- [Chat history still keys off `X-User-ID`] → Lesson harvest and the exact-query cache do not. #22 still copies an unverified header into `auth_user_id`. Other chat behavior stays there.
- [Prompt cap stays 8] → Older enabled lessons drop out of the prompt. The full list remains on `GET /api/agent-lessons`.

## Migration Plan

1. On startup, `ensureTranSQLiteSchema` adds columns, replaces the unique index, and runs the one-shot claim.
2. Rollback: stop calling the lesson routes and stop injecting lessons. Added columns can remain. The backfill decision row can remain; it only blocks a second claim.

## Open Questions

None. Thresholds for which sessions are harvested stay in the existing harness.
