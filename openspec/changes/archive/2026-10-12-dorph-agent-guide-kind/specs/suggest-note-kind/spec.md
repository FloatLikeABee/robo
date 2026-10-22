# Spec Delta

## Purpose

Map a user's ask to a Dorph note kind with a reason and a grill flag, using deterministic rules so an agent can pick a store without calling a model.

## ADDED Requirements

### Requirement: Tool returns kind, reason, and grill flag
`suggest_note_kind` MUST accept `intent` text and optional `content_hint`. It MUST return `kind`, `reason`, `use_grill`, and `confidence` between 0 and 1. It MUST NOT call a language model. Empty `intent` after trim MUST be `invalid_input` and MUST NOT invent a kind.

#### Scenario: Specific fact is a note
- **WHEN** a client calls `suggest_note_kind` with intent "API base URL is http://127.0.0.1:9090"
- **THEN** `kind` is `note`
- **AND** `use_grill` is false
- **AND** `reason` is non-empty

#### Scenario: Empty intent is invalid
- **WHEN** a client calls `suggest_note_kind` with an empty intent
- **THEN** the tool result is an error that names `invalid_input`
- **AND** no kind is returned

### Requirement: Actionable asks pick todo or task
A single human action MUST map to `todo`. Work with stages, an assignee, or a board MUST map to `task`.

#### Scenario: One action is a todo
- **WHEN** the intent is "Call the dentist to reschedule"
- **THEN** `kind` is `todo`

#### Scenario: Board work is a task
- **WHEN** the intent is "Assign the login bug to Bea and put it in Doing"
- **THEN** `kind` is `task`

### Requirement: Time-ordered asks pick timeline, event log, or stick
Dated steps or a chronology MUST map to `timeline`. One happening that should be logged MUST map to `event_log`. A reminder at a time MUST map to `stick_note`.

#### Scenario: Dated plan is a timeline
- **WHEN** the intent is "Trip itinerary for Japan: Day 1 Tokyo, Day 2 Kyoto"
- **THEN** `kind` is `timeline`

#### Scenario: Incident is an event log
- **WHEN** the intent is "Log the incident: API 500 at 14:02"
- **THEN** `kind` is `event_log`

#### Scenario: Timed reminder is a stick note
- **WHEN** the intent is "Calendar reminder Friday 9am: dentist"
- **THEN** `kind` is `stick_note`

### Requirement: Inquiry, long-form, and default
An inquiry that asks for sources or findings MUST map to `research`. Chaptered or long-form narrative MUST map to `big_note`. Any other specific fact or summary MUST map to `note`.

#### Scenario: Sourced inquiry is research
- **WHEN** the intent is "Research local LLM options and cite sources"
- **THEN** `kind` is `research`

#### Scenario: Chaptered doc is a big note
- **WHEN** the intent is "Draft the full story of the migration, chapter by chapter"
- **THEN** `kind` is `big_note`

### Requirement: Grill flag follows vagueness and share intent
`use_grill` MUST be true when the ask is under about 15 words with no specifics, when the user says help-me-write, or when the note will be shared. It MUST be false when `content_hint` already supplies rich context.

#### Scenario: Vague short ask uses the grill
- **WHEN** the intent is "write something"
- **THEN** `use_grill` is true

#### Scenario: Rich hint skips the grill
- **WHEN** the intent is "Save a session summary" and `content_hint` is a multi-sentence recap
- **THEN** `use_grill` is false

### Requirement: Table-driven examples stay deterministic
The repository MUST keep a table of about 30 example asks with expected `kind` and `use_grill`. Every row MUST pass without a language-model call.

#### Scenario: Table rows all match
- **WHEN** the example table is evaluated
- **THEN** every row returns the expected `kind` and `use_grill`
- **AND** no row invokes a language model
