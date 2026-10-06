## Purpose

Makes MorphNotes, opened from Morph AI, the only home for Event Logs, Content Maker, and Project.

## ADDED Requirements

### Requirement: MorphNotes lists the three modules
MorphNotes navigation MUST include Event Logs, Content Maker, and Project in addition to its current sections. Choosing one MUST show that module inside MorphNotes. Event Logs MUST load its events page (`/events-info` on its origin). The modules MUST keep their existing user-facing names.

#### Scenario: Project opens inside MorphNotes
- **WHEN** the operator opens Project from MorphNotes
- **THEN** the Project UI is shown inside MorphNotes
- **AND** the operator is not sent to a MorphUtils site

#### Scenario: Event Logs keeps its events page
- **WHEN** the operator opens Event Logs from MorphNotes
- **THEN** the embedded page is that module's `/events-info` path

### Requirement: A module that is down shows a hint
When Event Logs, Content Maker, or Project cannot be reached at its configured origin, MorphNotes MUST show a short hint and MUST NOT mount that iframe.

#### Scenario: Project origin is down
- **WHEN** the operator opens Project and its origin refuses the connection
- **THEN** MorphNotes shows a hint instead of an empty frame

### Requirement: MorphUtils is not an entry
The Morph AI header MUST NOT include a MorphUtils link or chip. Login copy MUST NOT tell the operator to use MorphUtils. `./start-all.sh` with no service argument, and `start all`, MUST NOT start the MorphUtils UI.

#### Scenario: Header has MorphNotes and not MorphUtils
- **WHEN** the operator views the Morph AI header
- **THEN** MorphNotes is available
- **AND** MorphUtils is not listed
