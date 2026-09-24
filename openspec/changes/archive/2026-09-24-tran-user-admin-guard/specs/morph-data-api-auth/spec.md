## MODIFIED Requirements

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
