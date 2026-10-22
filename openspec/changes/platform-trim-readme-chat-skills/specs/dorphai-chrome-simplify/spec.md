## Purpose

Simplifies Dorph AI chrome by dropping unused header shortcuts and the AI tools new-tab escape hatch.

## ADDED Requirements

### Requirement: Dorph AI header has no notes, knowledge, or JSON-export icons
The Dorph AI header bar MUST NOT include quick buttons for Notes & TODOs, Context & Knowledge, or Export session (JSON). Skills, AI tools, DorphNotes/DorphUtils app links, clear chat, and sign out MAY remain. Notes & TODOs in DorphNotes, and the agent workspace Notes & TODOs and Context & Knowledge tabs, MUST still exist.

#### Scenario: Header icons are gone
- **WHEN** a signed-in operator views the Dorph AI chat header
- **THEN** they do not see buttons titled Notes & TODOs, Context & Knowledge, or Export session (JSON)
- **AND** they still see Skills and AI tools

### Requirement: AI tools has no Open in tab
The AI tools workspace drawer MUST NOT offer an Open in tab (or Open in new tab) control. Operators work inside the Dorph AI drawer iframe.

#### Scenario: Drawer header has close only
- **WHEN** the operator opens AI tools from Dorph AI
- **THEN** the drawer header does not contain Open in tab
- **AND** they can still close the drawer
