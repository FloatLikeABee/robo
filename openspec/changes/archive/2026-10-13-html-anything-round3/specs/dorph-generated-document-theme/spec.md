# Spec Delta

## ADDED Requirements

### Requirement: Heading count ignores ###
A big note is thin unless it has at least three line-start `## ` headings. `###` headings MUST NOT count. If the note is still thin after one regenerate, Dorph MUST append deterministic `##` sections built from the idea.

#### Scenario: Two ## and many ### is still thin
- **WHEN** the markdown has `## Introduction`, `## Key Milestones`, and several `###` headings
- **THEN** the stored markdown has at least three `## ` headings

### Requirement: Big-note timeout is retryable
Big-note generate MUST wait longer than 90 seconds for Ollama. When the wait expires, the API MUST return a clean retryable error, not `context deadline exceeded`.

#### Scenario: Deadline becomes a retry error
- **WHEN** the model call hits the generate deadline
- **THEN** the error text tells the caller to retry
- **AND** it does not contain `context deadline exceeded`
