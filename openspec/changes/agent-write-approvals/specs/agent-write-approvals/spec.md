## Purpose

Pauses Morph AI management writes until the owning user approves them, while reads keep running, and records that decision so it can be completed once.

## ADDED Requirements

### Requirement: Write classification
The system MUST classify each management tool call before execution. GET, HEAD, and OPTIONS are reads. PUT, PATCH, and DELETE are writes. POST is a write unless its path is on the read-only POST allowlist. Classification MUST use the same path trimming and query splitting the executor uses, so a query string glued onto the path does not change the class. An unrecognized POST MUST be a write.

#### Scenario: Mutating methods are writes
- **WHEN** a management tool call uses PUT, PATCH, or DELETE
- **THEN** the call is classified as a write

#### Scenario: Ordinary GET is a read
- **WHEN** a management tool call uses GET or HEAD
- **THEN** the call is classified as a read and is executed immediately

#### Scenario: Unknown POST is a write
- **WHEN** a management tool call uses POST on a path that is not on the read-only allowlist
- **THEN** the call is classified as a write

### Requirement: Read-only POST allowlist
The system MUST treat only these POST paths as reads. Matching MUST be by path pattern, not by a substring such as "analyze" or "search".

Exact paths, including the `/api/formsx/` alias wherever `/api/sheetx/` is listed:

- `/api/graph/search`
- `/api/skills/improve`
- `/api/tran/extract-json`
- `/api/tran/generic-data/extract`
- `/api/tran/case-tasks/ai-draft`
- `/api/sheetx/ai/web-search`
- `/api/sheetx/ai/form-template-chat`
- `/api/sheetx/survey-bot/templates/ai-draft`
- `/api/composerx/ai/web-search`
- `/api/composerx/ai/composer-chat`
- `/api/composerx/ai/publish-chat`

One parameterized path:

- `/api/tran/big-notes/:id/analyze` with no extra segments

These POSTs MUST stay writes: `/api/tran/generic-data/:id/analyze`, `/api/tran/big-notes/:id/responses/:responseId/analyze`, `/api/tran/tool-notes/:id/read`, `/api/tran/big-notes/:id/regenerate`, `/api/tran/big-notes/:id/publish`, and `/api/tran/case-tasks/:id/send-email`.

#### Scenario: Graph search runs without approval
- **WHEN** the model calls POST `/api/graph/search`
- **THEN** the call is executed immediately
- **AND** no pending approval is stored

#### Scenario: Per-response analyze still asks
- **WHEN** the model calls POST `/api/tran/big-notes/9/responses/3/analyze`
- **THEN** the call is classified as a write
- **AND** it is not executed before approval

#### Scenario: All-response analyze is a read
- **WHEN** the model calls POST `/api/tran/big-notes/9/analyze`
- **THEN** the call is executed immediately

### Requirement: Ask before every write
Under the default policy the system MUST ask before every classified write. It MUST NOT execute that call while the approval is pending. It MUST persist one pending approval owned by the calling user and the chat session that proposed it. The approval MUST survive beyond the chat request that created it. The system MUST NOT store the caller's credential, session token, or authorization header on the approval.

#### Scenario: Write is stored and not executed
- **WHEN** the management tool loop proposes a classified write
- **THEN** the write is not executed
- **AND** a pending approval exists for that user and session
- **AND** the stored call can be executed later with the same method, path, query, and body

#### Scenario: An earlier read in the same turn still stands
- **WHEN** the loop has already executed a read in this turn and the next model reply is a write
- **THEN** that read is not rolled back
- **AND** the write is not executed

#### Scenario: No trusted auto-approve
- **WHEN** the default policy is in effect and the model proposes any classified write
- **THEN** the system asks for approval
- **AND** there is no user setting in this capability that skips the ask

### Requirement: Pending approval on the chat response
When a chat turn creates a pending write, that `POST /api/chat` MUST return HTTP 200 with a `pending_approval` object and an assistant message that says the assistant is waiting and that nothing has been changed yet. The object MUST include `id`, `session_id`, `method`, `path`, `expires_at`, and `status` of `pending`, plus `query` and `body` when the call has them. Morph AI chat has no server-sent event stream. The JSON body of `POST /api/chat` and of the decide response is the only carrier. The same object MUST be stored on the assistant transcript message. A repeated approve or deny whose outcome was already stored MUST use this same chat response shape.

#### Scenario: Non-streaming response carries the payload
- **WHEN** a chat turn pauses for a write
- **THEN** the JSON chat response includes `pending_approval`
- **AND** `response` tells the user the assistant is waiting
- **AND** the write's result is not in that response

#### Scenario: Transcript keeps the payload
- **WHEN** the user reloads the chat session before deciding
- **THEN** the assistant message still includes that `pending_approval`
- **AND** its `status` is the approval's current status

### Requirement: One pending approval per user and session
The system MUST allow at most one blocking approval (`pending` or `executing`) for a user and session. A later `POST /api/chat` on that session MUST NOT run management tools. It MUST return HTTP 409 with the existing `pending_approval`. Another session for the same user MUST still be able to chat. The model reply is one tool call: only the first JSON object is interpreted. The loop MUST stop on a write instead of executing further calls from that reply. A resumed loop MAY create a new approval after the previous one is no longer blocking.

#### Scenario: Second chat is refused while pending
- **WHEN** a session already has a pending approval and the user sends another chat message on that session
- **THEN** the response is HTTP 409
- **AND** the existing approval is returned
- **AND** no management tool call runs

#### Scenario: Two objects in one reply do not both run
- **WHEN** one model reply contains two JSON tool calls and the first is a write
- **THEN** only the first call becomes an approval
- **AND** the second call is not executed

#### Scenario: Other session is independent
- **WHEN** the user has a pending approval on one session and sends a chat message on a different session
- **THEN** that other session is not blocked by the first approval

### Requirement: Approve executes once and resumes
`POST /api/chat/approvals/:id/decide` with `decision` `approve` MUST execute the stored call at most once and then resume the management loop for that same user and session. Resume MUST include the tool results already produced in the interrupted turn plus the approved call's result, and MUST use the credentials on the decide request. The decide response MUST use the chat response shape. If resume proposes another write, the response MUST carry a new `pending_approval` and MUST NOT execute that new write. If the follow-up model call fails after the approved call has run, the response MUST still report that the call ran and its HTTP status. A repeated approve for an approval that already has a stored result MUST return that result and MUST NOT call the API again.

#### Scenario: Approve runs the call and continues
- **WHEN** the owning user approves a pending write
- **THEN** the stored call is executed once
- **AND** the loop continues from that result
- **AND** the decide response includes the assistant's follow-up

#### Scenario: Repeated approve does not run twice
- **WHEN** the owning user approves an approval that already has a stored result
- **THEN** the response returns the stored outcome
- **AND** the management API is not called again

#### Scenario: Resume can pause on the next write
- **WHEN** the resumed loop's next tool call is another write
- **THEN** that call is not executed
- **AND** the decide response includes a new `pending_approval`

### Requirement: Deny cancels without execution
`POST /api/chat/approvals/:id/decide` with `decision` `deny` MUST mark a pending approval denied, MUST NOT execute the call, and MUST NOT start another management tool round. The response MUST be HTTP 200 with an assistant message that the change was cancelled and nothing was modified, and without `pending_approval`. A repeated deny MUST return that same cancellation and MUST NOT execute.

#### Scenario: Deny does not call the API
- **WHEN** the owning user denies a pending write
- **THEN** the call is not executed
- **AND** the assistant message says the change was cancelled

#### Scenario: Approve after deny is rejected
- **WHEN** the owning user sends approve for an approval that is already denied
- **THEN** the response is HTTP 409
- **AND** the call is not executed

### Requirement: Owner, expiry, and in-flight decisions
The system MUST apply a decision only for the owning user. A missing approval and an approval owned by someone else MUST both be HTTP 404 with the same not-found error. A pending approval MUST expire 24 hours after it is created. Deciding an expired approval MUST return HTTP 410, MUST NOT execute, and MUST mark it expired. An expired blocking approval MUST NOT keep blocking the session. A second decide that arrives while the first approve is still executing MUST return HTTP 409 and MUST NOT start a second execution. If the process stops while an approval is `executing`, the next process start MUST mark that approval failed, MUST NOT retry the call, and MUST leave the session unblocked. The failure outcome MUST say the result is unknown.

#### Scenario: Another user cannot decide
- **WHEN** a user who does not own the approval calls decide
- **THEN** the response is HTTP 404
- **AND** the call is not executed
- **AND** the approval stays pending

#### Scenario: Expired approval does not run
- **WHEN** the owner approves an approval whose expiry time has passed
- **THEN** the response is HTTP 410
- **AND** the call is not executed

#### Scenario: Concurrent approve is single execution
- **WHEN** two approve requests for the same pending approval are handled together
- **THEN** the management call runs at most once

#### Scenario: Restart does not replay an in-flight call
- **WHEN** an approval is `executing` and the process starts again
- **THEN** the approval becomes failed without a new management call
- **AND** the session can chat again
