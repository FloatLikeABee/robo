# Spec Delta

## MODIFIED Requirements

### Requirement: Grill write prompt loads the html-anything skill
The grill write prompt MUST ask for plain structured markdown (headings and bullets) and MUST NOT include hex colour codes, card-fill instructions, or the compact html-anything skill. The draft MUST remain body text for the kind, not a raw HTML document, not a chat greeting, and not a signed letter. `#111827` card styling is applied by the grill page at render time.

#### Scenario: Write prompt includes the skill
- **WHEN** the grill builds the write prompt after the interview
- **THEN** the prompt asks for the note or TODO body only
- **AND** it does not contain `html-anything` or hex colour codes

### Requirement: Grill draft is note content not a letter
The grill write prompt MUST ask for the note or TODO body only and MUST say not to write a letter (no greeting, no sign-off, no `[Your Name]`). It MUST NOT include hex colour codes. A model reply that is a full HTML letter or that greets with `Dear` and signs off MUST be replaced with a body built from the interview facts.

#### Scenario: Write prompt forbids letter and colour names
- **WHEN** the grill builds the write prompt
- **THEN** it asks for the body only
- **AND** it does not contain `cyan #38bdf8` or other hex colour codes
- **AND** it forbids a letter greeting and sign-off

#### Scenario: Letter-shaped HTML is discarded
- **WHEN** the model returns a full HTML page that starts `Dear Team` and ends `[Your Name]`
- **THEN** the stored draft is not that letter
- **AND** the draft still uses the interview answers

## ADDED Requirements

### Requirement: Grill write is colour-free structured markdown
The grill write prompt MUST ask for plain structured markdown (headings and bullets). It MUST NOT include hex colour codes, card-fill instructions, or the compact html-anything skill. `#111827` card styling MUST be applied by the grill page CSS at render time.

#### Scenario: Write prompt has no hex
- **WHEN** the grill builds the write prompt
- **THEN** the prompt asks for headings or bullets
- **AND** it does not contain `#38bdf8`, `#a78bfa`, `#111827`, or `html-anything`

#### Scenario: Hex table draft is stripped
- **WHEN** the model returns a markdown table whose first column is hex codes
- **THEN** the stored draft has no `#111827` or `#E8E8EF` cells
- **AND** it still contains the answers (for example `30k RMB`)
