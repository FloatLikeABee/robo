# morph-mcp-stdio Specification

## Purpose

Let a local MCP client attach to Morph over stdio, complete the handshake, see which Morph user the server is acting as, and list, read, and create that user's own notes. List and get open SQLite read-only. create_note writes through a separate connection on the same file. The process does not open Badger.

## Requirements

### Requirement: Stdio initialize
The Morph module MUST build a stdio server that speaks MCP JSON-RPC on stdin and stdout. An `initialize` request at protocol revision 2025-06-18 MUST return that revision, a serverInfo name and version, and server capabilities. The process MUST NOT listen on a network port. A newer protocol revision the server supports MAY be negotiated when the client requests it.

#### Scenario: Initialize at 2025-06-18
- **WHEN** a client sends `initialize` with protocol version 2025-06-18
- **THEN** the response protocolVersion is 2025-06-18
- **AND** serverInfo includes a non-empty name and version
- **AND** capabilities include tools and resources

#### Scenario: No network listener
- **WHEN** the server process is running a stdio session
- **THEN** it does not bind a TCP or UDP port for MCP

### Requirement: Tools and an empty resource list
The server MUST advertise the tools capability and the resources capability. `tools/list` MUST include exactly four tools: `whoami`, `list_my_tasks`, and `get_task`, each marked read-only, and `create_note`, which MUST NOT be marked read-only. `resources/list` MUST succeed and return no resources. The server MUST NOT expose prompts, Engi project tools, or BK TCP tools.

#### Scenario: Three read-only tools
- **WHEN** a client calls `tools/list` after initialize
- **THEN** the tools are `whoami`, `list_my_tasks`, `get_task`, and `create_note`
- **AND** `whoami`, `list_my_tasks`, and `get_task` are marked read-only
- **AND** `create_note` is not marked read-only

#### Scenario: Resources list is empty
- **WHEN** a client calls `resources/list` after initialize
- **THEN** the call succeeds
- **AND** the resource list is empty

### Requirement: Whoami returns the verified Morph user
`whoami` MUST return the Morph user id, email, username, and roles taken from the verified session token. The result MUST NOT include the token or a password.

#### Scenario: Call whoami
- **WHEN** a client calls `whoami` with a verified token for a known user
- **THEN** the result id, email, username, and roles match that token's claims
- **AND** the token string is absent from the result

### Requirement: Identity fails closed
The server MUST treat `MORPH_MCP_TOKEN` as a Morph session JWT and verify it with `morph/auth` the way the Morph API does: HS256 signature against `JWT_SECRET`, `exp`, and in production the token-lifetime cap from #24. If the token is missing, invalid, expired, or has no subject, the process MUST exit non-zero before writing any MCP message, and the error MUST NOT contain the token text. A user id supplied without a valid token MUST NOT be accepted. If `JWT_SECRET` is empty or equal to the built-in development default the API substitutes when the variable is unset, the process MUST exit non-zero before writing any MCP message, and the error MUST NOT contain the secret. When `MORPH_ENV` is production, startup MUST also refuse a JWT secret that fails the API's production secret rules (length, placeholder, repeated character) and a `JWT_EXPIRY_HOURS` value the API would refuse, and MUST refuse a token whose lifetime exceeds the production maximum. An unrecognized `MORPH_ENV` MUST refuse startup. After the token verifies, the process MUST confirm the token subject still exists in `plat_users` and MUST exit non-zero when it does not. Each tool call MUST verify the token again, and an expired, invalid, or over-long production token MUST be a tool error that does not contain the token text.

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

#### Scenario: Production rejects a weak secret and a long-lived token
- **WHEN** `MORPH_ENV` is production and `JWT_SECRET` is shorter than the API minimum, a placeholder, or a repeated character, or the token's lifetime exceeds the production maximum
- **THEN** the process exits non-zero before any MCP message
- **AND** stderr does not contain the secret or the token

### Requirement: Startup does not take the API's Badger lock
The stdio process MUST NOT open the Morph Badger directories. It MUST NOT call `db.NewTranSQL` and MUST NOT run schema migrations or set the SQLite journal mode. It MUST open `TRAN_SQLITE_PATH` read-only (`mode=ro` and `query_only`) for reads so a running `morph-api` can keep that file in WAL mode. A write on that read-only connection MUST fail. `create_note` MUST write through a separate connection opened `mode=rw` on that same existing file, with a busy timeout, and MUST NOT create the database file or change the journal mode. A note that connection commits MUST be visible to the read-only connection.

#### Scenario: Read-only connection rejects writes
- **WHEN** the MCP read connection is open on a writable SQLite file
- **THEN** an insert or schema change on that connection fails

#### Scenario: Reader works beside a WAL writer
- **WHEN** another connection holds the database in WAL mode and the MCP read connection reads it
- **THEN** the read succeeds
- **AND** the writer can still commit

#### Scenario: Note writer does not migrate
- **WHEN** `create_note` commits a note while another connection holds the file in WAL mode
- **THEN** the journal mode is unchanged
- **AND** the read-only connection can read the new note

### Requirement: Logs stay off stdout
Stdout MUST contain only MCP JSON-RPC messages. Diagnostic logs, including the resolved user id, MUST go to stderr. Stderr MUST NOT contain the session token.

#### Scenario: Handshake bytes are JSON-RPC
- **WHEN** a client completes initialize, tools/list, and whoami
- **THEN** every non-empty stdout line is a JSON-RPC message
- **AND** log text is on stderr
- **AND** the token is on neither stream

### Requirement: Cursor run instructions
The repository MUST document how to build the stdio server and a Cursor `mcp.json` snippet that launches it with `MORPH_MCP_TOKEN`, `JWT_SECRET`, and `TRAN_SQLITE_PATH`. The doc MUST name `whoami`, `list_my_tasks`, `get_task`, and `create_note`, MUST state that the HTTP JSON tool catalogs are not MCP, and MUST explain how to connect this server to local Morph at `http://localhost:3031/`, including how to obtain the session token. The documented examples MUST NOT contain a real token or secret.

#### Scenario: Docs name the client config
- **WHEN** an operator reads the MCP stdio doc
- **THEN** it shows the build command and a Cursor `mcp.json` example including `TRAN_SQLITE_PATH`
- **AND** it names `whoami`, `list_my_tasks`, `get_task`, and `create_note`
- **AND** it states that the human UI is `http://localhost:3031/` and how to obtain the token
- **AND** it states that `/ai/mcp-tools` style catalogs are not the Model Context Protocol

### Requirement: List and get return only the caller's notes and todos
`list_my_tasks` and `get_task` MUST read `user_note_todo` rows for the verified JWT subject only. Every such query MUST constrain the WHERE clause with that subject (`plat_users.id`), joined to an active Tran `User` (`Deactivated = 0`) on email. They MUST NOT use a request user id, an email claim as a substitute for the subject, an `X-User-ID` header, or a default user id, and they MUST NOT create a `User` row. The owner predicate MUST live in one function so it can later call the shared visibility helper from #69. `list_my_tasks` MUST accept optional `type` (`all`, `note`, or `todo`), optional `status` (`all`, `open`, or `done`), and optional `limit`. The default limit is 50 and the applied limit MUST NOT exceed 100. Each task object MUST include id, title, status (`open` or `done`), item type, and the text body stored in SQLite, plus deadline, created, and updated timestamps when present. `get_task` for an id the caller does not own, or that does not exist, MUST be a not-found tool error and MUST NOT say the row is forbidden or include the other user's title or body. MorphNotes `CaseTask` rows MUST NOT be returned.

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

### Requirement: Create note stores the caller's note
`create_note` MUST insert one note owned by the active Tran user joined from the verified token subject on email, using the same owner predicate as list and get. The row MUST be item type `note`, open, and without a deadline. The tool result MUST include that note's id, title, item type, status, and body. A following `list_my_tasks` and `get_task` MUST return that same id and content. The tool MUST NOT call `ExposeRecord`, MUST NOT honor a request user id or identity header, and MUST NOT create a Tran user or use user id 1.

#### Scenario: Create then list and get
- **WHEN** a client with a valid session calls `create_note` with a title and body
- **THEN** the result includes a new id, a title that contains the caller's title, and a body that contains the caller's body
- **AND** `get_task` with that id returns the same content
- **AND** `list_my_tasks` includes that id when the note fits in the requested page
- **AND** the row is owned by that session's Tran user

#### Scenario: No Tran user writes nothing
- **WHEN** the token subject exists in `plat_users` but has no active Tran user
- **THEN** `create_note` is an error
- **AND** no note row is inserted
- **AND** user id 1 does not gain a note

#### Scenario: Another user cannot get the note
- **WHEN** user A has created a note and user B calls `get_task` with that id
- **THEN** the result is not-found
- **AND** the result does not include A's title or body

### Requirement: Agent notes are attributable
A note created by `create_note` MUST store a title that starts with `[morph-mcp]` and a body that starts with `source: morph-mcp`, followed by the caller's text when they sent any. The marker MUST be visible in the stored fields list and get return. The tool MUST NOT add a database column.

#### Scenario: Stored note shows the marker
- **WHEN** a client creates a note with title "Shift report" and a body
- **THEN** the stored title starts with `[morph-mcp]` and contains "Shift report"
- **AND** the stored body starts with `source: morph-mcp` and contains the caller's body

### Requirement: Create note fails closed
`create_note` MUST verify the session token before writing. A missing, invalid, or expired token MUST be a tool error, or a non-zero process exit before any tool runs, and MUST NOT insert a row. The error MUST NOT contain the token text. An empty title and an empty body MUST NOT insert a row.

#### Scenario: Invalid token writes nothing
- **WHEN** `create_note` is invoked with a missing or invalid token
- **THEN** the call fails
- **AND** the note table is unchanged
- **AND** the error does not contain the token

#### Scenario: Empty note is rejected
- **WHEN** the client omits both title and body
- **THEN** `create_note` is an error
- **AND** no row is inserted

### Requirement: Create note limits input
`create_note` MUST reject a caller title longer than 200 characters or a caller body longer than 32000 characters without inserting a row. A second call with the same title and body for the same user MUST return the existing note's id and MUST NOT insert another row. Errors MUST NOT include another user's title, body, or the token.

#### Scenario: Overlong input writes nothing
- **WHEN** the title or body exceeds its limit
- **THEN** `create_note` is an error
- **AND** no row is inserted

#### Scenario: Same note is not duplicated
- **WHEN** the client calls `create_note` twice with the same title and body
- **THEN** both results return the same id
- **AND** only one row exists for that text
