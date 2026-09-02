## Purpose

Lets Morph Data operators save a new Tasks row when no assignee is chosen, without a database NOT NULL failure on the legacy assignee columns.

## ADDED Requirements

### Requirement: Unassigned task create succeeds
The system MUST persist a new Morph Data case/task when the client sends a valid title and does not send assignees (or sends an empty assignee list). The save MUST NOT fail because `CaseTask.assignee_type` (or `assignee_id`) is NULL. Historical assignee rows remain readable; this change does not require the user to pick an assignee.

#### Scenario: Save new task with no assignees
- **WHEN** the user creates a new Tasks record with a title and no assignees
- **THEN** the task is stored and returned with an id
- **AND** the client does not receive a NOT NULL constraint error on `assignee_type`

#### Scenario: Empty title still rejected
- **WHEN** the user tries to save a new task with a blank title
- **THEN** the system rejects the request with a validation error
- **AND** no case/task row is created
