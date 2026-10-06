## Purpose

Lets MorphNotes and AI tools open as large panels that cover most of the Morph AI window.

## ADDED Requirements

### Requirement: MorphNotes opens as a large modal
The Morph AI MorphNotes control MUST open MorphNotes in a modal over the chat. The modal MUST stay inside Morph AI. Closing it MUST return the operator to the same chat. The modal body MUST show the MorphNotes UI, including its navigation.

#### Scenario: Open and close MorphNotes
- **WHEN** the operator activates MorphNotes in the Morph AI header
- **THEN** a modal shows the MorphNotes UI over the chat
- **AND** closing the modal leaves the chat in place

### Requirement: Both modals cover most of the window
The MorphNotes modal and the AI tools modal MUST each use about 96 percent of the viewport width and about 96 percent of the viewport height. Neither MUST be capped near 1200 pixels wide. Both MUST use the same size rule.

#### Scenario: AI tools is as wide as MorphNotes
- **WHEN** the operator opens AI tools and then MorphNotes on the same window
- **THEN** both modals are about 96 percent of the viewport width
- **AND** neither is limited to about 1200 pixels
