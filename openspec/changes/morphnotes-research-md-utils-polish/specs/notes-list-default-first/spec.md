## Purpose

Opening MorphNotes Timelines or Big notes immediately shows the first item in the module list so the operator is not left on an empty “select a …” pane when records exist.

## ADDED Requirements

### Requirement: Timelines selects the first item on enter
When the operator navigates to Timelines and the list has at least one item, the system MUST select the first item in the current list order and show its detail. An empty list MUST keep the empty state.

#### Scenario: Non-empty Timelines list auto-selects first
- **WHEN** the operator opens Timelines and at least one timeline exists
- **THEN** the first item in the list is selected
- **AND** its detail (including Markdown) is shown without an extra click

#### Scenario: Empty Timelines list stays empty
- **WHEN** the operator opens Timelines and the list is empty
- **THEN** no item is selected
- **AND** the empty-state message remains

### Requirement: Big notes selects the first item on enter
When the operator navigates to Big notes and the list has at least one item, the system MUST select the first item in the current list order and show its detail. An empty list MUST keep the empty state.

#### Scenario: Non-empty Big notes list auto-selects first
- **WHEN** the operator opens Big notes and at least one note exists
- **THEN** the first item in the list is selected
- **AND** its detail is shown without an extra click

#### Scenario: Empty Big notes list stays empty
- **WHEN** the operator opens Big notes and the list is empty
- **THEN** no item is selected
- **AND** the empty-state message remains

### Requirement: Selection after delete follows the list
When the selected Timeline or Big note is deleted and other items remain, the system MUST select the new first item in the list. When the last item is deleted, the empty state MUST return.

#### Scenario: Delete selected Timeline with siblings left
- **WHEN** the operator deletes the selected Timeline and at least one Timeline remains
- **THEN** the first remaining list item is selected

#### Scenario: Delete last Big note
- **WHEN** the operator deletes the only Big note
- **THEN** nothing is selected
- **AND** the empty state is shown
