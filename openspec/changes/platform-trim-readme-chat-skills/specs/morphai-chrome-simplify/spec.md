## Purpose

Simplifies Morph AI chrome by dropping unused header shortcuts and the AI tools new-tab escape hatch.

## ADDED Requirements

### Requirement: Morph AI header has no notes, knowledge, or JSON-export icons
The Morph AI header bar MUST NOT include quick buttons for Notes & TODOs, Context & Knowledge, or Export session (JSON). Skills, AI tools, MorphNotes/MorphUtils app links, clear chat, and sign out MAY remain. Notes & TODOs in MorphNotes, and the agent workspace Notes & TODOs and Context & Knowledge tabs, MUST still exist.

#### Scenario: Header icons are gone
- **WHEN** a signed-in operator views the Morph AI chat header
- **THEN** they do not see buttons titled Notes & TODOs, Context & Knowledge, or Export session (JSON)
- **AND** they still see Skills and AI tools

### Requirement: AI tools has no Open in tab
The AI tools workspace drawer MUST NOT offer an Open in tab (or Open in new tab) control. Operators work inside the Morph AI drawer iframe.

#### Scenario: Drawer header has close only
- **WHEN** the operator opens AI tools from Morph AI
- **THEN** the drawer header does not contain Open in tab
- **AND** they can still close the drawer
