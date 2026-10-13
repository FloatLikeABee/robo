# Spec Delta

## ADDED Requirements

### Requirement: Big notes have a minimum depth
A generated big note MUST have at least three `##` sections with body text under them. The generate prompt MUST ask for that depth. A one-paragraph reply MUST be expanded from the idea (retry once, then a local expand) so the stored markdown is not a single paragraph.

#### Scenario: Prompt asks for sections
- **WHEN** Morph builds the big-note generate prompt
- **THEN** the prompt requires at least three `##` sections

#### Scenario: One paragraph is expanded
- **WHEN** the model returns a single paragraph about the idea
- **THEN** the stored markdown has at least three `##` headings
- **AND** the headings or body still use words from the idea
