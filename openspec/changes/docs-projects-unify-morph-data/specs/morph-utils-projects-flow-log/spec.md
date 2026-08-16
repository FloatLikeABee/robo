## Purpose

Defines Projects (formerly Engi) as the home for Flow log after Booki is removed, including free-text flow categories and Projects branding.

## ADDED Requirements

### Requirement: Projects branding replaces Engi
Morph Utils and the Projects app chrome SHALL label the former Engi module as **Projects**. User-facing “Engi” as the primary product name MUST NOT remain.

#### Scenario: Utils shows Projects
- **WHEN** a signed-in user views Morph Utils app navigation
- **THEN** the projects/engi module is labeled Projects

### Requirement: Booki module removed from Morph Utils
Morph Utils MUST NOT list Booki as an app module. Booki AI, Bookings, and other Booki-only primary surfaces MUST NOT remain reachable via Morph Utils navigation.

#### Scenario: No Booki in Utils nav
- **WHEN** a signed-in user views Morph Utils app navigation
- **THEN** Booki is absent from the module list

### Requirement: Flow log lives in Projects
Projects SHALL provide Flow log as a primary section (header tab or equivalent). Users MUST be able to create, list, filter, and delete flow-log entries from Projects.

#### Scenario: Open Flow log in Projects
- **WHEN** a signed-in user opens Projects
- **THEN** Flow log is available as a primary navigation target
- **AND** the user can record income/expense entries with date, amount, title, and category

### Requirement: Flow log category is free text
Flow log category SHALL be entered via a free-text input. The system MUST allow any non-empty user-typed category string and MUST NOT require selecting only from a fixed closed list.

#### Scenario: Custom category entry
- **WHEN** a user creates a flow-log entry and types a category that is not in any suggestion list
- **THEN** the entry saves with that category value
