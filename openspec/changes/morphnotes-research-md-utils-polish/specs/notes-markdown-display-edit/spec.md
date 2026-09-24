## Purpose

MorphNotes document panels show Markdown as a rendered document by default, with an explicit switch into raw Markdown text when the operator needs to edit or copy the source.

## ADDED Requirements

### Requirement: Rendered Markdown is the default document view
Wherever MorphNotes shows a Markdown document body (Timelines, Big notes including analysis Markdown, Case/task Markdown tab, Generic data content Markdown, Notes & TODOs note body), the default view MUST be rendered Markdown (headings, lists, links, emphasis visible as formatted text), not a raw source dump. Chat message bubbles and HTML-source tabs are out of scope.

#### Scenario: Timeline Markdown opens rendered
- **WHEN** the operator opens a Timeline that has Markdown content
- **THEN** the Markdown document view shows rendered Markdown, not a monospace raw dump as the only view

#### Scenario: Big notes Markdown opens rendered
- **WHEN** the operator opens a Big note that has Markdown content
- **THEN** the Markdown document view shows rendered Markdown

#### Scenario: Generic data content stays rendered
- **WHEN** the operator opens Generic data content that includes Markdown
- **THEN** the Markdown is shown rendered (as it is today) and a Raw mode is available

### Requirement: Raw Markdown mode for stored bodies
For stored Markdown bodies (Timeline, Big note body, Generic data `content_markdown`, Notes & TODOs body, Research conclusion Markdown), the operator MUST be able to switch to a raw Markdown text mode, edit the source, and persist the change through the existing save path for that record. Switching back to the rendered view MUST show the current source (including unsaved edits still on screen).

#### Scenario: Edit Timeline Markdown in Raw and keep it
- **WHEN** the operator switches a Timeline to Raw, changes the Markdown source, and saves
- **THEN** the stored Timeline Markdown is the edited source
- **AND** the rendered view shows that edited document

#### Scenario: Toggle does not discard unsaved Raw edits
- **WHEN** the operator edits Raw text then switches to the rendered view without leaving the record
- **THEN** the rendered view reflects the unsaved Raw text still in the editor

### Requirement: Derived Case/task Markdown is Raw-viewable, not a second editor
The Case/task Markdown tab MUST default to rendered Markdown. Raw mode MUST show the derived source as text. Editing that Raw text MUST NOT become a second saved Case/task body; the Details form remains the source of truth, and Save continues to persist form fields only.

#### Scenario: Case/task Markdown renders then shows Raw source
- **WHEN** the operator opens Case/task Markdown for a draft with a title and description
- **THEN** they first see rendered Markdown of that draft
- **AND** switching to Raw shows the derived source text including title and description

#### Scenario: Case/task Raw does not replace Details save
- **WHEN** the operator types in Case/task Raw mode then saves from Details
- **THEN** the case is stored from the Details form fields
- **AND** no extra Markdown column is required on Case/task

### Requirement: Dark document chrome
Rendered Markdown panels MUST use the existing dark MorphNotes surface (no light-theme document chrome, no theme switch).

#### Scenario: Rendered Markdown stays dark
- **WHEN** the operator views rendered Markdown in Timelines or Big notes
- **THEN** the panel uses the dark product UI
- **AND** there is no light/dark theme toggle on that panel
