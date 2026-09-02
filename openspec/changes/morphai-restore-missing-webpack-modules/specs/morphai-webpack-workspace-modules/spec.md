## Purpose

Morph AI and MorphNotes admin screens load in the browser without webpack module-not-found errors for session workspace, applied-assistant, and extract-JSON UI.

## ADDED Requirements

### Requirement: Morph AI chat bundle resolves workspace modules
The Morph AI frontend SHALL compile and serve the main chat screen without missing-module errors for the session workspace, AI Tools drawer, applied-assistant channel, and agent context helpers the chat screen already imports.

#### Scenario: Developer loads Morph AI
- **WHEN** a developer starts the Morph AI frontend and opens the main chat
- **THEN** webpack does not report `Can't resolve` for the session workspace, AI Tools drawer, applied-assistant, or agent-context modules

#### Scenario: Session workspace files tab
- **WHEN** the main Morph AI chat renders the agent workspace Files tab
- **THEN** the folder picker and recents list load (no missing-module error for the files tab)

### Requirement: MorphNotes grid extract-JSON dialog resolves
MorphNotes admin data grids SHALL compile the extract-JSON-from-text dialog the grid already imports.

#### Scenario: Developer loads a MorphNotes data grid
- **WHEN** a developer compiles MorphNotes admin and opens a data grid that offers extract JSON from text
- **THEN** webpack does not report `Can't resolve` for that dialog module
