# Spec Delta

## ADDED Requirements

### Requirement: Generated header chrome uses Arial
The header meta line and other generated document chrome on stories, research, timelines, big notes, and case-task HTML MUST use the Arial sans stack. Those rules MUST NOT use `system-ui`.

#### Scenario: Big-note header is Arial
- **WHEN** Morph renders a big note as HTML
- **THEN** the header meta font stack includes Arial
- **AND** the stylesheet does not set the header font to `system-ui`

### Requirement: Big-note generate stays on the idea
When Morph asks a model to write a big note from an idea, the prompt MUST tell the model to use only facts named in the idea and MUST forbid inventing features, products, agent types, or capabilities the idea does not name.

#### Scenario: Idea-grounded prompt
- **WHEN** Morph builds the big-note generate prompt
- **THEN** the prompt includes the idea text
- **AND** it forbids inventing features not in the idea
