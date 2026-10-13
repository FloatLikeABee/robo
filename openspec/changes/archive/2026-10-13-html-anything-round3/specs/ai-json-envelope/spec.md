# Spec Delta

## ADDED Requirements

### Requirement: Parser unwraps backtick-quoted values
When a model JSON object uses JS backticks around a string value instead of JSON quotes, Morph MUST unwrap that value into a JSON string and parse the object. The stored markdown MUST be the inner document, not the envelope.

#### Scenario: Backtick-wrapped markdown field
- **WHEN** the model returns `{"title":"Weekly Review Habit Guide","markdown": `## Setting Up...`}`
- **THEN** parse succeeds
- **AND** title is `Weekly Review Habit Guide`
- **AND** markdown starts with a heading, not `{`

### Requirement: Failed parse does not store the raw envelope
If generate still cannot parse after retry, it MUST NOT store a `{...}` JSON envelope as the note body. It MUST use the idea as the title when the parsed title is empty.

#### Scenario: Envelope is not saved as the body
- **WHEN** the only model replies are a JSON envelope that still cannot be parsed
- **THEN** the stored markdown is not that envelope
- **AND** the stored title is taken from the idea
