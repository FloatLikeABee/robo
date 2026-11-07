# Spec Delta

## ADDED Requirements

### Requirement: New DorphNotes HTML uses the bento colors
HTML newly generated for DorphNotes research, timelines, big notes, case tasks, and Contents & Data publish MUST use page `#221c34`, card `#322a4a`, and accents `#8b5cf6`, `#a78bfa`, and `#f97316`. It MUST NOT use a light page. Pages already stored MAY stay unchanged until the next generate.

#### Scenario: New research or publish page
- **WHEN** DorphNotes writes a new HTML page
- **THEN** the page background is `#221c34`
- **AND** the stylesheet includes `#322a4a`, `#8b5cf6`, `#a78bfa`, and `#f97316`

#### Scenario: Older stored page
- **WHEN** an operator opens HTML that was saved before this palette
- **THEN** that stored HTML is not rewritten until the next generate
