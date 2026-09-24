## Purpose

Stop a signed-in non-admin from creating, rewriting, or deactivating Tran users, and stop an ambiguous login email from attaching the session to an arbitrary Tran user.

## ADDED Requirements

### Requirement: Tran user create, update, and delete require a platform admin
`POST /api/tran/users`, `PUT /api/tran/users/:id`, and `DELETE /api/tran/users/:id` MUST return 401 when the request has no Morph session. They MUST return 403 when the session belongs to a non-admin. They MUST succeed when the session belongs to a platform admin.

Platform admin MUST be the role on the server-side user record for the verified session, using the same check as `/api/admin/users`. The check MUST re-read that record. It MUST NOT treat a client header (`X-User-Role` or any other identity header), a request body field, the Tran `User.Administrator` column, or a JWT role claim by itself as admin. An empty user id MUST be 401 and MUST NOT be treated as admin.

A non-admin MUST NOT change their own administrator flag or deactivated state through `PUT /api/tran/users/:id` or through `PUT /api/tran/users/me`.

#### Scenario: No session is 401
- **WHEN** a client with no session sends `POST /api/tran/users`, `PUT /api/tran/users/:id`, or `DELETE /api/tran/users/:id`
- **THEN** each response status is 401

#### Scenario: Non-admin is 403
- **WHEN** a signed-in non-admin sends `POST /api/tran/users`, `PUT /api/tran/users/:id`, or `DELETE /api/tran/users/:id`
- **THEN** each response status is 403
- **AND** no Tran user row is created, updated, or deactivated

#### Scenario: Admin succeeds
- **WHEN** a signed-in platform admin sends `POST /api/tran/users`, `PUT /api/tran/users/:id`, or `DELETE /api/tran/users/:id` with a valid body
- **THEN** each request succeeds

#### Scenario: Header and Tran administrator flag do not grant admin
- **WHEN** a signed-in non-admin sends `PUT /api/tran/users/:id` with `X-User-Role: admin` and that caller's Tran `User.Administrator` value is true
- **THEN** the response status is 403

#### Scenario: Stale JWT admin claim is not admin
- **WHEN** a token was issued while the user was a platform admin and the server-side user record is no longer an admin
- **THEN** `POST /api/tran/users` returns 403

#### Scenario: Self-promotion is 403
- **WHEN** a signed-in non-admin sends `PUT /api/tran/users/:id` for their own Tran user id with `administrator` true
- **THEN** the response status is 403
- **AND** that row's administrator flag stays false

### Requirement: Profile email follows the login account
`PUT /api/tran/users/me` MUST NOT change the Tran user's email. An `email` field in the body MUST be ignored. When the body contains another allowed profile field, that field MAY be saved and the stored email MUST stay unchanged. When `email` is the only field that would have been written, the response MUST be 400 and the row MUST be unchanged. `administrator` and `deactivated` MUST NOT be writable on this route.

#### Scenario: Email in a profile save is ignored
- **WHEN** a signed-in user sends `PUT /api/tran/users/me` with a new email and a new last name
- **THEN** the response succeeds
- **AND** the stored email is the previous email
- **AND** the stored last name is the new last name

#### Scenario: Email-only profile save writes nothing
- **WHEN** a signed-in user sends `PUT /api/tran/users/me` with only an `email` field
- **THEN** the response status is 400
- **AND** the stored email is unchanged

#### Scenario: Profile save cannot set administrator or deactivated
- **WHEN** a signed-in non-admin sends `PUT /api/tran/users/me` with `administrator` true, `deactivated` true, and a new last name
- **THEN** the stored administrator flag stays false
- **AND** the stored deactivated flag stays false

### Requirement: Ambiguous active email resolution fails closed
When more than one active Tran `User` row matches the session email on the email column, resolution MUST fail. It MUST NOT choose one of those rows. It MUST NOT insert a new row. `GET /api/tran/users/me` and `PUT /api/tran/users/me` MUST return 409.

When no active row matches the email column and more than one active row matches the same value on LoginID, resolution MUST fail the same way, with no chosen row and no insert, and the profile routes MUST return 409.

The Notes and TODOs identity lookup MUST NOT return either matching user id in the ambiguous email case, and MUST NOT fall through to a different user id.

A single active email match MUST still resolve to that row. When no active email or LoginID row matches, the profile routes MUST still create one profile row, as they do today.

#### Scenario: Two active emails do not resolve or create
- **WHEN** two active Tran users share the session email and the client sends `GET /api/tran/users/me`
- **THEN** the response status is 409
- **AND** the Tran user row count is unchanged
- **AND** the resolved Notes and TODOs user id is neither of those two ids

#### Scenario: Two active LoginID matches do not resolve
- **WHEN** no active row matches the session email and two active rows match that value on LoginID
- **THEN** `GET /api/tran/users/me` returns 409
- **AND** no new Tran user row is created

#### Scenario: One active match still resolves
- **WHEN** exactly one active Tran user has the session email
- **THEN** `GET /api/tran/users/me` returns that user id

#### Scenario: No match still creates one profile
- **WHEN** no active Tran user matches the session email or LoginID
- **THEN** `GET /api/tran/users/me` creates one active row for that email
- **AND** a second request returns the same user id

#### Scenario: A deactivated duplicate does not make the match ambiguous
- **WHEN** one active Tran user and one deactivated Tran user share the session email
- **THEN** `GET /api/tran/users/me` returns the active user id

### Requirement: Takeover by rewriting another user's email fails
A non-admin MUST NOT be able to point another Tran user's email at their own login email and then deactivate their own Tran row. After those attempts, the attacker's resolved Tran user id MUST still be the attacker's id.

#### Scenario: Email rewrite and self-deactivate are rejected
- **WHEN** signed-in user A sends `PUT /api/tran/users/:id` to set user B's email to A's email, then deactivates A's Tran row
- **THEN** both responses are 403
- **AND** B's email is unchanged
- **AND** A's resolved Tran user id is A's id

### Requirement: Profile form does not edit email
The Morph profile form MUST show the account email as read-only and MUST NOT submit a replacement email when the user saves the rest of the profile.

#### Scenario: Saving the profile leaves email unchanged
- **WHEN** a user saves the profile form after the email control is present
- **THEN** the save request does not send a new email value

### Requirement: Tran user reads stay available to signed-in users
`GET /api/tran/users`, `GET /api/tran/users/:id`, and `GET /api/tran/users/me` MUST stay available to any signed-in user. A missing session on those reads MUST still be 401.

#### Scenario: Signed-in non-admin can list users
- **WHEN** a signed-in non-admin sends `GET /api/tran/users`
- **THEN** the response status is not 403
