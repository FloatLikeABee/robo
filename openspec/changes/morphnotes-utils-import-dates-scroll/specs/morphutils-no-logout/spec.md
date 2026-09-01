## Purpose

Removes logout, sign-out, and Out from MorphUtils so operators stay on Morph JWT without a second sign-out control in the shell or its modules.

## ADDED Requirements

### Requirement: MorphUtils shell has no sign-out control
The MorphUtils chrome MUST NOT show **Sign out**, **Logout**, or **Out**. It MUST NOT clear the shared Morph session from a header or footer button.

#### Scenario: MorphUtils header
- **WHEN** an operator is signed in and opens MorphUtils
- **THEN** they do not see Sign out, Logout, or Out in the MorphUtils shell

### Requirement: MorphUtils modules have no sign-out control
Event Logs, Content Maker, Data Access, and Project, when used inside MorphUtils, MUST NOT show **Sign out**, **Logout**, or **Out**. Morph AI Sign out on the Morph home chat MUST remain.

#### Scenario: Event Logs inside MorphUtils
- **WHEN** an operator opens Event Logs in MorphUtils
- **THEN** they do not see Logout, Sign out, or Out

#### Scenario: Content Maker inside MorphUtils
- **WHEN** an operator opens Content Maker in MorphUtils
- **THEN** they do not see Out, Sign out, or Logout

#### Scenario: Data Access inside MorphUtils
- **WHEN** an operator opens Data Access in MorphUtils
- **THEN** they do not see Sign out or Logout

#### Scenario: Project inside MorphUtils
- **WHEN** an operator opens Project in MorphUtils
- **THEN** they do not see Sign out or Logout

#### Scenario: Morph AI still can sign out
- **WHEN** an operator is on Morph AI chat
- **THEN** Morph AI Sign out remains available
