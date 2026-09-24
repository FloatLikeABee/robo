## Purpose

Simplifies creating a MorphNotes task to prompt or upload generate, a title, start and end times, and a markdown plus HTML document.

## ADDED Requirements

### Requirement: Create form is generate, title, and times

When creating a MorphNotes task, the operator MUST see: generate from prompt, generate from upload (or both), title, start date/time, and end date/time. The create form MUST NOT show description, map area, JSON detail editor, or attachments.

#### Scenario: New task dialog

- **WHEN** the operator opens Create case/task
- **THEN** they can enter a prompt and/or upload a file and click Generate
- **AND** they can set title, start, and end
- **AND** description, map, JSON detail, and attachments are not on that create form

### Requirement: Outcome is markdown and HTML

Generated and saved create outcome MUST be a markdown document and an HTML preview of that document. The AI draft MUST return markdown (and title/times), not a location polygon or JSON detail object as the document.

#### Scenario: Generate fills markdown and HTML

- **WHEN** generate succeeds
- **THEN** the operator can view Markdown and HTML for the draft
- **AND** those views are the task document, not a JSON tree

#### Scenario: Save stores the markdown document

- **WHEN** the operator saves a newly created task
- **THEN** the task keeps title, start, end, and the markdown document
- **AND** HTML is the preview of that markdown

### Requirement: Edit of existing tasks keeps stored fields

The create-form simplification MUST NOT hide extra fields when opening an already saved task. This capability applies to **create** only.

#### Scenario: Existing task still opens

- **WHEN** the operator opens a task that already has extra stored fields
- **THEN** they can still view that task
- **AND** map, JSON detail, or attachments that were stored remain available on edit
