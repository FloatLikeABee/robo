## Purpose

Makes a Morph AI chat with an opened Files folder a durable workspace session: the same chat, Files pane, and folder come back the next time the operator enters Morph AI.

## ADDED Requirements

### Requirement: Opened folder is the chat’s workspace session

When the operator opens a local folder on Morph AI Files (agent shell, not embedded / `singleSession`), that chat session MUST become the workspace session for that folder. The folder MUST stay bound to that session until they close it or open a different folder in the same session. If the session title is still generic (`New chat`, `Chat`, or empty), Morph AI MUST set the session title to the folder name so the session rail identifies the workspace.

#### Scenario: Open folder on the current chat

- **WHEN** the operator opens a folder while a chat session is active
- **THEN** Files lists that folder as that session’s workspace
- **AND** switching away and back to the same session still shows that folder

#### Scenario: Generic session takes the folder name

- **WHEN** the operator opens a folder in a session titled `New chat`, `Chat`, or with an empty title
- **THEN** that session’s title becomes the folder name

#### Scenario: Custom title is kept

- **WHEN** the operator opens a folder in a session that already has a non-generic title
- **THEN** the session title MUST NOT be overwritten by the folder name

### Requirement: Last workspace session is restored on enter

On the Morph AI agent shell, the last active chat `sessionId` MUST persist in the browser for that origin. The next visit MUST select that session when it still exists. If it no longer exists, Morph AI MUST select another remaining session that still has a folder binding, otherwise the default session.

#### Scenario: Return visit restores the last chat

- **WHEN** the operator works in session A (with or without a folder)
- **AND** they leave Morph AI and later open it again on the same origin
- **THEN** session A is the active chat without requiring them to pick it from the rail

#### Scenario: Deleted last session falls back to a folder workspace

- **WHEN** the last session was deleted
- **AND** another remaining session still has a folder binding
- **THEN** Morph AI opens that remaining folder-workspace session

#### Scenario: No last session stored

- **WHEN** the operator has never had a stored last session on this origin
- **THEN** Morph AI opens the default session as today

### Requirement: Folder workspace opens by default on enter

When the restored (or current) session has a folder binding, entering Morph AI MUST show the right workspace open on **Files** with that folder’s listing (or the existing reconnect prompt). The operator MUST NOT have to reopen the pane, switch to Files, or pick the folder again. They MAY still hide the pane during the visit.

#### Scenario: Bound folder is visible on the next visit

- **WHEN** the operator had a folder open for the session that is restored on enter
- **AND** they return to Morph AI
- **AND** the browser still allows access to that folder
- **THEN** the workspace pane is open on Files
- **AND** Files lists that folder immediately

#### Scenario: Permission needed after return

- **WHEN** the restored session still has a folder binding but the browser requires permission again
- **THEN** Files is shown with the folder name and a reconnect action
- **AND** chat remains usable

#### Scenario: Session with no folder keeps prior pane preference

- **WHEN** the restored session has no folder binding
- **THEN** Morph AI MUST NOT force the workspace pane open solely because of this capability
- **AND** the existing open/closed preference still applies

### Requirement: Embedded chat unchanged

Embedded / `singleSession` MorphNotes chat MUST NOT persist last session, auto-open Files, or retitle from a folder.

#### Scenario: Embedded chat skips restore

- **WHEN** Morph AI runs as embedded / `singleSession` chat
- **THEN** last-session restore, folder-title, and default-open Files MUST NOT apply
