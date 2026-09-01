## Purpose

Makes MorphNotes Task start and end dates required so every case/task has a calendar window, defaulting new drafts to the current day.

## ADDED Requirements

### Requirement: Start and end dates are required
Creating or updating a MorphNotes Task MUST include both a start date and an end date. Empty start or end MUST be rejected. The operator MUST see both fields as required in the create/edit form.

#### Scenario: Save without start date
- **WHEN** an operator tries to save a Task with an empty start date
- **THEN** the Task is not saved
- **AND** the form or API indicates start date is required

#### Scenario: Save without end date
- **WHEN** an operator tries to save a Task with an empty end date
- **THEN** the Task is not saved
- **AND** the form or API indicates end date is required

#### Scenario: Save with both dates
- **WHEN** an operator saves a Task with both start and end set
- **THEN** the Task is persisted with those dates

### Requirement: New tasks default to today
Opening a new MorphNotes Task draft MUST pre-fill start and end to the operator’s current local calendar day. The operator MAY change either date before save.

#### Scenario: Open new task form
- **WHEN** an operator opens New task
- **THEN** start date is today’s local date
- **AND** end date is today’s local date

#### Scenario: Operator changes the default
- **WHEN** an operator changes start or end away from today and saves
- **THEN** the saved Task uses the dates they entered, not a forced today overwrite
