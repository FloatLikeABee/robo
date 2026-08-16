## Purpose

Provides Docs AI inside DataX with a two-pane session layout and content-file uploads that stay separate from DataX tabular data imports.

## ADDED Requirements

### Requirement: Two-pane Docs AI layout
Docs AI SHALL present chat session history in a left pane and the active conversation in a right pane.

#### Scenario: Switch sessions from the left list
- **WHEN** the user selects a prior session in the left history list
- **THEN** the right chat pane loads that session’s messages without leaving the Docs AI view

### Requirement: Create and continue sessions
The system SHALL allow users to start a new Docs AI session and continue an existing session from the left history list.

#### Scenario: New session
- **WHEN** the user starts a new Docs AI session
- **THEN** a new session appears in the left history and the right pane is empty or shows the welcome state for that session

### Requirement: Content uploads distinct from data imports
Docs AI and Docs library uploads SHALL accept content document types (PDF and TXT at minimum). DataX data-table / data import flows SHALL continue to accept tabular data formats (CSV and JSON) and SHALL NOT be the path for Docs PDF/TXT content uploads.

#### Scenario: Reject tabular-only path for Docs content
- **WHEN** a user attaches a PDF in Docs AI
- **THEN** the attachment is handled by the Docs content pipeline, not the DataX CSV/JSON data-table import pipeline

#### Scenario: Data import remains tabular
- **WHEN** a user imports data through DataX data tables
- **THEN** the accepted formats remain tabular (CSV/JSON and existing data-import types), not Docs PDF/TXT content publishing
