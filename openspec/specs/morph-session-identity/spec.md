# morph-session-identity Specification

## Purpose

Define which requests count as a Morph session so chat, admin, and other protected API routes cannot be opened by spoofed identity headers.

## Requirements

### Requirement: Protected routes reject identity headers without a JWT
Chat routes under `/api/chat` and admin routes under `/api/admin` MUST return 401 when the request has no valid Morph JWT, including when `X-User-ID` and `X-User-Role: admin` are set. The same rejection applies to `/api/data-collector`.

#### Scenario: Header-only admin cannot call chat
- **WHEN** a client sends `POST /api/chat` with `X-User-ID` and `X-User-Role: admin` and no `Authorization` header
- **THEN** the response status is 401

#### Scenario: Header-only admin cannot call admin
- **WHEN** a client sends `GET /api/admin/users` with `X-User-ID` and `X-User-Role: admin` and no `Authorization` header
- **THEN** the response status is 401

### Requirement: A valid JWT is the only identity
When a request carries a valid Morph JWT for an existing user, the user id and role attached to the request MUST come from that token's user record. `X-User-ID`, `X-User-Role`, `X-User-Roles`, `X-User-Email`, and `X-User-Permissions` sent by the client MUST NOT replace that user or role. An employee token plus `X-User-Role: admin` MUST NOT gain admin access.

#### Scenario: Spoofed admin role does not upgrade an employee token
- **WHEN** a client sends `GET /api/admin/users` with a valid Morph JWT for a non-admin user and `X-User-ID` plus `X-User-Role: admin` for someone else
- **THEN** the response status is 403
- **AND** the attached user id is the token user's id

#### Scenario: Token user is visible on a protected route
- **WHEN** a client sends `POST /api/chat` with a valid Morph JWT and a different `X-User-ID`
- **THEN** the response status is not 401
- **AND** the attached user id is the token user's id
- **AND** the attached role is the token user's role

### Requirement: OPTIONS is not an authenticated method
An `OPTIONS` request to a protected `/api` path MUST NOT be rejected with 401 for lack of a session.

#### Scenario: Chat preflight is not 401
- **WHEN** a client sends `OPTIONS /api/chat` with no `Authorization` header and no identity headers
- **THEN** the response status is not 401

### Requirement: CORS allow-list names the headers in use
Cross-origin responses from the Morph API MUST set `Access-Control-Allow-Headers` to an explicit list that includes `Authorization`, `Content-Type`, and `Accept`. The list MUST NOT be `*` and MUST NOT include `X-User-ID`, `X-User-Role`, `X-User-Roles`, `X-User-Email`, or `X-User-Permissions`.

#### Scenario: Preflight allow-list names Authorization
- **WHEN** a client sends `OPTIONS /api/chat` with an `Origin` header
- **THEN** `Access-Control-Allow-Headers` contains `Authorization`
- **AND** `Access-Control-Allow-Headers` is not `*`
- **AND** `Access-Control-Allow-Headers` does not contain `X-User-ID`
