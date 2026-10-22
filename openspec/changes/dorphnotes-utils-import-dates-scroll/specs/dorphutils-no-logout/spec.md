## Purpose

Removes logout, sign-out, and Out from DorphUtils so operators stay on Dorph JWT without a second sign-out control in the shell or its modules.

## ADDED Requirements

### Requirement: DorphUtils shell has no sign-out control
The DorphUtils chrome MUST NOT show **Sign out**, **Logout**, or **Out**. It MUST NOT clear the shared Dorph session from a header or footer button.

#### Scenario: DorphUtils header
- **WHEN** an operator is signed in and opens DorphUtils
- **THEN** they do not see Sign out, Logout, or Out in the DorphUtils shell

### Requirement: DorphUtils modules have no sign-out control
Event Logs, Content Maker, Data Access, and Project, when used inside DorphUtils, MUST NOT show **Sign out**, **Logout**, or **Out**. Dorph AI Sign out on the Dorph home chat MUST remain.

#### Scenario: Event Logs inside DorphUtils
- **WHEN** an operator opens Event Logs in DorphUtils
- **THEN** they do not see Logout, Sign out, or Out

#### Scenario: Content Maker inside DorphUtils
- **WHEN** an operator opens Content Maker in DorphUtils
- **THEN** they do not see Out, Sign out, or Logout

#### Scenario: Data Access inside DorphUtils
- **WHEN** an operator opens Data Access in DorphUtils
- **THEN** they do not see Sign out or Logout

#### Scenario: Project inside DorphUtils
- **WHEN** an operator opens Project in DorphUtils
- **THEN** they do not see Sign out or Logout

#### Scenario: Dorph AI still can sign out
- **WHEN** an operator is on Dorph AI chat
- **THEN** Dorph AI Sign out remains available
