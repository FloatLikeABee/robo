# Spec Delta

## ADDED Requirements

### Requirement: Grill write prompt loads the html-anything skill
The prompt that writes the grill draft MUST include the compact html-anything skill and Ge overlay. The draft MUST remain body text for the kind, not a raw HTML document and not a chat greeting.

#### Scenario: Write prompt includes the skill
- **WHEN** the grill builds the write prompt after the interview
- **THEN** the prompt includes the compact html-anything skill
- **AND** it still asks for the note or TODO body only
