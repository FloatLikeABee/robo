## Purpose

Defines Events & Info multi-source AI ingest: upload files, fetch a URL, and/or paste content so AI extracts operational events and info into draft records.

## ADDED Requirements

### Requirement: Primary CTA is ingest not Share collection
The Events & Info list surface SHALL present an ingest action (upload / import from sources) as the primary alternative to manual add for bulk or document-based entry. The control formerly labeled **Share collection** MUST NOT remain the primary CTA for this job under that label.

#### Scenario: Open Events & Info
- **WHEN** a user opens Events & Info
- **THEN** they can start an ingest flow for files, URL, and/or pasted content
- **AND** they are not steered primarily by a **Share collection** button for that ingest purpose

### Requirement: At least one source required
Ingest SHALL accept any combination of file upload (`.txt`, `.md`, `.pdf`), URL, and pasted text. The user MUST supply at least one non-empty source before AI extraction runs. Multiple sources MAY be combined in one ingest request.

#### Scenario: Submit with no sources
- **WHEN** a user tries to run AI extract with no file, no URL, and no paste content
- **THEN** the system rejects the request with a clear validation error

#### Scenario: File only
- **WHEN** a user uploads a supported `.txt`, `.md`, or `.pdf` and runs extract
- **THEN** the system processes that file as the content source

#### Scenario: URL and paste together
- **WHEN** a user provides both a URL and pasted text (with or without a file)
- **THEN** the system combines those sources for extraction
- **AND** extraction proceeds because at least one source is present

### Requirement: AI extracts events and info drafts
The system SHALL use AI to extract one or more Events & Info draft records from the provided content. Drafts MUST NOT be persisted as final records until the user confirms create (or an equivalent explicit save step).

#### Scenario: Extraction yields drafts
- **WHEN** AI successfully extracts events/info from the sources
- **THEN** the user sees draft record(s) for review
- **AND** no Events & Info rows are created until the user confirms

#### Scenario: AI unavailable
- **WHEN** AI is not configured or the provider fails
- **THEN** the user receives a clear error and no records are created

### Requirement: Unsupported file types rejected
Uploads other than `.txt`, `.md`, and `.pdf` MUST be rejected with an explicit unsupported-type error for this ingest flow.

#### Scenario: Unsupported extension
- **WHEN** a user uploads a file that is not txt, md, or pdf
- **THEN** the system rejects that file before AI extraction
