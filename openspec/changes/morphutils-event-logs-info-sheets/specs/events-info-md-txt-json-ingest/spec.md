## Purpose

Lets an Event Logs operator upload markdown, plain text, or JSON so AI can turn that file into Events & Info draft records for review before anything is saved.

## ADDED Requirements

### Requirement: Upload md, txt, or json as source
Events & Info SHALL accept a file upload of type `.md` (including `.markdown`), `.txt`, or `.json` as source material for a new entry. The system MUST extract readable text from the file (pretty-printed JSON when the file is JSON) and send it to AI to propose Events & Info drafts. Generate MUST NOT persist records until the operator saves selected drafts.

#### Scenario: Markdown file
- **WHEN** the operator uploads a `.md` file and runs ingest
- **THEN** AI returns one or more Events & Info drafts from that file
- **AND** no Events & Info row is stored until the operator saves

#### Scenario: Text file
- **WHEN** the operator uploads a `.txt` file and runs ingest
- **THEN** AI returns draft event records from that text
- **AND** nothing is stored until the operator saves

#### Scenario: JSON file
- **WHEN** the operator uploads a `.json` file and runs ingest
- **THEN** the JSON content is used as source material
- **AND** AI returns draft event records
- **AND** nothing is stored until the operator saves

#### Scenario: Unsupported type
- **WHEN** the operator uploads a file that is not md, txt, json, or another type Events & Info already accepts
- **THEN** ingest fails with an unsupported-type error
- **AND** no record is created

### Requirement: Review before save
The operator MUST be able to review, edit, and choose which AI drafts to save. Canceling ingest MUST leave Events & Info unchanged.

#### Scenario: Cancel after drafts
- **WHEN** ingest returns drafts
- **AND** the operator closes without saving
- **THEN** no new Events & Info records exist from that ingest

#### Scenario: Save selected drafts
- **WHEN** ingest returns drafts
- **AND** the operator saves one or more selected drafts
- **THEN** those drafts appear as Events & Info records

### Requirement: AI unavailable
When AI is not configured, file ingest MUST fail with a clear unavailable error and MUST NOT create Events & Info records.

#### Scenario: Ingest without AI
- **WHEN** the operator uploads a supported file and runs ingest
- **AND** AI is not configured
- **THEN** the flow shows that AI ingest is unavailable
- **AND** no record is created
