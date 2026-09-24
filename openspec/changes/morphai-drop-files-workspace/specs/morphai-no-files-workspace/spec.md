## Purpose

Removes Morph AI’s local-folder Files workspace so operators use Context & Knowledge (and Notes & TODOs) as the only agent-shell side surfaces for files and knowledge.

## ADDED Requirements

### Requirement: Workspace has no Files tab

On the Morph AI agent shell, the right workspace pane MUST provide **Notes & TODOs** and **Context & Knowledge** only. It MUST NOT show a **Files** tab, Open folder, recent folders, reconnect-folder, or a local-folder file tree.

#### Scenario: Agent shell tabs

- **WHEN** the operator opens the Morph AI agent workspace pane
- **THEN** the tab list is Notes & TODOs and Context & Knowledge
- **AND** there is no Files tab

#### Scenario: Stored Files tab is ignored

- **WHEN** a previous visit stored Files as the last workspace tab for a session
- **AND** the operator returns to that session
- **THEN** the workspace MUST NOT show a Files tab
- **AND** it MUST land on Context & Knowledge or Notes & TODOs

### Requirement: Context & Knowledge is the file surface

Session HybridContext (upload, paste, sources, attach/detach) and the Knowledge Library (list, upload, delete) MUST remain available on **Context & Knowledge**. Operators MUST NOT need a local-folder workspace to attach files to a chat.

#### Scenario: Attach files without a folder workspace

- **WHEN** the operator is on Context & Knowledge for the current chat session
- **THEN** they can manage that session’s HybridContext sources and the Knowledge Library as today
- **AND** chat still works with no local folder opened

### Requirement: No local-folder pins on send

The agent shell MUST NOT offer a composer control that includes or excludes pinned local-folder files. Send MUST NOT depend on a browser directory handle or a Files-workspace pin list.

#### Scenario: Include bar without Files

- **WHEN** the operator looks at the context-include chips on the next send
- **THEN** Notes and Knowledge remain available
- **AND** there is no Files chip for local-folder pins

### Requirement: Return visit does not restore a folder workspace

Returning to Morph AI MUST restore the last chat session when it still exists. Workspace open/closed preference and the last selected Notes or Knowledge tab MUST still apply. Morph AI MUST NOT bind, list, or reopen a local folder as that session’s workspace, MUST NOT retitle a session from a folder name, and MUST NOT open the workspace on Files.

#### Scenario: Reload without a folder pane

- **WHEN** the operator had previously opened a local folder on Files
- **AND** they reload Morph AI
- **THEN** chat restores without showing a Files workspace or requiring reconnect
- **AND** Context & Knowledge still works for that session

#### Scenario: Last chat still restores

- **WHEN** the operator worked in session A and later opens Morph AI on the same origin
- **THEN** session A is the active chat when it still exists

### Requirement: Embedded chat unchanged

Embedded / `singleSession` MorphNotes chat MUST NOT gain or lose a Files workspace as part of this change.

#### Scenario: Embedded chat

- **WHEN** Morph AI runs as embedded / `singleSession` chat
- **THEN** it still has no Files workspace tab
- **AND** HybridContext and notes continue to use their existing embedded surfaces
