# agent-tokens Specification

## Purpose

Lets a signed-in Dorph user mint, list, and revoke hashed personal access tokens so an outside agent can call dorph-mcp without putting `JWT_SECRET` in the client environment.

## Requirements

### Requirement: Session user can create a personal access token
`POST /api/agent-tokens` MUST require a Dorph session JWT. It MUST accept a name and optional scopes and expiry days. When scopes are omitted, the token MUST receive `notes:read`, `notes:write`, and `pool:read`. `pool:publish` and `admin:pool` MUST NOT be granted unless the caller asked for them. The plaintext token MUST be returned once, MUST start with `dorph_pat_`, and MUST NOT be stored. The store MUST keep only a hash of the token. The default expiry MUST be 90 days.

#### Scenario: Create returns the secret once
- **WHEN** a signed-in user posts a token name
- **THEN** the response includes a `dorph_pat_` token, an id, the scopes, and an expiry
- **AND** a later list of that user's tokens does not include the plaintext token
- **AND** the stored value is not the plaintext token

#### Scenario: Unauthenticated create is refused
- **WHEN** the request has no valid session JWT
- **THEN** the response is 401
- **AND** no token row is inserted

### Requirement: Session user can list and revoke their tokens
`GET /api/agent-tokens` MUST list only the caller's tokens (id, name, scopes, created_at, expires_at, last_used_at, revoked_at) and MUST NOT include hashes or secrets. `DELETE /api/agent-tokens/:id` MUST set `revoked_at` on that caller's token. Another user's id MUST be treated as not found.

#### Scenario: List is owner scoped
- **WHEN** user A and user B each have a token
- **THEN** A's list does not include B's token id or name

#### Scenario: Revoke is owner scoped
- **WHEN** user A tries to revoke user B's token id
- **THEN** the response is not found
- **AND** B's token stays usable

### Requirement: dorph-mcp accepts a PAT without a client JWT secret
When `DORPH_MCP_TOKEN` is a live personal access token, startup MUST succeed without `JWT_SECRET` in the environment. The process MUST resolve the token subject from the hashed row and MUST confirm that subject still exists in `plat_users`. A revoked, expired, or deleted-user token MUST fail startup or the next tool call. The error MUST NOT contain the token text.

#### Scenario: PAT starts without JWT_SECRET
- **WHEN** the process starts with a live PAT and no `JWT_SECRET`
- **THEN** it becomes ready
- **AND** `whoami` returns that token's user id and scopes

#### Scenario: Revoked PAT fails the next call
- **WHEN** a running process holds a PAT that is then revoked
- **THEN** the next tool call is an error
- **AND** the error does not contain the token

#### Scenario: Expired PAT is refused
- **WHEN** the process starts or a tool is called with an expired PAT
- **THEN** the call fails
- **AND** the error does not contain the token

### Requirement: Missing scope is forbidden
A PAT that lacks `notes:read` MUST NOT run a read notes tool. A PAT that lacks `notes:write` MUST NOT run a write notes tool. The error MUST be `forbidden_scope` and MUST NOT leak another user's content.

#### Scenario: Read-only PAT cannot create
- **WHEN** a PAT with only `notes:read` calls `create_note`
- **THEN** the result is `forbidden_scope`
- **AND** no row is inserted
