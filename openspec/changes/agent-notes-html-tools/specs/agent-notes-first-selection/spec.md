# Spec Delta

## Purpose

Opens Agent Notes on the first saved note so the operator sees a note without an extra click.

## ADDED Requirements

### Requirement: First note is selected on load
When Agent Notes finishes loading a non-empty list and nothing is selected yet, it SHALL select the first note. An empty list SHALL stay on the empty state.

#### Scenario: Notes exist
- **WHEN** Agent Notes loads one or more notes
- **THEN** the first note is selected and its body is shown

#### Scenario: No notes
- **WHEN** Agent Notes loads an empty list
- **THEN** no note body is shown

### Requirement: Back stays on the list
After the operator leaves the selected note to return to the list, the dialog SHALL stay on the list until they choose a note.

#### Scenario: Back from the first note
- **WHEN** the operator chooses back from the open note
- **THEN** the list is shown without a selected note
