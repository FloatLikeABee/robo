## Purpose

Operators can open MorphNotes Case/task details and switch to Markdown and HTML views of that case without leaving the drawer, using the same fields they already edit.

## ADDED Requirements

### Requirement: Case/task details has Details, Markdown, and HTML tabs
The MorphNotes Case/task details surface SHALL present three tabs: **Details** (the existing edit form), **Markdown**, and **HTML**. Switching tabs MUST NOT discard unsaved form fields.

#### Scenario: Open existing case shows three tabs
- **WHEN** the operator opens Case/task details for a saved case
- **THEN** they see tabs labeled Details, Markdown, and HTML
- **AND** Details is the default tab and still contains title, description, JSON detail, and save/cancel

#### Scenario: Create case also has the tabs
- **WHEN** the operator opens Create case/task
- **THEN** the same three tabs are available
- **AND** Markdown and HTML reflect the current draft (including empty fields)

#### Scenario: Switching tabs keeps the draft
- **WHEN** the operator edits the title on Details then switches to Markdown and back
- **THEN** the title field still shows the edited value

### Requirement: Markdown tab shows a case document
The Markdown tab SHALL show a Markdown document derived from the current draft: title, description, assignment/times/location when present, and the JSON detail as a fenced `json` block. The document MUST update when Details fields change (without requiring Save). The tab MUST show source Markdown (copyable), not an empty placeholder when the case has a title.

#### Scenario: Description and JSON appear in Markdown
- **WHEN** the draft has title "Route delay", a non-empty description, and JSON detail
- **THEN** the Markdown tab includes the title, the description text, and a fenced JSON block containing that detail

#### Scenario: Unsaved edit is visible in Markdown
- **WHEN** the operator changes the description on Details without saving
- **THEN** the Markdown tab shows the new description

### Requirement: HTML tab shows a rendered preview
The HTML tab SHALL show a rendered HTML preview of the same document as the Markdown tab (Markdown converted to HTML). The preview MUST use a dark page suitable for MorphNotes (not a light-theme document). Empty optional fields MUST NOT break the preview.

#### Scenario: HTML preview renders the title
- **WHEN** the draft title is non-empty and the operator opens the HTML tab
- **THEN** they see a rendered page that includes that title as visible text

#### Scenario: HTML follows Markdown
- **WHEN** the operator changes the description on Details
- **THEN** the HTML tab preview includes the updated description after they open that tab

### Requirement: Views are derived, not a second stored body
The system MUST NOT add persisted `markdown_content` / `html_content` columns on CaseTask for this change. Saving the case MUST continue to persist the existing fields only (title, description, times, location, JSON detail, assignees, attachments).

#### Scenario: Save still writes the form fields
- **WHEN** the operator saves from Details after viewing Markdown and HTML
- **THEN** the case is stored with the form fields they edited
- **AND** no extra markdown/html columns are required on the CaseTask table
