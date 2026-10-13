## ADDED Requirements

### Requirement: Unified notes tools take a kind
The server MUST advertise `get_note`, `list_notes`, `update_note`, `append_to_note`, `archive_note`, and `search_my_notes` in addition to the existing tools. `create_note` MUST accept `kind` (`note`, `todo`, `big_note`, `timeline`, `research`, `stick`, `task`). When `kind` is omitted, `create_note` MUST keep creating a Notes & TODOs `note` and MUST still return an integer `id` so existing clients keep working. Unified get, update, append, and archive MUST accept an id of the form `kind:rowid`. Old tool names (`list_my_tasks`, `get_task`, `create_stick_note`, `list_stick_notes`, `get_stick_note`, `create_timeline`, `list_timelines`, `get_timeline`, `create_story`, `list_stories`, `get_story`, `create_research`, `list_research`, `get_research`) MUST remain as aliases. Every kind MUST be scoped to the token subject. Another user's id MUST be not found and MUST NOT say forbidden.

#### Scenario: Create each kind then get as owner
- **WHEN** the caller creates one record of each kind
- **THEN** `get_note` with that `kind:id` returns the caller's title
- **AND** another user's `get_note` for the same id is not found

#### Scenario: Omitted kind stays a note
- **WHEN** a client calls `create_note` with only title and body
- **THEN** a `user_note_todo` note is stored
- **AND** the result includes an integer `id`

#### Scenario: Alias still lists stick notes
- **WHEN** the caller created a stick note through `create_note` with kind `stick`
- **THEN** `list_stick_notes` includes that row for the owner
- **AND** another user does not see it

### Requirement: Create is idempotent when a key is sent
`create_note` MUST accept optional `idempotency_key`. A second create for the same owner and key MUST return the existing record and MUST NOT insert another row.

#### Scenario: Same key returns the same id
- **WHEN** the caller creates a note twice with the same idempotency key
- **THEN** both results name the same record
- **AND** only one row exists

### Requirement: Update honors expected_version
`update_note` MUST apply the patch only when `expected_version` is omitted or matches the stored version. A mismatch MUST be `conflict` and MUST NOT change the row. A successful update MUST increment version.

#### Scenario: Stale version is conflict
- **WHEN** the caller updates with an expected_version that is not the current version
- **THEN** the result is `conflict`
- **AND** the stored body is unchanged

### Requirement: Archive is a soft delete
`archive_note` MUST set `archived_at` on the owner's record. Later `get_note`, `list_notes`, and `search_my_notes` MUST treat that id as not found. Another user's archive MUST be not found.

#### Scenario: Archived note disappears
- **WHEN** the owner archives `note:id`
- **THEN** `get_note` for that id is not found
- **AND** `list_notes` does not include it

### Requirement: Tags and search are owner only
Notes MAY carry lowercase tags (at most 10, each at most 32 characters). `search_my_notes` MUST search only the caller's unarchived notes with a sanitized FTS query. Another user's title or body MUST NOT appear. A malformed query MUST NOT raise a database error; it MUST return an empty page or `invalid_input`.

#### Scenario: Search does not leak
- **WHEN** user A and user B each have a note that matches the query
- **THEN** A's `search_my_notes` returns only A's hit
- **AND** the result does not include B's title or body

#### Scenario: Quoted junk is safe
- **WHEN** the query contains FTS operators or quotes
- **THEN** the tool returns an empty page or `invalid_input`
- **AND** the process stays up

### Requirement: Stick and task rows have an owner
`CaseTask` rows MUST store `owner_user_id`. Existing rows MUST be backfilled to the first Admin plat user, or the oldest plat user when no Admin exists. MCP list, get, create, update, append, and archive for kinds `stick` and `task` MUST constrain that column to the token subject. The admin CaseTask HTTP list MUST still return every row so the existing board does not go empty for the admin.

#### Scenario: Other user cannot get a stick note
- **WHEN** user A created a stick note
- **THEN** user B's `get_stick_note` and `get_note` for that id are not found

#### Scenario: Admin list still returns legacy rows
- **WHEN** the migration has backfilled existing CaseTask rows
- **THEN** `GET /api/tran/case-tasks` still returns those rows
- **AND** the response still includes title and dates

### Requirement: Whoami reports token scopes
`whoami` MUST include the token's scopes. For a PAT it MUST also include the token expiry when one is stored. A session JWT MUST be treated as having at least `notes:read` and `notes:write`.

#### Scenario: PAT whoami
- **WHEN** the process is a PAT with `notes:read` and `notes:write`
- **THEN** `whoami` lists those scopes
- **AND** it includes an expiry when the token has one

## MODIFIED Requirements

### Requirement: Identity fails closed
The server MUST treat `MORPH_MCP_TOKEN` as either a Morph session JWT or a personal access token. A JWT MUST be verified with `morph/auth` the way the Morph API does: HS256 signature against `JWT_SECRET`, `exp`, and in production the token-lifetime cap from #24. A PAT MUST be verified against the hashed `agent_token` row and MUST NOT require `JWT_SECRET` in the client environment. If the token is missing, invalid, expired, revoked, or has no subject, the process MUST exit non-zero before writing any MCP message, or the tool call MUST be an error, and the error MUST NOT contain the token text. A user id supplied without a valid token MUST NOT be accepted. If the token is a JWT and `JWT_SECRET` is empty or equal to the built-in development default the API substitutes when the variable is unset, the process MUST exit non-zero before writing any MCP message, and the error MUST NOT contain the secret. When `MORPH_ENV` is production and the token is a JWT, startup MUST also refuse a JWT secret that fails the API's production secret rules (length, placeholder, repeated character) and a `JWT_EXPIRY_HOURS` value the API would refuse, and MUST refuse a token whose lifetime exceeds the production maximum. An unrecognized `MORPH_ENV` MUST refuse startup. After the token verifies, the process MUST confirm the token subject still exists in `plat_users` and MUST exit non-zero when it does not. Each tool call MUST verify the token again, including revoke, expiry, and that the subject still exists in `plat_users`. A deleted user MUST fail the next tool call.

#### Scenario: Missing token
- **WHEN** the process starts without a token
- **THEN** it exits non-zero
- **AND** stdout is empty
- **AND** stderr states that the token is required

#### Scenario: Invalid token is not echoed
- **WHEN** the process starts with a token that does not verify
- **THEN** it exits non-zero
- **AND** stdout is empty
- **AND** stderr does not contain the token text

#### Scenario: Development JWT secret is refused
- **WHEN** the process starts with a JWT (not a PAT) and `JWT_SECRET` empty or set to the built-in development default
- **THEN** it exits non-zero
- **AND** stdout is empty
- **AND** stderr names `JWT_SECRET` and does not include the secret value

#### Scenario: Deleted user cannot start
- **WHEN** the process starts with a token whose subject is not in `plat_users`
- **THEN** it exits non-zero
- **AND** stdout is empty

#### Scenario: Expired token fails the tool call
- **WHEN** a tool is called with a token that has expired since startup
- **THEN** the tool result is an error
- **AND** the error does not contain the token text

#### Scenario: Production rejects a weak secret and a long-lived token
- **WHEN** `MORPH_ENV` is production, the token is a JWT, and `JWT_SECRET` is shorter than the API minimum, a placeholder, or a repeated character, or the token's lifetime exceeds the production maximum
- **THEN** the process exits non-zero before any MCP message
- **AND** stderr does not contain the secret or the token

#### Scenario: Deleted user fails the next call
- **WHEN** a tool is called after the token subject has been removed from `plat_users`
- **THEN** the tool result is an error
- **AND** the error does not contain the token text
