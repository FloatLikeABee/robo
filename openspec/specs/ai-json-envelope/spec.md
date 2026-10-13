# ai-json-envelope Specification

## Purpose

Lets Morph accept JSON from small local models that put raw newlines inside strings, wrap the object in fences, or add prose around it.

## Requirements

### Requirement: Parser accepts messy model JSON
When Morph parses a model reply that is meant to be one JSON object, it MUST succeed if the object is wrapped in markdown fences, surrounded by extra text, or contains raw newlines or tabs inside string values. A well-formed JSON object MUST still parse. The same parser MUST be used for big-note generate and timeline generate.

#### Scenario: Raw newlines inside a string
- **WHEN** the model returns a JSON object whose `markdown` string contains a raw newline
- **THEN** parse succeeds
- **AND** the markdown value keeps those line breaks

#### Scenario: Fences and extra text
- **WHEN** the model returns prose, a ` ```json ` fence, then a JSON object, then more prose
- **THEN** parse succeeds and yields that object

### Requirement: Parser repairs missing commas between properties
When Morph parses a model JSON object, it MUST succeed if two properties are separated only by whitespace (the `invalid character '"' after object key:value pair` failure). Fences around that object MUST still be stripped.

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
