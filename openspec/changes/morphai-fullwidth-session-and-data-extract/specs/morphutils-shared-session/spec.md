## Purpose

Lets Morph Utils reuse the Morph AI login the same way Morph Data does, without a second sign-in.

## ADDED Requirements

### Requirement: Morph Utils is signed in after Morph AI login
If the user is already signed in to Morph AI, Morph Utils SHALL treat them as signed in. Opening Morph Utils from the Morph AI Apps menu MUST carry the session automatically. Morph Utils MUST NOT require a separate login form when that Morph AI session is valid.

#### Scenario: Open Morph Utils from Morph AI Apps
- **WHEN** a user is signed in to Morph AI
- **AND** they open Morph Utils from the Morph AI Apps menu
- **THEN** Morph Utils is authenticated
- **AND** they can use Utils modules without signing in again

#### Scenario: Morph Data remains same-session
- **WHEN** a user is signed in to Morph AI
- **AND** they open Morph Data
- **THEN** Morph Data continues to use the same Morph AI session with no extra login

### Requirement: Utils does not drop a valid Morph AI session
Morph Utils MUST NOT clear the shared session solely because a session-validation request fails due to a network or proxy error. It MAY clear the session only when Morph AI reports the credential is invalid or the user Signs out.

#### Scenario: Transient auth check failure
- **WHEN** Morph Utils has a Morph AI session token
- **AND** a session-validation request fails because the auth service is temporarily unreachable
- **THEN** Morph Utils keeps the session
- **AND** the user is not treated as signed out

#### Scenario: Morph AI reports invalid session
- **WHEN** Morph Utils validates the session
- **AND** Morph AI reports the token is invalid or expired
- **THEN** Morph Utils treats the user as signed out
- **AND** directs them to sign in on Morph AI
