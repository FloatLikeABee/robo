## Purpose

Provides a standalone, minimal onboarding app where admins issue invitation codes and new users redeem codes to receive login credentials without admin hand-typing accounts.

## ADDED Requirements

### Requirement: Standalone invite-signup application

The platform SHALL ship a separate Invite Signup web app (not embedded in MorphUtils or MorphNotes) served on its own dev port. The app SHALL have two public flows: **Admin** (authenticated) and **Redeem** (unauthenticated).

#### Scenario: App is reachable independently

- **WHEN** a user opens the Invite Signup app URL in a browser
- **THEN** they see a simple landing with paths to redeem a code or sign in as admin
- **AND** the UI does not require opening MorphNotes or MorphUtils

### Requirement: Admin creates invitation codes

An authenticated **admin** Morph user SHALL be able to sign in on the Admin page and create one or more invitation codes. Each code SHALL be a human-shareable secret (e.g. 8–12 alphanumeric characters) shown once at creation time.

#### Scenario: Admin login

- **WHEN** a user with admin privileges signs in on the Admin page with valid Morph credentials
- **THEN** the app stores the session token and shows the code-creation UI

#### Scenario: Non-admin rejected

- **WHEN** a signed-in non-admin attempts to access the Admin code-creation UI
- **THEN** the app shows a forbidden message and does not create codes

#### Scenario: Create code

- **WHEN** an admin clicks create invitation code
- **THEN** the app requests a new code from the platform API
- **AND** displays the code prominently for copy/share

### Requirement: User redeems code for credentials

The Redeem page SHALL accept an invitation code without requiring login. On successful redemption the app SHALL display a generated **username** and **password** exactly once in plain text with copy-friendly layout.

#### Scenario: Successful redemption

- **WHEN** a visitor submits a valid unused invitation code
- **THEN** the platform creates a new `plat_users` account
- **AND** the app shows a simple unique username and simple password
- **AND** instructs the user to sign in on Morph AI with those credentials

#### Scenario: Invalid or used code

- **WHEN** a visitor submits an unknown, expired, or already-redeemed code
- **THEN** the app shows an error and does not reveal credentials

#### Scenario: Generated credential shape

- **WHEN** redemption succeeds
- **THEN** the username is lowercase alphanumeric, 6–12 characters, unique in `plat_users`
- **AND** the password is 8–10 characters from an easy-to-type charset (letters and digits, no symbols required)
