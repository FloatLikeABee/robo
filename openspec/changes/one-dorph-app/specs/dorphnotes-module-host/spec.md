## Purpose

Makes DorphNotes, opened from Dorph AI, the only home for Event Logs, Content Maker, and Project.

## ADDED Requirements

### Requirement: DorphNotes lists the three modules
DorphNotes navigation MUST include Event Logs, Content Maker, and Project in addition to its current sections. Choosing one MUST show that module inside DorphNotes. Event Logs MUST load its events page (`/events-info` on its origin). The modules MUST keep their existing user-facing names.

#### Scenario: Project opens inside DorphNotes
- **WHEN** the operator opens Project from DorphNotes
- **THEN** the Project UI is shown inside DorphNotes
- **AND** the operator is not sent to a DorphUtils site

#### Scenario: Event Logs keeps its events page
- **WHEN** the operator opens Event Logs from DorphNotes
- **THEN** the embedded page is that module's `/events-info` path

### Requirement: A module that is down shows a hint
When Event Logs, Content Maker, or Project cannot be reached at its configured origin, DorphNotes MUST show a short hint and MUST NOT mount that iframe.

#### Scenario: Project origin is down
- **WHEN** the operator opens Project and its origin refuses the connection
- **THEN** DorphNotes shows a hint instead of an empty frame

### Requirement: DorphUtils is not an entry
The Dorph AI header MUST NOT include a DorphUtils link or chip. Login copy MUST NOT tell the operator to use DorphUtils. `./start-all.sh` with no service argument, and `start all`, MUST NOT start the DorphUtils UI.

#### Scenario: Header has DorphNotes and not DorphUtils
- **WHEN** the operator views the Dorph AI header
- **THEN** DorphNotes is available
- **AND** DorphUtils is not listed
