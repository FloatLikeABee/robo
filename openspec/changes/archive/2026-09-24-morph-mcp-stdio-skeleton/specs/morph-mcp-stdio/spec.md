## Purpose

Let a local MCP client attach to Morph over stdio, complete the handshake, and see which Morph user the server is acting as, without opening the stores the API process already holds.

## ADDED Requirements

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
The server MUST advertise the tools capability and the resources capability. `tools/list` MUST include exactly one tool, `whoami`, marked read-only. `resources/list` MUST succeed and return no resources. The server MUST NOT expose MorphNotes write tools, Engi project tools, or BK TCP tools.

#### Scenario: Whoami is the only tool
- **WHEN** a client calls `tools/list` after initialize
- **THEN** the only tool is `whoami`
- **AND** that tool is marked read-only

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
The server MUST treat `MORPH_MCP_TOKEN` as a Morph session JWT and verify it with the same HS256 secret the Morph API uses. If the token is missing, invalid, expired, or has no subject, the process MUST exit non-zero before writing any MCP message, and the error MUST NOT contain the token text. A user id supplied without a valid token MUST NOT be accepted.

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

### Requirement: Startup does not take the API's Badger lock
The stdio process MUST NOT open the Morph Badger directories or the Morph SQLite database. A running `morph-api` that holds those Badger directory locks MUST be able to keep running, and the stdio process MUST be able to start beside it.

#### Scenario: Skeleton starts without the data files
- **WHEN** the stdio process starts with a valid token and the Badger and SQLite paths are not opened
- **THEN** initialize still succeeds

### Requirement: Logs stay off stdout
Stdout MUST contain only MCP JSON-RPC messages. Diagnostic logs, including the resolved user id, MUST go to stderr. Stderr MUST NOT contain the session token.

#### Scenario: Handshake bytes are JSON-RPC
- **WHEN** a client completes initialize, tools/list, and whoami
- **THEN** every non-empty stdout line is a JSON-RPC message
- **AND** log text is on stderr
- **AND** the token is on neither stream

### Requirement: Cursor run instructions
The repository MUST document how to build the stdio server and a Cursor `mcp.json` snippet that launches it with `MORPH_MCP_TOKEN` and `JWT_SECRET`. The doc MUST state that the HTTP JSON tool catalogs are not MCP. The documented examples MUST NOT contain a real token or secret.

#### Scenario: Docs name the client config
- **WHEN** an operator reads the MCP stdio doc
- **THEN** it shows the build command and a Cursor `mcp.json` example
- **AND** it states that `/ai/mcp-tools` style catalogs are not the Model Context Protocol
