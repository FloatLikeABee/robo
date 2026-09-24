# morph-data-api-auth Specification

## Purpose

Stop anonymous clients from creating, updating, or deleting MorphNotes and sibling Morph data, while published HTML pages and signed-in mutations keep working.

## Requirements

### Requirement: Mutating Morph data APIs require a session
The system MUST reject `POST`, `PUT`, `PATCH`, and `DELETE` on `/api/tran/*`, `/api/forms/*`, `/api/knowledge/*`, and `/api/graph/*` with HTTP 401 when the request has no Morph session. Research create and Research publish are included. A Morph session is a valid Morph JWT for an existing user. `X-User-ID`, `X-User-Role`, `X-User-Roles`, `X-User-Email`, and `X-User-Permissions` are not a session.

#### Scenario: Anonymous Research create is rejected
- **WHEN** a client sends `POST /api/tran/research` with a prompt and no session
- **THEN** the response status is 401
- **AND** the research row count is unchanged

#### Scenario: Anonymous Research publish is rejected
- **WHEN** a client sends `POST /api/tran/research/:id/publish` with no session
- **THEN** the response status is 401

#### Scenario: Anonymous patch, put, and delete are rejected
- **WHEN** a client sends `PATCH`, `PUT`, or `DELETE` to a `/api/tran/*` route with no session
- **THEN** the response status is 401

#### Scenario: Sibling write prefixes are rejected
- **WHEN** a client with no session sends `POST` to `/api/forms/templates`, `/api/knowledge/files`, or `/api/graph/search`
- **THEN** each response status is 401

#### Scenario: Identity headers are not a session
- **WHEN** a client sends `POST /api/tran/research` with `X-User-ID` and `X-User-Role: admin` and no `Authorization` header
- **THEN** the response status is 401
- **AND** the research row count is unchanged

### Requirement: A valid Morph JWT allows the same mutations
When the request carries a valid Morph JWT for an existing user, the same mutating routes MUST proceed to their handlers and succeed as they do for a signed-in caller today, except `POST /api/tran/users`, `PUT /api/tran/users/:id`, and `DELETE /api/tran/users/:id`. Those three routes MUST still reach their handlers for a valid session, and a non-admin session MUST then be rejected with 403 as specified by tran-user-admin. Identity headers sent with that token MUST NOT change the user id or role taken from the token.

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

#### Scenario: Invalid bearer plus identity header is rejected
- **WHEN** a client sends `POST /api/tran/research` with a non-empty `Authorization` bearer token that is not a valid Morph JWT and with `X-User-ID` set
- **THEN** the response status is 401
- **AND** the research row count is unchanged

#### Scenario: Spoofed identity headers do not override the token
- **WHEN** a client sends a mutating Morph data request with a valid Morph JWT for one user and `X-User-ID` plus `X-User-Role: admin` for a different user
- **THEN** the middleware does not respond with 401
- **AND** the attached user id and role are the token user's, not the header values

#### Scenario: A signed-in non-admin does not succeed at Tran user administration
- **WHEN** a client with a valid Morph JWT for a non-admin sends `POST /api/tran/users`
- **THEN** the response status is 403

### Requirement: Published HTML pages stay public
`GET` and `HEAD` of `/api/tran/public/{kind}/{slug}` MUST succeed without a session when `kind` is exactly `big-notes`, `timelines`, or `research`, `slug` is one non-empty path segment that does not contain `/`, is not `.` or `..`, and is not empty, and the stored record for that slug is published. A record is published only when its published slug is non-empty after trimming. The match is case-sensitive and ignores the query string. Any other shape under `/api/tran/public/` that reaches the session middleware, including an empty segment (`//`), a `%2F` that decodes to an extra segment, a dot segment, a different kind case, or a method other than `GET` or `HEAD`, MUST return 401 without a session. A trailing slash is not a public match; the router may redirect it to the exact slug instead of returning the published HTML. An unpublished record MUST NOT be returned for a guessed slug.

#### Scenario: Anonymous GET of a published research page
- **WHEN** a client with no session requests `GET /api/tran/public/research/:slug` for a published job
- **THEN** the response status is 200
- **AND** the body is the published HTML

#### Scenario: Anonymous HEAD of published pages
- **WHEN** a client with no session sends `HEAD` to a published big-note, timeline, or research slug
- **THEN** the response status is not 401

#### Scenario: Query string does not change the public match
- **WHEN** a client with no session requests `GET /api/tran/public/research/:slug` with a query string for a published job
- **THEN** the response status is 200

#### Scenario: Public prefix does not allow writes
- **WHEN** a client with no session sends `POST /api/tran/public/research/:slug`
- **THEN** the response status is 401

#### Scenario: Unknown public kind is not open
- **WHEN** a client with no session sends `GET /api/tran/public/other/:slug`
- **THEN** the response status is 401

#### Scenario: Loose path shapes are not public
- **WHEN** a client with no session requests `GET` of `/api/tran/public/research//slug`, `/api/tran/public/research/slug/extra`, `/api/tran/public/Research/slug`, `/api/tran/public/research/.`, or `/api/tran/public/research/..`
- **THEN** each response status is 401

#### Scenario: Trailing slash is not served as the published page
- **WHEN** a client with no session requests `GET /api/tran/public/research/:slug/`
- **THEN** the response body is not the published HTML
- **AND** the response is either 401 or a redirect whose location is the exact slug with no trailing slash

#### Scenario: Encoded slash is not a single slug
- **WHEN** a client with no session requests `GET /api/tran/public/research/slug%2Fextra`
- **THEN** the response status is 401

#### Scenario: Unpublished research is not reachable by slug
- **WHEN** a client with no session requests `GET /api/tran/public/research/:slug` and no research row is published under that slug
- **THEN** the response status is not 200
- **AND** the body does not include the unpublished record

### Requirement: Private reads require a session
The system MUST reject `GET` and `HEAD` on `/api/tran/*`, `/api/forms/*`, `/api/knowledge/*`, and `/api/graph/*` with HTTP 401 when the request has no Morph session. That includes lists, details, downloads, graph health, graph search, and the personal-data lists `/api/tran/users`, `/api/tran/members`, `/api/tran/employees`, and `/api/tran/contacts`. Paths that match the published-page allowlist are not private reads.

#### Scenario: Anonymous list and detail reads are rejected
- **WHEN** a client with no session sends `GET` or `HEAD` to `/api/tran/research` or `GET /api/tran/research/:id`
- **THEN** each response status is 401
- **AND** the body does not include the record

#### Scenario: Anonymous form, knowledge, and graph reads are rejected
- **WHEN** a client with no session sends `GET /api/forms/templates`, `GET /api/knowledge/files`, or `GET /api/graph/health`
- **THEN** each response status is 401

#### Scenario: Anonymous personal-data lists are rejected
- **WHEN** a client with no session sends `GET /api/tran/users`, `GET /api/tran/members`, `GET /api/tran/employees`, or `GET /api/tran/contacts`
- **THEN** each response status is 401

#### Scenario: Signed-in private read proceeds
- **WHEN** a client with a valid Morph JWT sends `GET /api/tran/research`
- **THEN** the response status is not 401

### Requirement: Graph health does not write or leak connection details
`GET /api/graph/health` MUST NOT create or alter schema. It MUST NOT include the Neo4j URI or a raw connection error in the response body. Schema for the graph knowledge tables is created when the Tran store opens.

#### Scenario: Health response omits the URI and the raw error
- **WHEN** an authenticated client requests `GET /api/graph/health`
- **THEN** the JSON body does not contain a Neo4j URI
- **AND** the JSON body does not contain a raw driver or connection error string

### Requirement: Logged-in management tool calls keep working
An internal Morph AI management call to a mutating `/api/tran/*` route MUST succeed when the outer request has a valid Morph JWT. The internal request MUST carry that same `Authorization` header value, exactly. The internal request MUST NOT authenticate by copying `X-User-ID` or `X-User-Role`.

#### Scenario: Tool loop creates research with the caller JWT
- **WHEN** the management tool executor calls `POST /api/tran/research` while the outer request has a valid Morph JWT
- **THEN** the inner request's `Authorization` header equals the outer request's `Authorization` header
- **AND** the inner response is the same success the handler returns for that signed-in caller

#### Scenario: Tool loop does not authenticate with an identity header
- **WHEN** the management tool executor calls `POST /api/tran/research` and the outer request has no `Authorization` header
- **THEN** the inner response status is 401
