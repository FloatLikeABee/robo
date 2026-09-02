## Purpose

Lets Morph Data Tasks fill a case/task draft from a user prompt, an uploaded file, or both, so title, description, optional dates, optional map location, and especially detail JSON come from the source material instead of being typed from scratch.

## ADDED Requirements

### Requirement: Prompt, file, or both as source material
The Tasks create/edit flow SHALL let the user provide a prompt, upload a supported file, or provide both together. The system MUST treat the combined material as the source for one AI draft. The system MUST reject a generate request that has neither a prompt nor a file.

#### Scenario: Prompt only
- **WHEN** the user enters a prompt and generates a draft with no file
- **THEN** the system produces a case/task draft from that prompt

#### Scenario: File only
- **WHEN** the user uploads a supported file and generates a draft with no prompt
- **THEN** the system extracts text from the file and produces a case/task draft from that text

#### Scenario: Prompt and file together
- **WHEN** the user enters a prompt and uploads a supported file and generates a draft
- **THEN** the system uses both the prompt and the extracted file text as source material for one draft

#### Scenario: Neither prompt nor file
- **WHEN** the user generates a draft with an empty prompt and no file
- **THEN** the system does not call AI
- **AND** it tells the user to provide a prompt, a file, or both

### Requirement: Supported files and extraction errors
The system SHALL accept text-bearing files of type `.txt`, `.md`, `.markdown`, `.pdf`, `.csv`, and `.xlsx`. Image-only / scanned PDFs with no extractable text MUST fail with a clear error. Unsupported types MUST fail with an unsupported-type error. Generate MUST NOT create or update a case/task record.

#### Scenario: Import a text PDF
- **WHEN** the user uploads a PDF that has a text layer and generates a draft
- **THEN** the system uses extracted document text as source material

#### Scenario: Scanned PDF with no text
- **WHEN** the user uploads a PDF with no extractable text layer
- **THEN** generate fails with an error that the PDF has no extractable text
- **AND** no case/task record is created or updated

#### Scenario: Unsupported file type
- **WHEN** the user uploads a file type that is not in the supported list
- **THEN** generate fails with an unsupported-type error
- **AND** no case/task record is created or updated

### Requirement: Draft fills task fields from the material
A successful generate MUST return a draft the UI applies to the open form: a non-empty **title**, a **description** (empty only if the source has no extra prose), **start** and **end** date/times when the source implies them (otherwise left unset), **map location** when the source implies a place, and a **detail JSON object** derived from the material. Detail JSON is required and MUST be a JSON object (not an array or primitive). Dates and location MUST be omitted rather than invented when the source does not support them.

#### Scenario: Fill core fields and detail JSON
- **WHEN** generate succeeds on material that describes a piece of work
- **THEN** the form title is filled
- **AND** the description is filled when the source has descriptive prose
- **AND** the Details JSON is a non-empty object inferred from the material

#### Scenario: Dates only when implied
- **WHEN** the source includes a start and/or end time
- **THEN** those datetime fields are filled on the draft
- **WHEN** the source has no schedule
- **THEN** start and end remain unset

#### Scenario: Map area only when a place is implied
- **WHEN** the source names a place or includes coordinates
- **THEN** the draft location includes a label and, if coordinates appear in the source, an area point list the map can show
- **WHEN** the source has no place information
- **THEN** map area remains unset
- **AND** the system MUST NOT invent a polygon for a named place that has no coordinates in the source

### Requirement: Review before save
Generating a draft MUST NOT create or update a case/task. The user MUST still save through the existing create or update action after reviewing and optionally editing the filled fields. Manual create without generate MUST still work.

#### Scenario: Generate does not persist
- **WHEN** generate succeeds
- **THEN** the open form shows the proposed fields
- **AND** no new case/task exists until the user saves
- **AND** an existing case/task is not updated until the user saves

#### Scenario: User cancels after generate
- **WHEN** the user generates a draft and closes the drawer without saving
- **THEN** no case/task is created or updated from that draft

#### Scenario: Manual create still works
- **WHEN** the user opens Create case/task, types fields without generating, and saves
- **THEN** the case/task is created the same way as today

#### Scenario: Confirm before replacing a filled form
- **WHEN** the form already has a title, description, or detail JSON
- **AND** the user generates a new draft
- **THEN** the UI asks for confirmation before replacing those fields

### Requirement: AI unavailable or invalid output
When the AI service is not configured, generate MUST fail with a clear unavailable error and MUST NOT persist a record. When the model does not return a usable title plus JSON object, generate MUST fail with an extract error and MUST NOT persist a record.

#### Scenario: AI not configured
- **WHEN** the user generates a draft
- **AND** the AI service is not configured
- **THEN** the flow shows that AI generate is unavailable
- **AND** no record is created or updated

#### Scenario: Model output is not a valid draft
- **WHEN** generate runs
- **AND** the model response cannot be parsed into a title and a JSON object for details
- **THEN** the flow shows an extract error
- **AND** the existing form fields are left unchanged
- **AND** no record is created or updated
