## Purpose

Lets DorphNotes and AI tools open as large panels that cover most of the Dorph AI window.

## ADDED Requirements

### Requirement: DorphNotes opens as a large modal
The Dorph AI DorphNotes control MUST open DorphNotes in a modal over the chat. The modal MUST stay inside Dorph AI. Closing it MUST return the operator to the same chat. The modal body MUST show the DorphNotes UI, including its navigation.

#### Scenario: Open and close DorphNotes
- **WHEN** the operator activates DorphNotes in the Dorph AI header
- **THEN** a modal shows the DorphNotes UI over the chat
- **AND** closing the modal leaves the chat in place

### Requirement: Both modals cover most of the window
The DorphNotes modal and the AI tools modal MUST each use about 96 percent of the viewport width and about 96 percent of the viewport height. Neither MUST be capped near 1200 pixels wide. Both MUST use the same size rule.

#### Scenario: AI tools is as wide as DorphNotes
- **WHEN** the operator opens AI tools and then DorphNotes on the same window
- **THEN** both modals are about 96 percent of the viewport width
- **AND** neither is limited to about 1200 pixels
