# Spec Delta

## MODIFIED Requirements

### Requirement: Every HTML-producing path loads the skill
Note AI output, generated HTML or markdown, and agent-facing skill docs MUST include the compact html-anything skill and Ge overlay. Grill write is plain markdown and MUST NOT receive that skill. Ordinary chat that is not producing HTML MUST NOT receive the full vendor SKILL.md.

#### Scenario: Note AI and generated documents
- **WHEN** Morph builds a note-AI, generated HTML/MD, or form-HTML prompt
- **THEN** that prompt includes the compact skill text
- **AND** it includes Ge's Arial, dark, multi-accent overlay

#### Scenario: Grill final content
- **WHEN** the grill builds the write prompt for the draft
- **THEN** the prompt does not include the compact skill text
- **AND** it asks for plain structured markdown with no colour codes

#### Scenario: Ordinary chat stays compact
- **WHEN** Morph builds a normal chat reply prompt that is not asking for HTML
- **THEN** it does not append the full vendor SKILL.md
