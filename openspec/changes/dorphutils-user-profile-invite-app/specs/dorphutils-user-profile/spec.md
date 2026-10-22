## Purpose

Lets any signed-in Dorph user manage their own login credentials from the DorphUtils shell without opening DorphNotes configuration.

## ADDED Requirements

### Requirement: Account entry point in DorphUtils shell

DorphUtils SHALL expose a compact **account** control (user icon) in the shell chrome when a Dorph session is present. The control SHALL NOT appear when the user is not signed in.

#### Scenario: Signed-in user sees account icon

- **WHEN** DorphUtils loads with a valid shared Dorph session token
- **THEN** the shell shows a user/account icon affordance in the sidebar or footer area

#### Scenario: Unsigned user does not see account icon

- **WHEN** DorphUtils loads without a valid Dorph session
- **THEN** the account icon is hidden and the existing “Sign in on Dorph AI” handoff remains available

### Requirement: Self-service profile modal

Clicking the account control SHALL open an in-app modal scoped to the **current session user only**. The modal SHALL display the current username and SHALL allow updating username and password. The modal MUST NOT list or edit other users.

#### Scenario: Open profile modal

- **WHEN** a signed-in user activates the account icon
- **THEN** DorphUtils opens a modal showing the authenticated user's current username

#### Scenario: Update username

- **WHEN** the user submits a new unique username in the modal
- **THEN** DorphUtils calls the platform self-service account API
- **AND** on success the modal shows confirmation and reflects the new username

#### Scenario: Update password

- **WHEN** the user submits a new password together with the current password
- **THEN** DorphUtils calls the platform self-service account API
- **AND** on success the modal confirms the change without displaying the new password again

#### Scenario: Validation errors

- **WHEN** the platform rejects the update (duplicate username, weak password, wrong current password)
- **THEN** the modal shows the error inline and keeps the user's edits for correction
