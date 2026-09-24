## Purpose

Keeps the Morph AI Files folder that the operator last opened for a chat session visible after a page refresh, including the folder name and file listing, even when the browser cannot keep a live directory handle.

## ADDED Requirements

### Requirement: Opened folder listing survives refresh

After the operator opens a local folder on Morph AI Files (agent shell), a reload of Morph AI on the same origin MUST still show that session’s folder name and the last known file listing. Morph AI MUST persist this even when the browser provides no directory handle (file-picker fallback). Files MUST NOT show the empty “No folder open” state solely because of a refresh.

#### Scenario: Refresh after opening a folder

- **WHEN** the operator opens a folder in Files for the current chat
- **AND** they reload Morph AI
- **THEN** Files still shows that folder’s name
- **AND** Files still lists the files from that open (or the last stored listing)

#### Scenario: Open without a directory handle still persists

- **WHEN** the operator opens a folder through a picker that does not yield a directory handle
- **AND** they reload Morph AI
- **THEN** Files still shows that folder name and listing
- **AND** Morph AI MUST NOT treat that open as clearing the workspace

### Requirement: Close folder is the only forget

Morph AI MUST forget the session’s folder workspace only when the operator closes the folder. A successful open MUST NOT clear an existing binding just because a directory handle is missing.

#### Scenario: Close then refresh

- **WHEN** the operator clicks Close folder
- **AND** they reload Morph AI
- **THEN** Files shows no folder open for that session

### Requirement: Reconnect is additive, not a blank workspace

When a stored directory handle exists but the browser requires permission again, Files MUST still show the folder name and last listing, plus a reconnect action. Chat remains usable.

#### Scenario: Permission needed after refresh

- **WHEN** the session has a stored folder and a handle that is not currently readable
- **THEN** Files shows the folder name and last listing
- **AND** a reconnect action is available
- **AND** the empty “No folder open” copy is not shown

### Requirement: Embedded chat unchanged

Embedded / `singleSession` MorphNotes chat MUST NOT persist a Files folder workspace.

#### Scenario: Embedded skip

- **WHEN** Morph AI runs as embedded / `singleSession` chat
- **THEN** this folder-persist behavior MUST NOT apply
