## Purpose

Lets operators author Morph AI skills in a wide dark split editor, including markdown file upload and an AI pass that drafts name, description, and instructions to save.

## ADDED Requirements

### Requirement: Skills modal is a wide dark split editor
The Morph AI Skills dialog MUST be dark (no light-theme skin). It MUST be wider than the current ~720px panel. The upload form MUST sit on the left and the catalog on the right. When the catalog exceeds the pane height, it MUST scroll vertically without growing the whole dialog unbounded. The instructions field MUST be a tall textarea (longer than five short rows).

#### Scenario: Layout is side by side and dark
- **WHEN** the operator opens Skills from the Morph AI header
- **THEN** they see a dark dialog with upload fields on the left and the catalog on the right
- **AND** a long catalog scrolls inside the right pane

### Requirement: Operators can upload a markdown skill file
The upload pane MUST accept a `.md` file. Choosing a file MUST fill Name (from heading or filename) and Instructions (from the markdown body) so the operator can save. Saving MUST still create a skill via the existing skills API.

#### Scenario: Markdown file populates the form
- **WHEN** the operator selects a `.md` file in Skills
- **THEN** Name and Instructions are populated from that file
- **AND** they can save the skill without retyping the body

### Requirement: Improve with AI drafts the skill fields
The upload pane MUST include a control to improve the draft with AI. That action MUST send the current Name, Description, and Instructions to Morph AI and replace those fields with a revised draft. It MUST NOT save until the operator uploads/saves. Empty required fields MUST not call AI successfully (show an error instead).

#### Scenario: AI fills a draft the operator can save
- **WHEN** the operator has Name and Instructions and clicks Improve with AI
- **THEN** Name, Description, and Instructions update to the AI draft
- **AND** the skill is not persisted until they save/upload
