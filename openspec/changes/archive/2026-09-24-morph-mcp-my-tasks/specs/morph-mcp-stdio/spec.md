# Spec Delta

## MODIFIED Requirements

### Requirement: Tools and an empty resource list
The server MUST advertise the tools capability and the resources capability. `tools/list` MUST include exactly three tools, `whoami`, `list_my_tasks`, and `get_task`, each marked read-only. `resources/list` MUST succeed and return no resources. The server MUST NOT expose write tools, prompts, Engi project tools, or BK TCP tools.

#### Scenario: Three read-only tools
- **WHEN** a client calls `tools/list` after initialize
- **THEN** the tools are `whoami`, `list_my_tasks`, and `get_task`
- **AND** each tool is marked read-only

#### Scenario: Resources list is empty
- **WHEN** a client calls `resources/list` after initialize
- **THEN** the call succeeds
- **AND** the resource list is empty

### Requirement: Identity fails closed
The server MUST treat `MORPH_MCP_TOKEN` as a Morph session JWT and verify it with the same HS256 secret the Morph API uses. If the token is missing, invalid, expired, or has no subject, the process MUST exit non-zero before writing any MCP message, and the error MUST NOT contain the token text. A user id supplied without a valid token MUST NOT be accepted. If `JWT_SECRET` is empty or equal to the built-in development default the API substitutes when the variable is unset, the process MUST exit non-zero before writing any MCP message, and the error MUST NOT contain the secret. On startup the process MUST confirm the token subject still exists in `plat_users` and MUST exit non-zero when it does not. Each tool call MUST verify the token again, and an expired or invalid token MUST be a tool error that does not contain the token text.

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
- **WHEN** the process starts with `JWT_SECRET` empty or set to the built-in development default
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

### Requirement: Startup does not take the API's Badger lock
The stdio process MUST NOT open the Morph Badger directories. It MUST NOT call `db.NewTranSQL` and MUST NOT run schema migrations or set the SQLite journal mode. It MUST open `TRAN_SQLITE_PATH` read-only (`mode=ro` and `query_only`) so a running `morph-api` can keep that file in WAL mode. A write on the MCP connection MUST fail.

#### Scenario: Read-only connection rejects writes
- **WHEN** the MCP connection is open on a writable SQLite file
- **THEN** an insert or schema change on that connection fails

#### Scenario: Reader works beside a WAL writer
- **WHEN** another connection holds the database in WAL mode and the MCP connection reads it
- **THEN** the read succeeds
- **AND** the writer can still commit

### Requirement: Cursor run instructions
The repository MUST document how to build the stdio server and a Cursor `mcp.json` snippet that launches it with `MORPH_MCP_TOKEN`, `JWT_SECRET`, and `TRAN_SQLITE_PATH`. The doc MUST name `whoami`, `list_my_tasks`, and `get_task`, and MUST state that the HTTP JSON tool catalogs are not MCP. The documented examples MUST NOT contain a real token or secret.

#### Scenario: Docs name the client config
- **WHEN** an operator reads the MCP stdio doc
- **THEN** it shows the build command and a Cursor `mcp.json` example including `TRAN_SQLITE_PATH`
- **AND** it names the three tools
- **AND** it states that `/ai/mcp-tools` style catalogs are not the Model Context Protocol

## ADDED Requirements

### Requirement: List and get return only the caller's notes and todos
`list_my_tasks` and `get_task` MUST read `user_note_todo` rows whose `UserID` is the active Tran `User` row for the token email (case-insensitive trimmed email, `Deactivated = 0`). They MUST NOT use a request user id, an `X-User-ID` header, or a default user id, and they MUST NOT create a `User` row. `list_my_tasks` MUST accept optional `type` (`all`, `note`, or `todo`), optional `status` (`all`, `open`, or `done`), and optional `limit`. The default limit is 50 and the applied limit MUST NOT exceed 100. Each task object MUST include id, title, status (`open` or `done`), item type, and the text body stored in SQLite, plus deadline, created, and updated timestamps when present. `get_task` for an id the caller does not own, or that does not exist, MUST be a not-found tool error and MUST NOT say the row is forbidden or include the other user's title or body. MorphNotes `CaseTask` rows MUST NOT be returned.

#### Scenario: List returns only mine
- **WHEN** the database has notes or todos for the caller and for another user
- **THEN** `list_my_tasks` returns only the caller's rows
- **AND** the result does not include the other user's title or body

#### Scenario: Get returns the caller's task
- **WHEN** the caller requests `get_task` with their own `user_note_todo` id
- **THEN** the result includes that row's id, title, status, and body

#### Scenario: Get of another user's task is not found
- **WHEN** the caller requests `get_task` with another user's `user_note_todo` id
- **THEN** the tool result is not-found
- **AND** the result does not include that row's title or body
- **AND** the result does not say the call is forbidden

#### Scenario: Limit is capped
- **WHEN** the caller requests `list_my_tasks` with a limit above 100
- **THEN** the number of returned rows is at most 100

#### Scenario: Status filter keeps done rows out
- **WHEN** the caller requests `list_my_tasks` with status `open`
- **THEN** completed rows are absent
