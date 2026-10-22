# Spec Delta

## ADDED Requirements

### Requirement: Parser repairs missing commas between properties
When Dorph parses a model JSON object, it MUST succeed if two properties are separated only by whitespace (the `invalid character '"' after object key:value pair` failure). Fences around that object MUST still be stripped.

#### Scenario: Missing comma after a string value
- **WHEN** the model returns `{"title":"Hi" "markdown":"body"}` or the same object with a newline instead of a space
- **THEN** parse succeeds
- **AND** title is `Hi` and markdown is `body`

#### Scenario: Fenced object with a missing comma
- **WHEN** that same object is wrapped in a markdown `json` fence
- **THEN** parse succeeds

### Requirement: Big-note generate never hard-fails on parse
When big-note generate cannot parse the first model reply as JSON, it MUST retry the model once. If the retry is also not JSON, it MUST store the reply as plain markdown (fences stripped) and MUST NOT return a parse error. A transport or configuration failure MAY still fail the request.

#### Scenario: Fenced non-JSON reply
- **WHEN** the model returns a fenced markdown note with no JSON object
- **THEN** generate retries once
- **AND** if the retry is also not JSON, the stored markdown is that note body
- **AND** the caller does not receive a parse error

#### Scenario: Non-JSON prose
- **WHEN** the model returns prose with no `{...}` object
- **THEN** generate retries once, then stores that prose as markdown
- **AND** the caller does not receive a parse error

#### Scenario: Parse succeeds on the first reply
- **WHEN** the first reply is a JSON object the parser accepts
- **THEN** generate does not call the model a second time
