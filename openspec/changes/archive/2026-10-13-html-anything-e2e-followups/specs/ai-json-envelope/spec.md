# Spec Delta

## Purpose

Lets Dorph accept JSON from small local models that put raw newlines inside strings, wrap the object in fences, or add prose around it.

## ADDED Requirements

### Requirement: Parser accepts messy model JSON
When Dorph parses a model reply that is meant to be one JSON object, it MUST succeed if the object is wrapped in markdown fences, surrounded by extra text, or contains raw newlines or tabs inside string values. A well-formed JSON object MUST still parse. The same parser MUST be used for big-note generate and timeline generate.

#### Scenario: Raw newlines inside a string
- **WHEN** the model returns a JSON object whose `markdown` string contains a raw newline
- **THEN** parse succeeds
- **AND** the markdown value keeps those line breaks

#### Scenario: Fences and extra text
- **WHEN** the model returns prose, a ` ```json ` fence, then a JSON object, then more prose
- **THEN** parse succeeds and yields that object
