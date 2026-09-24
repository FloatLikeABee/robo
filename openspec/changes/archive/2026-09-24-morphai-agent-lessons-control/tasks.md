## 1. Ownership migration

- [x] 1.1 Add `enabled` (default true) and `owner_user_id`, replace the session-only unique index with `(owner_user_id, source_session_id)`, and claim legacy rows at most once: wait while there are zero accounts, assign when the first decision sees exactly one account, and never assign after a multi-account decision. Skip the claim when `plat_users` is missing so startup still succeeds.
- [x] 1.2 Tests: existing rows stay enabled; a later single-account startup does not claim new unowned rows; a multi-account decision stays unowned after the database drops to one account; a missing `plat_users` table does not fail migration.

## 2. Operator API and identity

- [x] 2.1 `GET /api/agent-lessons` returns the bearer user's lessons including disabled. `PATCH` toggles `enabled` and `DELETE` removes a lesson. Missing and other users' ids are 404. No bearer token is 401, including when the SQLite handle is nil. `X-User-ID` and `auth_user_id` are not identity.
- [x] 2.2 Tests: patch, delete, cross-user 404, spoofed header through the current middleware, and 401 when the store is nil and only a header is sent.

## 3. Prompt and skills catalog

- [x] 3.1 Prompt injection and `GET /api/skills` include only enabled lessons for the bearer user. Harvest stores that user id and does not run for a header-only caller.
- [x] 3.2 Tests: disabled and other users' rules stay out of the prompt and the skills `lessons` array; a header without a bearer embeds no lessons.

## 4. Docs

- [x] 4.1 Document the endpoints, bearer identity, 404, one-shot legacy claim, and enabled-only catalog in `docs/agents/03-morph.md`, `docs/agents/02-ai-integration.md`, and `morph/README.md`.
