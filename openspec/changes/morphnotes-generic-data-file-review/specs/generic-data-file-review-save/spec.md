## Purpose

Lets MorphNotes operators import a Generic data file by extracting JSON and Markdown, editing those drafts, saving only when they confirm, and running AI analysis on the saved record.

## ADDED Requirements

### Requirement: File extract yields JSON and Markdown drafts
When the user chooses a supported Generic data file and runs extract, the system MUST return both a JSON draft and a Markdown draft derived from that file. No Generic data record MUST be created by extract alone.

#### Scenario: Extract from a file
- **WHEN** the user selects a supported file (CSV, Excel, JSON, PDF, or Markdown) and runs extract
- **THEN** the import UI shows editable JSON and Markdown
- **AND** no Generic data list row is created yet

#### Scenario: Extract with parse seed when AI is unavailable
- **WHEN** the user runs extract on a parseable file (CSV, Excel, JSON, Markdown, or PDF with a text layer)
- **AND** Morph AI is not configured or the AI call fails
- **THEN** the import UI still shows parse-based JSON and Markdown drafts
- **AND** no record is created until Save

#### Scenario: Extract with nothing to draft
- **WHEN** the user runs extract
- **AND** the file has no parseable content (for example a scanned PDF with no text layer)
- **AND** Morph AI is not configured or the AI call fails
- **THEN** the UI shows that extract is unavailable
- **AND** no record is created

### Requirement: Save after edit
The user MUST be able to edit the JSON and Markdown drafts and persist them with a Save action. Cancel or close without Save MUST NOT create a record. Invalid JSON on Save MUST be rejected without writing a row.

#### Scenario: Save reviewed drafts
- **WHEN** the user has extracted drafts, optionally edited them, and clicks Save
- **THEN** a Generic data record is stored
- **AND** opening it shows the saved JSON and Markdown content

#### Scenario: Cancel after extract
- **WHEN** the user extracts drafts and closes without Save
- **THEN** no Generic data record is created

#### Scenario: Invalid JSON on save
- **WHEN** the JSON draft is not valid JSON
- **AND** the user clicks Save
- **THEN** the save is rejected with a validation error
- **AND** no record is created

### Requirement: AI analysis on saved records
After a Generic data record exists, the user MUST be able to run Morph AI analysis on its stored content. Extract and Save MUST NOT run analysis automatically.

#### Scenario: Run analysis after save
- **WHEN** the user has saved a Generic data record
- **AND** they choose Run AI analysis
- **THEN** the record stores an analysis they can read
- **AND** they can run it again later

#### Scenario: Analysis unavailable
- **WHEN** the user runs AI analysis
- **AND** Morph AI is not configured
- **THEN** the UI shows that analysis is unavailable
- **AND** the saved content is unchanged
