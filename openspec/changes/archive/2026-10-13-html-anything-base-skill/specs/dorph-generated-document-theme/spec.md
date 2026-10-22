# Spec Delta

## ADDED Requirements

### Requirement: Generation prompts load the html-anything skill
When Dorph asks a model to write HTML or markdown that becomes a generated document, the prompt MUST include the compact html-anything skill and Ge's Arial dark multi-accent overlay. It MUST NOT request a light page or a single-accent palette.

#### Scenario: Form HTML prompt
- **WHEN** Dorph builds the form HTML generation prompt
- **THEN** the prompt includes the compact html-anything skill
- **AND** it names a dark theme, a sans-serif font, and more than one accent color

#### Scenario: Generated markdown prompt
- **WHEN** Dorph builds a timeline, research, or story generation prompt
- **THEN** the prompt includes the compact html-anything skill
- **AND** it still asks for markdown that the server can render, not raw HTML unless the path is form HTML
