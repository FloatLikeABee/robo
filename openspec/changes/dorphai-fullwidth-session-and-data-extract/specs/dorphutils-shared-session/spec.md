## Purpose

Lets Dorph Utils reuse the Dorph AI login the same way Dorph Data does, without a second sign-in.

## ADDED Requirements

### Requirement: Dorph Utils is signed in after Dorph AI login
If the user is already signed in to Dorph AI, Dorph Utils SHALL treat them as signed in. Opening Dorph Utils from the Dorph AI Apps menu MUST carry the session automatically. Dorph Utils MUST NOT require a separate login form when that Dorph AI session is valid.

#### Scenario: Open Dorph Utils from Dorph AI Apps
- **WHEN** a user is signed in to Dorph AI
- **AND** they open Dorph Utils from the Dorph AI Apps menu
- **THEN** Dorph Utils is authenticated
- **AND** they can use Utils modules without signing in again

#### Scenario: Dorph Data remains same-session
- **WHEN** a user is signed in to Dorph AI
- **AND** they open Dorph Data
- **THEN** Dorph Data continues to use the same Dorph AI session with no extra login

### Requirement: Utils does not drop a valid Dorph AI session
Dorph Utils MUST NOT clear the shared session solely because a session-validation request fails due to a network or proxy error. It MAY clear the session only when Dorph AI reports the credential is invalid or the user Signs out.

#### Scenario: Transient auth check failure
- **WHEN** Dorph Utils has a Dorph AI session token
- **AND** a session-validation request fails because the auth service is temporarily unreachable
- **THEN** Dorph Utils keeps the session
- **AND** the user is not treated as signed out

#### Scenario: Dorph AI reports invalid session
- **WHEN** Dorph Utils validates the session
- **AND** Dorph AI reports the token is invalid or expired
- **THEN** Dorph Utils treats the user as signed out
- **AND** directs them to sign in on Dorph AI
