# Spec Delta

## ADDED Requirements

### Requirement: Grill draft is a body fragment
The stored grill draft MUST be a note body fragment. It MUST NOT contain `<!DOCTYPE`, `<html`, `<head`, `<body`, or `<style`. A model reply that includes those wrappers MUST be sanitized to the inner fragment, or replaced with a body built from the interview facts when the fragment is empty.

#### Scenario: Full HTML page with style is stripped
- **WHEN** the model returns the #153 grill draft (style block, doctype, html, and body around a note)
- **THEN** the stored draft has none of those wrappers
- **AND** the draft still contains the inner note headings or paragraphs that remain after grounding checks

### Requirement: Grill draft stays on the answers
The write prompt MUST tell the model to use only the interview answers, keep the user's units and currency, and not invent dates or places. A draft that changes the stated currency (for example `$30,000` when the answers say `30k RMB`) or invents a weekday clock time or place that is not in the answers MUST be replaced with a body built from those answers.

#### Scenario: Write prompt grounds units
- **WHEN** the grill builds the write prompt
- **THEN** the prompt says to keep the user's units and currency
- **AND** it forbids inventing dates or places

#### Scenario: Dollar rewrite of RMB is discarded
- **WHEN** the answers include `30k RMB` and the model draft says `$30,000`
- **THEN** the stored draft does not contain `$30,000`
- **AND** the stored draft contains `RMB` or `30k`

### Requirement: Grill card is dark with accent borders
The grill page MUST paint cards with background `#111827` and an accent colour as the border. Question and draft cards MUST NOT fill with `#a78bfa` or use mint (`#34d399`) text on a violet fill.

#### Scenario: Card chrome
- **WHEN** the grill HTML is read
- **THEN** it sets `--card` to `#111827`
- **AND** the question or draft card uses that background and an accent border
- **AND** it does not set a violet (`#a78bfa`) card fill
