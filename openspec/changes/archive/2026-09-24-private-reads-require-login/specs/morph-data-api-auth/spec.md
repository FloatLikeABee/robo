## REMOVED Requirements

### Requirement: Private reads on the former open prefixes stay available
**Reason**: Issue #69 requires a Morph session for private reads on a hosted instance. Anonymous list, detail, and download access is the leak this change closes.
**Migration**: Sign in. The Morph SPA sends the session bearer token. Published HTML stays on the existing public page routes.

## ADDED Requirements

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

## MODIFIED Requirements

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
