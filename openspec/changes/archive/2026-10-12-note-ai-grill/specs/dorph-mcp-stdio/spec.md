## MODIFIED Requirements

### Requirement: Tools and the notes app resource
The server MUST advertise the tools capability and the resources capability. `tools/list` MUST include `whoami`, `list_my_tasks`, `get_task`, `create_note`, and `note_grill`. `whoami`, `list_my_tasks`, and `get_task` MUST be marked read-only. `create_note` and `note_grill` MUST NOT be marked read-only. `resources/list` MUST succeed and MUST include the Dorph notes app at `ui://dorph/notes` and the Dorph grill app at `ui://dorph/grill`, each with MIME type `text/html;profile=mcp-app`. The resources capability MUST NOT advertise `listChanged`. The server MUST NOT expose prompts, Engi project tools, or BK TCP tools.

#### Scenario: Three read-only tools
- **WHEN** a client calls `tools/list` after initialize
- **THEN** the list includes `whoami`, `list_my_tasks`, `get_task`, `create_note`, and `note_grill`
- **AND** `whoami`, `list_my_tasks`, and `get_task` are marked read-only
- **AND** `create_note` and `note_grill` are not marked read-only

#### Scenario: Resources list is the notes app
- **WHEN** a client calls `resources/list` after initialize
- **THEN** the call succeeds
- **AND** the list includes `ui://dorph/notes` and `ui://dorph/grill`
- **AND** each of those resources has MIME type `text/html;profile=mcp-app`
- **AND** initialize capabilities do not set resources `listChanged` to true

## ADDED Requirements

### Requirement: Grill tool is session scoped
`note_grill` MUST run only after the token verifies, as the verified session user. It MUST accept actions to start an interview from a title, record an answer, skip a question, and finish early. It MUST NOT take a user id. A grill id that is missing or belongs to another user MUST be reported as not found, and the error MUST NOT include the draft or say the row is forbidden. The tool MUST NOT write a Notes & TODOs row. Storing the note remains `create_note`. Calling the tool MUST NOT require the client to read the UI resource first, and the tool result MUST be structured grill state rather than an HTML document.

#### Scenario: Start returns a question
- **WHEN** a client calls `note_grill` to start with a non-empty title
- **THEN** the result is grill state for that session user
- **AND** the state includes a question about the title
- **AND** the result is not an HTML document

#### Scenario: Another user's grill id
- **WHEN** a client asks for a grill id started for a different user
- **THEN** the tool result is an error that the grill is not found
- **AND** the error does not include the draft

#### Scenario: Grill does not insert a note
- **WHEN** a client starts a grill, answers, and finishes
- **THEN** no Notes & TODOs row is written
- **AND** `create_note` is still required to store the draft

### Requirement: Grill tools declare the grill app
`note_grill` MUST declare `_meta.ui.resourceUri` equal to `ui://dorph/grill`, with visibility that includes both the model and the app, and the same URI on the flat `_meta["ui/resourceUri"]` key. `whoami` MUST NOT declare a UI resource. `list_my_tasks`, `get_task`, and `create_note` MUST keep `_meta.ui.resourceUri` equal to `ui://dorph/notes`.

#### Scenario: Grill tool points at the grill app
- **WHEN** a client calls `tools/list` after initialize
- **THEN** `note_grill` has `_meta.ui.resourceUri` `ui://dorph/grill`
- **AND** that tool is visible to the model and to the app
- **AND** `whoami` has no UI resource URI
- **AND** `list_my_tasks`, `get_task`, and `create_note` still point at `ui://dorph/notes`

### Requirement: Grill app HTML is closed
`resources/read` of `ui://dorph/grill` MUST return one content item whose MIME type is `text/html;profile=mcp-app` and whose text is a single HTML document. That document MUST NOT contain an external URL (`http:`, `https:`, or a protocol-relative URL). It MUST NOT contain an HTML-sink assignment (`innerHTML`, `insertAdjacentHTML`, `outerHTML`, or `document.write`). It MUST include a content security policy that blocks external connections, frames, and objects. It MUST call `note_grill` by name through the host bridge (`tools/call`). It MUST use an Arial or equivalent sans-serif font stack on a dark background with light text. The document MUST NOT contain a session token, a JWT secret, or `DORPH_MCP_TOKEN`. The resource metadata MUST NOT name an external connect, script, or frame origin.

#### Scenario: Read the grill app
- **WHEN** a client calls `resources/read` for `ui://dorph/grill`
- **THEN** the content MIME type is `text/html;profile=mcp-app`
- **AND** the HTML names `note_grill` and `tools/call`
- **AND** the HTML uses Arial or an equivalent sans-serif stack and a dark background
- **AND** the HTML contains no external URL, no HTML-sink assignment, and no session token
