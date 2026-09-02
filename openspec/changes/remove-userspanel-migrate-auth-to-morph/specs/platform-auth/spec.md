## Purpose

Defines Morph as the single platform authentication and simple user-management service: it validates credentials, issues a shared JWT used for single sign-on across every robo app, exposes session/user and permission lookups, and provides admin-only user CRUD — replacing the standalone UsersPanel service.

## ADDED Requirements

### Requirement: Morph is the sole authentication provider

Morph SHALL be the only service that validates credentials and issues platform session tokens. The platform SHALL NOT require the standalone UsersPanel service to be running for any authentication, session, permission, or user-management operation.

#### Scenario: Login without UsersPanel running

- **WHEN** a user submits valid email/username and password to Morph's `POST /api/auth/login` while no UsersPanel service is running
- **THEN** Morph verifies the password against its `plat_users` store and returns a signed JWT plus the public user profile and permissions

#### Scenario: Invalid credentials

- **WHEN** a user submits credentials that do not match an active account
- **THEN** Morph responds with `401 Unauthorized` and does not issue a token

### Requirement: UsersPanel-compatible session endpoints

Morph SHALL expose session endpoints compatible with the contract previously served by UsersPanel so that existing app auth clients work unchanged: `GET /api/auth/user` (current user), `GET /api/auth/permissions` (permission list), and `GET /api/auth/me` (user + permissions). Each SHALL accept the JWT via `Authorization: Bearer <token>`.

#### Scenario: Session lookup with a valid token

- **WHEN** an app calls `GET /api/auth/user` with a valid Bearer token issued by Morph
- **THEN** Morph returns the authenticated user's public profile in the same envelope shape consuming apps already parse

#### Scenario: Session lookup with a missing or invalid token

- **WHEN** an app calls `GET /api/auth/user`, `GET /api/auth/permissions`, or `GET /api/auth/me` without a valid Bearer token
- **THEN** Morph responds with `401 Unauthorized`

### Requirement: Shared single sign-on token across apps

The JWT issued by Morph SHALL be accepted as the single sign-on credential by all consuming apps (formx, composerx, booki, morph-engi, SharpReport, academi). Each app SHALL resolve the session by presenting the token to Morph's auth endpoints and SHALL treat Morph's auth base URL as its configured identity provider.

#### Scenario: Cross-app session reuse

- **WHEN** a user authenticates once via Morph and an embedded/companion app receives that token
- **THEN** the app validates it against Morph's `/api/auth/*` endpoints and grants access without a second login

#### Scenario: Auth base URL configuration

- **WHEN** an app reads its auth provider configuration (via `USERS_PANEL_BASE_URL` or an equivalent Morph auth base-URL setting)
- **THEN** the value resolves to the Morph API and the app performs all auth calls against it

### Requirement: Admin user management

Morph SHALL provide admin-only user management over the `plat_users` store: list users, create a user with email/password and admin flag, update a user's email/password/admin flag, and delete/deactivate a user. These operations SHALL require an authenticated admin and SHALL be reachable from a Morph admin UI.

#### Scenario: Admin creates a user

- **WHEN** an authenticated admin submits a valid email and password to `POST /api/admin/users`
- **THEN** Morph creates the account (hashing the password) and returns the new user's public profile

#### Scenario: Non-admin is rejected

- **WHEN** a non-admin (or unauthenticated caller) invokes any `/api/admin/users` operation
- **THEN** Morph responds with a forbidden/unauthorized error and makes no change

#### Scenario: Last admin is protected

- **WHEN** an admin attempts to delete or demote the only remaining admin account
- **THEN** Morph rejects the operation and keeps at least one admin

### Requirement: Retire the standalone UsersPanel service and its integrations

The standalone UsersPanel project SHALL be removed, and platform components SHALL NOT depend on an external UsersPanel service, its `:5001` port, its admin SPA, or reverse-proxy/sync integrations that targeted it.

#### Scenario: No build or deploy artifacts reference UsersPanel

- **WHEN** the dev launcher, deployment manifests, and build scripts are inspected
- **THEN** none of them start, build, package, or order services around a UsersPanel service, and the `UsersPanel/` project directory no longer exists

#### Scenario: Legacy UsersPanel proxy path

- **WHEN** a client requests a legacy `/api/users-panel/*` auth path against Morph (if the compatibility route is retained)
- **THEN** Morph serves it from its own local auth handlers rather than forwarding to any external UsersPanel service
