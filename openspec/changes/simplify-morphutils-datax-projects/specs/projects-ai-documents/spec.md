## Purpose

Turn Morph Engi Projects into an AI document workflow that organizes requirements and specs from files or paste into publishable markdown and HTML.

## ADDED Requirements

### Requirement: Create project from file and/or paste with at least one source
The Projects create/extract flow SHALL accept uploaded file(s) and/or pasted text as sources. The user MAY provide more than one source type together. If neither file nor paste is provided, the system MUST reject the attempt with a clear warning or validation error and MUST NOT call AI generation or create a project document.

#### Scenario: File-only succeeds
- **WHEN** the user uploads at least one supported content file and submits without paste
- **THEN** the system proceeds with AI organization into a project document

#### Scenario: Paste-only succeeds
- **WHEN** the user pastes non-empty content and submits without a file
- **THEN** the system proceeds with AI organization into a project document

#### Scenario: File and paste together succeed
- **WHEN** the user provides both a file and paste and submits
- **THEN** the system uses both sources for generation

#### Scenario: No source warns
- **WHEN** the user submits with no file and empty paste
- **THEN** the UI shows a warning and the API returns a validation error without creating a project document

### Requirement: AI produces organized project markdown and HTML
Given accepted source text (from files and/or paste), the system SHALL use AI to produce an organized project content document as markdown, SHALL persist that markdown, and SHALL produce HTML suitable for in-app preview and publishing. Source material is expected to be requirements, specifications, or descriptive project content; the output MUST be structured project documentation rather than unstructured raw paste.

#### Scenario: Generation stores both formats
- **WHEN** generation completes successfully
- **THEN** the project record stores non-empty markdown and corresponding HTML

#### Scenario: User can preview both formats
- **WHEN** the user opens a generated project
- **THEN** they can view markdown and an HTML preview

### Requirement: User can publish project HTML
The system SHALL allow the project owner to publish the HTML to a public URL so unauthenticated readers can open it.

#### Scenario: Publish allocates public access
- **WHEN** the owner publishes a project that has HTML content
- **THEN** the system returns a public URL or path for the published HTML

#### Scenario: Public page serves HTML
- **WHEN** an unauthenticated client requests a published project page
- **THEN** the system returns HTML with content type `text/html`

### Requirement: Paste during extract is saved into Files
When pasted content is used as a generation source, the system SHALL also persist that paste as a content file (or equivalent Files library entry) so it appears in the Files list.

#### Scenario: Paste appears in Files after extract
- **WHEN** the user creates a project using pasted content (alone or with files)
- **THEN** a new Files entry exists representing that pasted content

#### Scenario: File uploads appear in Files
- **WHEN** the user creates a project using uploaded files
- **THEN** those uploads are retained in the Files list (or linked so the user can find them under Files)
