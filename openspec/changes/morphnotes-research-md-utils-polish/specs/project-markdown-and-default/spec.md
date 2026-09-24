## Purpose

MorphUtils Project opens on the first project in the menu and shows project Markdown as a rendered document, with a switch to raw Markdown when the operator needs to edit.

## ADDED Requirements

### Requirement: First project is selected on enter
When the operator opens the Project module and the project list has at least one item, the system MUST select the first item in the current list order and show its document. An empty list MUST keep the empty state.

#### Scenario: Non-empty project list auto-selects first
- **WHEN** the operator opens Project and at least one project exists
- **THEN** the first list item is selected
- **AND** its Markdown/HTML document pane is shown without an extra click

#### Scenario: Empty project list stays empty
- **WHEN** the operator opens Project and the list is empty
- **THEN** no project is selected
- **AND** the empty-state prompt to select or create remains

#### Scenario: Delete selected with siblings left
- **WHEN** the operator deletes the selected project and at least one project remains
- **THEN** the first remaining list item is selected

### Requirement: Markdown displays rendered with Raw edit
The Project Markdown view MUST default to rendered Markdown (formatted headings, lists, and emphasis), not a raw source dump. The operator MUST be able to switch to raw Markdown text, edit it, and save. Saving MUST persist the Markdown and keep the HTML document in sync with that Markdown.

#### Scenario: Markdown tab is rendered
- **WHEN** the operator views a project that has Markdown content
- **THEN** the Markdown view shows rendered Markdown, not only a monospace raw dump

#### Scenario: Raw edit saves
- **WHEN** the operator switches to Raw, edits the Markdown, and saves
- **THEN** the stored project Markdown is the edited source
- **AND** the rendered Markdown view shows the edited document
- **AND** the HTML view matches the updated Markdown

#### Scenario: HTML tab still available
- **WHEN** the operator is viewing a project document
- **THEN** they can still switch to the HTML preview
- **AND** publishing HTML continues to work
