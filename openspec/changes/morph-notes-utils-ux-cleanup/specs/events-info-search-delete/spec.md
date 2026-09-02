## Purpose

Lets operators find Events & Info rows by title and delete one event or several at once from the list, with confirmation.

## ADDED Requirements

### Requirement: Title search on the Events & Info list

The Events & Info list SHALL provide a title search control. Typing a query SHALL show only events whose title matches (case-insensitive substring). Clearing the query SHALL show the unfiltered list again.

#### Scenario: Filter by title

- **WHEN** a user types a title fragment in the Events & Info search field
- **THEN** the list shows only events whose title contains that fragment (case-insensitive)

#### Scenario: Empty search shows all

- **WHEN** the search field is empty
- **THEN** the list is not title-filtered

### Requirement: Delete one event from the list

The Events & Info list SHALL let the user delete a single event without opening it only for that purpose. Delete MUST ask for confirmation. After confirm, the event MUST be removed from storage and from the list.

#### Scenario: Delete one event

- **WHEN** a user chooses delete on one event and confirms
- **THEN** that event is gone from the list
- **AND** it is not returned by a later list load

#### Scenario: Cancel delete

- **WHEN** a user starts delete and cancels the confirmation
- **THEN** the event remains

### Requirement: Batch delete events

The Events & Info list SHALL let the user select multiple events and delete the selection in one confirmed action. Batch delete MUST NOT run without confirmation.

#### Scenario: Batch delete selected

- **WHEN** a user selects two or more events and confirms batch delete
- **THEN** those events are gone from the list
- **AND** unselected events remain

#### Scenario: Batch delete disabled with none selected

- **WHEN** no events are selected
- **THEN** batch delete is not offered as an active action
