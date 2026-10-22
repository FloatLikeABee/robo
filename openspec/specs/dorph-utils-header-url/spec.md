# dorph-utils-header-url Specification

## Purpose

Keeps DorphUtils off the Dorph AI header. DorphNotes is the in-app entry for notes and the hosted modules.

## Requirements

### Requirement: The header has no DorphUtils chip
The Dorph AI header MUST NOT render a DorphUtils chip in development or production, including when `REACT_APP_DORPH_UTILS_URL` is a public URL. DorphNotes MUST open the in-app DorphNotes modal.

#### Scenario: A public DorphUtils URL does not add a chip
- **WHEN** the UI is built with `REACT_APP_DORPH_UTILS_URL` set to a public HTTPS origin
- **THEN** the header does not show a DorphUtils chip
- **AND** activating DorphNotes opens the in-app modal
