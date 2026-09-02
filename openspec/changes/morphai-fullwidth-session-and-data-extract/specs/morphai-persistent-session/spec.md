## Purpose

Keeps a Morph AI login active until the user explicitly signs out, including after closing the browser.

## ADDED Requirements

### Requirement: Login persists until Sign out
After a successful Morph AI login, the system SHALL keep the user signed in across browser restarts and new tabs on the same browser profile. The session MUST be cleared only when the user chooses Sign out, or when the credential is otherwise invalidated by an explicit logout/admin action.

#### Scenario: Close and reopen the browser
- **WHEN** a user signs in to Morph AI
- **AND** they close the browser and later reopen Morph AI on the same profile
- **THEN** they remain signed in
- **AND** they are not sent to the login page

#### Scenario: Explicit Sign out
- **WHEN** a signed-in user chooses Sign out
- **THEN** the session is cleared
- **AND** the next visit to Morph AI requires login

### Requirement: Login has no Remember-me opt-out
The Morph AI login form SHALL always persist the session. It MUST NOT offer a control that stores the session only until the browser closes.

#### Scenario: Sign in from the login page
- **WHEN** a user submits valid credentials on Morph AI login
- **THEN** the session is persisted on the device
- **AND** the form does not show a Remember me checkbox (or equivalent session-only option)
