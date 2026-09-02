## Purpose

Keeps Morph AI Files as a durable workspace tied to each chat session: the opened folder stays put, switching sessions brings back that session’s files and pins, and recent folders are one click away—without a tutorial lede on the Files tab.

## ADDED Requirements

### Requirement: Folder stays the session workspace

When the user opens a local folder on the Morph AI Files tab, that folder MUST remain the workspace for the **current chat session** until they close it or open a different folder in that session. Switching to another chat session MUST show that other session’s workspace (or empty Files if it has none), not a shared global folder. Reloading Morph AI MUST restore the current session’s folder when the browser still grants access to it.

#### Scenario: Open folder and stay in the session

- **WHEN** the user opens a folder on Files while in a chat session
- **THEN** Files lists that folder
- **AND** navigating away from Files and back in the same session still shows that folder

#### Scenario: Switch chat sessions

- **WHEN** session A has folder F open (with any pins)
- **AND** the user switches to session B then back to session A
- **THEN** session A still shows folder F and the same pins
- **AND** session B does not inherit folder F unless the user opened F there too

#### Scenario: Reload Morph AI

- **WHEN** the user has a folder open for the current session and reloads the Morph AI page
- **AND** the browser still allows access to that folder
- **THEN** Files shows that folder again without requiring Open folder first

### Requirement: Recent folders for quick choose

The Files tab MUST keep a history of folders the user has opened and let them reopen one from that list without using Open folder, when the browser still allows access. The list MUST be ordered with the most recently opened first. Duplicate opens of the same folder MUST not flood the list.

#### Scenario: Reopen from recents

- **WHEN** the user has previously opened at least one folder
- **AND** they choose that folder from recents
- **THEN** that folder becomes the current session’s workspace
- **AND** Files lists its files

#### Scenario: Recents after a new open

- **WHEN** the user opens a folder they have not opened recently
- **THEN** that folder appears at the top of recents

### Requirement: No Files tutorial lede

The Files tab MUST NOT show the copy “Open a local folder to browse files. Pin files to include them in the agent. Morph never writes to this folder.” Empty Files MAY show a short status and Open folder plus recents only.

#### Scenario: Files tab with no folder

- **WHEN** the current session has no workspace folder
- **THEN** that tutorial sentence is not on the page
- **AND** Open folder (and recents if any) remain available
