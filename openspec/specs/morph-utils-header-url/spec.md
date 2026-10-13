# morph-utils-header-url Specification

## Purpose

Keeps MorphUtils off the Morph AI header. MorphNotes is the in-app entry for notes and the hosted modules.

## Requirements

### Requirement: The header has no MorphUtils chip
The Morph AI header MUST NOT render a MorphUtils chip in development or production, including when `REACT_APP_MORPH_UTILS_URL` is a public URL. MorphNotes MUST open the in-app MorphNotes modal.

#### Scenario: A public MorphUtils URL does not add a chip
- **WHEN** the UI is built with `REACT_APP_MORPH_UTILS_URL` set to a public HTTPS origin
- **THEN** the header does not show a MorphUtils chip
- **AND** activating MorphNotes opens the in-app modal
