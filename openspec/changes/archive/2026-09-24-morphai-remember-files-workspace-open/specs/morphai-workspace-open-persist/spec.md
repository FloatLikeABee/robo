## Purpose

Keeps Morph AI's right-hand workspace (especially Files) visible and on the tab the operator last used across browser restarts, and restores a session's folder workspace immediately when IndexedDB still has a binding.

## ADDED Requirements

### Requirement: Workspace pane open state persists across visits

On the Morph AI agent shell (not `singleSession` / embedded), the operator MUST be able to show or hide the right workspace pane. The open/closed choice MUST be stored in `localStorage` and restored on the next visit to Morph AI on the same origin.

#### Scenario: Operator opens workspace and returns later

- **WHEN** the operator opens the right workspace pane on main Morph AI
- **AND** they close the browser or navigate away and later open Morph AI again
- **THEN** the workspace pane is open without requiring another click

#### Scenario: Operator collapses workspace and returns later

- **WHEN** the operator hides the right workspace pane
- **AND** they return to Morph AI on a later visit
- **THEN** the workspace pane stays hidden until they open it again

#### Scenario: Embedded chat unchanged

- **WHEN** Morph AI runs as embedded / `singleSession` chat
- **THEN** workspace open persistence and the show/hide control MUST NOT apply

### Requirement: Active workspace tab persists across visits

The last selected workspace tab (**Files**, **Notes & TODOs**, or **Context & Knowledge**) MUST persist per chat `sessionId` in `localStorage` (not only `sessionStorage`) and MUST be restored when that session becomes active.

#### Scenario: Files tab survives browser restart

- **WHEN** the operator selects the Files tab in session A
- **AND** they close the browser and later open Morph AI with session A active
- **THEN** the Files tab is selected in the workspace pane

#### Scenario: Switch sessions restores each session's tab

- **WHEN** session A last used Notes & TODOs and session B last used Files
- **AND** the operator switches between A and B
- **THEN** each session shows its own last-selected tab

### Requirement: Return visit restores Files workspace when bound

When the current chat session has a folder binding in the existing Files workspace store, Morph AI MUST show the Files tab with that folder name and listing (or the existing reconnect prompt if permission is needed) as soon as the workspace pane is open—without requiring Open folder first.

#### Scenario: Folder binding on reload

- **WHEN** the operator had a folder open for the current session
- **AND** they reload Morph AI or return on a later visit with the workspace pane open
- **AND** the browser still allows access to that folder
- **THEN** the Files tab lists that folder's files immediately

#### Scenario: Permission needed after return

- **WHEN** the session still has a folder binding but the browser requires permission again
- **THEN** Files shows the folder name and a reconnect action
- **AND** chat remains usable
