# Spec Delta

## MODIFIED Requirements

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

### Requirement: Cursor run instructions
The repository MUST document how to build the stdio server and a Cursor `mcp.json` snippet that launches it with `MORPH_MCP_TOKEN`, `JWT_SECRET`, and `TRAN_SQLITE_PATH`. The doc MUST name `whoami`, `list_my_tasks`, `get_task`, and `create_note`, MUST state that the HTTP JSON tool catalogs are not MCP, and MUST explain how to connect this server to local Morph at `http://localhost:3031/`, including how to obtain the session token. The documented examples MUST NOT contain a real token or secret.

#### Scenario: Docs name the client config
- **WHEN** an operator reads the MCP stdio doc
- **THEN** it shows the build command and a Cursor `mcp.json` example including `TRAN_SQLITE_PATH`
- **AND** it names `whoami`, `list_my_tasks`, `get_task`, and `create_note`
- **AND** it states that the human UI is `http://localhost:3031/` and how to obtain the token
- **AND** it states that `/ai/mcp-tools` style catalogs are not the Model Context Protocol

## ADDED Requirements

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
