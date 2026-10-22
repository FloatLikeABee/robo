## MODIFIED Requirements

### Requirement: Admin user management

Dorph SHALL provide admin-only user management over the `plat_users` store: list users, update a user's email/password/admin flag, and delete/deactivate a user. **New end-user accounts SHALL be created through invitation-code redemption, not through a DorphNotes admin UI.** Direct `POST /api/admin/users` MAY remain for programmatic/bootstrap use but SHALL NOT be exposed in DorphNotes. Admin invitation-code creation SHALL require an authenticated admin.

#### Scenario: Admin lists users via API

- **WHEN** an authenticated admin calls `GET /api/admin/users`
- **THEN** Dorph returns the user list as today

#### Scenario: DorphNotes no longer hosts user CRUD UI

- **WHEN** a user navigates DorphNotes Settings
- **THEN** there is no Users page for creating or deleting platform accounts

#### Scenario: Non-admin is rejected

- **WHEN** a non-admin (or unauthenticated caller) invokes any `/api/admin/users` or invite-admin operation
- **THEN** Dorph responds with a forbidden/unauthorized error and makes no change

#### Scenario: Last admin is protected

- **WHEN** an admin attempts to delete or demote the only remaining admin account
- **THEN** Dorph rejects the operation and keeps at least one admin

## ADDED Requirements

### Requirement: Self-service account update

Dorph SHALL allow an authenticated user to update **their own** `plat_users` username and password via `PATCH /api/auth/me`. Username changes MUST enforce uniqueness. Password changes MUST require the current password.

#### Scenario: Update own username

- **WHEN** an authenticated user PATCHes `{ "username": "newname" }` to `/api/auth/me`
- **THEN** Dorph updates the username if it is unique and returns the updated public profile

#### Scenario: Duplicate username rejected

- **WHEN** an authenticated user PATCHes a username already taken by another account
- **THEN** Dorph responds with `409 Conflict` and leaves the account unchanged

#### Scenario: Update own password

- **WHEN** an authenticated user PATCHes `{ "current_password": "...", "password": "..." }`
- **THEN** Dorph verifies the current password, stores a new bcrypt hash, and returns success

#### Scenario: Wrong current password

- **WHEN** the current password does not match
- **THEN** Dorph responds with `401 Unauthorized` and does not change the password

### Requirement: Invitation code lifecycle

Dorph SHALL persist invitation codes and support admin creation and public one-time redemption that provisions a new non-admin `plat_users` row with generated username and password.

#### Scenario: Admin creates invite code

- **WHEN** an authenticated admin calls `POST /api/admin/invite-codes`
- **THEN** Dorph stores a new unused code with creator id and created timestamp
- **AND** returns the code string once in the response body

#### Scenario: Redeem invite code

- **WHEN** an unauthenticated caller POSTs `{ "code": "<secret>" }` to `/api/invite/redeem`
- **THEN** Dorph marks the code used, creates a new verified non-admin user with generated username/password, and returns `{ username, password }` once

#### Scenario: Double redeem blocked

- **WHEN** a caller redeems a code that is already used
- **THEN** Dorph responds with `410 Gone` or `404 Not Found` and creates no user
