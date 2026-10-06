## Purpose

Makes the Project module read as opaque dark blue instead of the mauve header, selection, and preview wash.

## ADDED Requirements

### Requirement: Project chrome is opaque dark blue
The Project header bar MUST be an opaque dark blue, the same family as the page background. It MUST NOT be a translucent wash. The selected section tab, the selected project row, and the selected Markdown or HTML pill MUST use a blue accent. Those controls MUST NOT use a mauve, pink, or indigo fill.

#### Scenario: Header and selection are blue
- **WHEN** an operator opens Project and selects a project and the HTML view
- **THEN** the header bar is opaque dark blue
- **AND** the selected project row and the HTML pill are blue, not mauve

### Requirement: The document preview is not an indigo wash
The Project document preview background MUST stay in the dark-blue family. It MUST NOT start its gradient from indigo `#1e1b4b`.

#### Scenario: Preview background is navy
- **WHEN** the operator views a project document preview
- **THEN** the preview background is dark blue
- **AND** it does not use an indigo wash at the top
