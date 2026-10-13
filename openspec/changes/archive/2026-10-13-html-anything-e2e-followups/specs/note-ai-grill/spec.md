# Spec Delta

## ADDED Requirements

### Requirement: Question count covers each title part
The grill interview MUST plan at least one question per title part. Parts are the segments of the title split on commas, semicolons, `and`, and facet prepositions (`for`, `in`, `with`, `about`, `versus`, `vs`). A short single-intent title with no parts still plans one question. The count MUST stay at most four.

#### Scenario: Three-part offsite title
- **WHEN** the grill starts with title `Team offsite plan for 12 people in Hangzhou`
- **THEN** the planned question count is at least 3
- **AND** the count is at most 4

#### Scenario: Short single-intent title
- **WHEN** the grill starts with title `buy milk`
- **THEN** the planned question count is 1

### Requirement: Grill draft is note content not a letter
The grill write prompt MUST ask for the note or TODO body only, MUST use hex colour values with no colour word before a hex, and MUST say not to write a letter (no greeting, no sign-off, no `[Your Name]`). A model reply that is a full HTML letter or that greets with `Dear` and signs off MUST be replaced with a body built from the interview facts.

#### Scenario: Write prompt forbids letter and colour names
- **WHEN** the grill builds the write prompt
- **THEN** the prompt includes the compact html-anything skill
- **AND** it asks for the body only
- **AND** it does not contain `cyan #38bdf8`
- **AND** it forbids a letter greeting and sign-off

#### Scenario: Letter-shaped HTML is discarded
- **WHEN** the model returns a full HTML page that starts `Dear Team` and ends `[Your Name]`
- **THEN** the stored draft is not that letter
- **AND** the draft still uses the interview answers

## MODIFIED Requirements

### Requirement: Grill write prompt loads the html-anything skill
The prompt that writes the grill draft MUST include the compact html-anything skill and Ge overlay. The draft MUST remain body text for the kind, not a raw HTML document, not a chat greeting, and not a signed letter.

#### Scenario: Write prompt includes the skill
- **WHEN** the grill builds the write prompt after the interview
- **THEN** the prompt includes the compact html-anything skill
- **AND** it still asks for the note or TODO body only
