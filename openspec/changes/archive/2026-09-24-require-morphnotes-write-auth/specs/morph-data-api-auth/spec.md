## Purpose

Stop anonymous clients from creating, updating, or deleting MorphNotes and sibling Morph data, while published HTML pages and signed-in mutations keep working.

## ADDED Requirements

### Requirement: Mutating Morph data APIs require a session
The system MUST reject `POST`, `PUT`, `PATCH`, and `DELETE` on `/api/tran/*`, `/api/forms/*`, `/api/knowledge/*`, and `/api/graph/*` with HTTP 401 when the request has no Morph session. Research create and Research publish are included. A request with no `Authorization` header and no legacy user header is not a session.

#### Scenario: Anonymous Research create is rejected
- **WHEN** a client sends `POST /api/tran/research` with a prompt and no session
- **THEN** the response status is 401
- **AND** no research job is created

#### Scenario: Anonymous Research publish is rejected
- **WHEN** a client sends `POST /api/tran/research/:id/publish` with no session
- **THEN** the response status is 401

#### Scenario: Anonymous patch, put, and delete are rejected
- **WHEN** a client sends `PATCH`, `PUT`, or `DELETE` to a `/api/tran/*` route with no session
- **THEN** the response status is 401

#### Scenario: Sibling write prefixes are rejected
- **WHEN** a client with no session sends `POST` to `/api/forms/templates`, `/api/knowledge/files`, or `/api/graph/search`
- **THEN** each response status is 401

### Requirement: A valid Morph JWT allows the same mutations
When the request carries a valid Morph JWT for an existing user, the same mutating routes MUST proceed to their handlers and succeed as they do for a signed-in caller today.

#### Scenario: Signed-in Research create and publish succeed
- **WHEN** a client with a valid Morph JWT creates a Research job and then publishes it
- **THEN** create succeeds
- **AND** publish succeeds
- **AND** the published HTML is available at the returned public path

#### Scenario: Signed-in sibling writes proceed
- **WHEN** a client with a valid Morph JWT sends `POST` to `/api/forms/templates`, `/api/knowledge/files`, or `/api/graph/search`
- **THEN** the middleware does not respond with 401

#### Scenario: Invalid JWT does not count as anonymous success
- **WHEN** a client sends `POST /api/tran/research` with a non-empty `Authorization` bearer token that is not a valid Morph JWT
- **THEN** the response status is 401

### Requirement: Published HTML pages stay public
`GET` and `HEAD` of these paths MUST succeed without a session: `/api/tran/public/big-notes/:slug`, `/api/tran/public/timelines/:slug`, and `/api/tran/public/research/:slug`, when the slug is a single path segment. Other methods on `/api/tran/public/*`, and `GET` of an unknown kind under that prefix, MUST return 401.

#### Scenario: Anonymous GET of a published research page
- **WHEN** a client with no session requests `GET /api/tran/public/research/:slug` for a published job
- **THEN** the response status is 200
- **AND** the body is the published HTML

#### Scenario: Anonymous HEAD of published pages
- **WHEN** a client with no session sends `HEAD` to a published big-note, timeline, or research slug
- **THEN** the response status is not 401

#### Scenario: Public prefix does not allow writes
- **WHEN** a client with no session sends `POST /api/tran/public/research/:slug`
- **THEN** the response status is 401

#### Scenario: Unknown public kind is not open
- **WHEN** a client with no session sends `GET /api/tran/public/other/:slug`
- **THEN** the response status is 401

### Requirement: Private reads on the former open prefixes stay available
`GET` and `HEAD` on `/api/tran/*`, `/api/forms/*`, `/api/knowledge/*`, and `/api/graph/*` MUST remain reachable without a session so MorphNotes can still list and open records without a login redirect. Paths under `/api/tran/public/` that are not on the published-page allowlist are not private reads; those follow the published-page requirement and MUST return 401.

#### Scenario: Anonymous list read is not rejected by the middleware
- **WHEN** a client with no session sends `GET /api/tran/research`
- **THEN** the response status is not 401

#### Scenario: Anonymous form, knowledge, and graph reads are not rejected
- **WHEN** a client with no session sends `GET /api/forms/templates`, `GET /api/knowledge/files`, or `GET /api/graph/health`
- **THEN** each response status is not 401

### Requirement: Logged-in management tool calls keep working
An internal Morph AI management call to a mutating `/api/tran/*` route MUST succeed when the outer request has a valid Morph JWT. The internal call MUST carry that same `Authorization` header.

#### Scenario: Tool loop creates research with the caller JWT
- **WHEN** the management tool executor calls `POST /api/tran/research` while the outer request has a valid Morph JWT
- **THEN** the inner response is the same success the handler returns for that signed-in caller

### Requirement: Legacy header session is unchanged
A request with `X-User-ID` and no JWT MUST still count as a session for these mutating routes. Removing that fallback is out of scope.

#### Scenario: Header-only caller can still mutate
- **WHEN** a client sends `POST /api/tran/research` with `X-User-ID` set and no `Authorization` header
- **THEN** the middleware does not respond with 401
