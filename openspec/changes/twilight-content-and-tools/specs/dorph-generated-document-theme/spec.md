# Spec Delta

## ADDED Requirements

### Requirement: Generated pages use the twilight accent set
New HTML pages Dorph or Content generate MUST use light text on a dark purple background and MUST include the accents `#8b5cf6`, `#a78bfa`, and `#f97316`. They MUST NOT use `#38bdf8`, `#34d399`, or `#070b14` as the page palette, and MUST NOT use a light page background. Already stored pages MAY stay as saved until the next generate or publish.

#### Scenario: New article or dashboard page
- **WHEN** a model or builder writes a new HTML page
- **THEN** the page background is dark purple
- **AND** the stylesheet includes `#8b5cf6`, `#a78bfa`, and `#f97316`
- **AND** it does not use a light page background

#### Scenario: Stored page is left alone
- **WHEN** an operator opens a page that was published before this palette
- **THEN** that stored HTML is not rewritten until the next generate or publish
