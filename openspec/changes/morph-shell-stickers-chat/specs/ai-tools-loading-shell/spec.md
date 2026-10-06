## Purpose

Replaces the white wait inside the Morph AI tools modal with a dark loading surface.

## ADDED Requirements

### Requirement: AI tools waits on a dark surface
While the AI tools modal is open and its embedded app has not finished loading, the modal body MUST show a dark background and a visible loading indicator. The modal body MUST NOT use a white background during that wait. Opening the modal again MUST show the same dark loading state until the embedded app finishes loading again.

#### Scenario: First open
- **WHEN** the operator opens AI tools and the embedded app has not finished loading
- **THEN** the modal body is dark and shows a loading indicator
- **AND** the body is not a white panel

#### Scenario: Reopen
- **WHEN** the operator closes AI tools and opens it again before the embedded app has finished loading
- **THEN** the dark loading indicator is shown again

#### Scenario: Load finished
- **WHEN** the embedded app finishes loading
- **THEN** the loading indicator is no longer shown
- **AND** the embedded app is visible
