# Spec Delta

## RENAMED Requirements

- FROM: `### Requirement: Tools and an empty resource list`
- TO: `### Requirement: Tools and the notes app resource`

## MODIFIED Requirements

### Requirement: Tools and the notes app resource
The server MUST advertise the tools capability and the resources capability. `tools/list` MUST include exactly four tools: `whoami`, `list_my_tasks`, and `get_task`, each marked read-only, and `create_note`, which MUST NOT be marked read-only. `resources/list` MUST succeed and return exactly one resource, the Morph notes app at `ui://morph/notes` with MIME type `text/html;profile=mcp-app`. The resources capability MUST NOT advertise `listChanged`. The server MUST NOT expose prompts, Engi project tools, or BK TCP tools.

#### Scenario: Three read-only tools
- **WHEN** a client calls `tools/list` after initialize
- **THEN** the tools are `whoami`, `list_my_tasks`, `get_task`, and `create_note`
- **AND** `whoami`, `list_my_tasks`, and `get_task` are marked read-only
- **AND** `create_note` is not marked read-only

#### Scenario: Resources list is the notes app
- **WHEN** a client calls `resources/list` after initialize
- **THEN** the call succeeds
- **AND** the only resource URI is `ui://morph/notes`
- **AND** that resource MIME type is `text/html;profile=mcp-app`
- **AND** initialize capabilities do not set resources `listChanged` to true

## ADDED Requirements

### Requirement: Note tools declare the Morph notes app
`list_my_tasks`, `get_task`, and `create_note` MUST each declare `_meta.ui.resourceUri` equal to `ui://morph/notes`, with visibility that includes both the model and the app. `whoami` MUST NOT declare a UI resource. A client that does not render MCP Apps MUST still receive the same tool result it receives today: structured note data for those three tools, and the verified user for `whoami`. Calling a tool MUST NOT require the client to read the UI resource first.

#### Scenario: Note tools point at the notes app
- **WHEN** a client calls `tools/list` after initialize
- **THEN** `list_my_tasks`, `get_task`, and `create_note` each have `_meta.ui.resourceUri` `ui://morph/notes`
- **AND** each of those tools is visible to the model and to the app
- **AND** `whoami` has no UI resource URI

#### Scenario: Plain tool result without an app host
- **WHEN** a client calls `list_my_tasks`, `get_task`, or `create_note` and does not read the UI resource
- **THEN** the result is the note data for that tool
- **AND** the result is not an HTML document

### Requirement: Notes app document is closed
`resources/read` of `ui://morph/notes` MUST return one content item whose MIME type is `text/html;profile=mcp-app` and whose text is a single HTML document. That document MUST NOT contain an external URL (`http:`, `https:`, or a protocol-relative URL). It MUST NOT contain an HTML-sink assignment (`innerHTML`, `insertAdjacentHTML`, `outerHTML`, or `document.write`). It MUST render note titles and bodies as text. It MUST include a content security policy that blocks external connections, frames, and objects. It MUST call `list_my_tasks`, `get_task`, and `create_note` by name through the host bridge (`tools/call`). The document MUST NOT contain a session token, a JWT secret, or `MORPH_MCP_TOKEN`. The resource metadata MUST NOT name an external connect, script, or frame origin.

#### Scenario: Read the notes app
- **WHEN** a client calls `resources/read` for `ui://morph/notes`
- **THEN** the content MIME type is `text/html;profile=mcp-app`
- **AND** the HTML contains no external URL and no HTML-sink assignment
- **AND** the HTML names `list_my_tasks`, `get_task`, and `create_note`
- **AND** the HTML does not contain the session token

### Requirement: MCP Apps extension is advertised
The initialize result MUST include the extension `io.modelcontextprotocol/ui` with `mimeTypes` containing `text/html;profile=mcp-app`.

#### Scenario: Extension is present
- **WHEN** a client completes initialize
- **THEN** server capabilities include extension `io.modelcontextprotocol/ui`
- **AND** that extension lists MIME type `text/html;profile=mcp-app`

### Requirement: Localhost MCP app attach steps
The MCP stdio docs MUST explain how to launch this same binary for local Morph at `http://localhost:3031/` from at least one client that renders MCP Apps, using `MORPH_MCP_TOKEN`, `JWT_SECRET`, and `TRAN_SQLITE_PATH`. The docs MUST say how to obtain the token from local Morph login, MUST state that the notes app lists, opens, and creates notes, and MUST state that a client which does not render MCP Apps still receives the plain tool results. The documented examples MUST NOT contain a real token or secret.

#### Scenario: Docs name an MCP Apps client and the plain-tools fallback
- **WHEN** an operator reads the MCP stdio doc
- **THEN** it shows a localhost launch config for an MCP-Apps-capable client with those three environment variables
- **AND** it explains how to obtain the token from local Morph login
- **AND** it states that clients without MCP Apps still use the tool results
