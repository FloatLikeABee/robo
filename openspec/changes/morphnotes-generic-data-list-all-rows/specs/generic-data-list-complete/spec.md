## Purpose

Ensures MorphNotes Generic data list and grid show every saved record, including a file just imported and saved, instead of skipping rows so older records appear in their place.

## ADDED Requirements

### Requirement: List returns every stored record
The Generic data list MUST include every stored record (up to the existing list cap). It MUST NOT omit alternate rows. Newest updates MUST appear first.

#### Scenario: Multiple saved records all listed
- **WHEN** four Generic data records exist
- **THEN** the list contains four entries
- **AND** each record's id is present

#### Scenario: Newest saved import is listed
- **WHEN** the user saves a new Generic data import
- **THEN** that record's id is in the list
- **AND** it appears before older records that were not updated

### Requirement: Grid matches the saved record
After Save, the Generic data grid MUST show the saved row. Opening the grid row MUST match the record that was saved (not a different older row in its place).

#### Scenario: Save then see the new row
- **WHEN** the user extracts a file, saves, and the import dialog closes
- **THEN** the grid includes a row for that saved title or filename
- **AND** opening that row shows the saved JSON and Markdown content
