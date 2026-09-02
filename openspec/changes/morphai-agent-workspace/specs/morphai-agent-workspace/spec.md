## Purpose

Puts Morph AI agent work in a right-hand pane with three tabs: a folder of files, Notes & TODOs, and Context & Knowledge.

## ADDED Requirements

### Requirement: Three workspace tabs
The workspace pane MUST provide three tabs labeled **Files**, **Notes & TODOs**, and **Context & Knowledge**. The last selected tab MUST persist for the current session. Header shortcuts that today open overlay drawers MUST open the matching workspace tab instead.

#### Scenario: Switch tabs
- **WHEN** the user selects Notes & TODOs
- **THEN** that tab’s content is visible in the workspace pane
- **AND** Files and Context & Knowledge are not shown at the same time

#### Scenario: Header opens tab not overlay
- **WHEN** the user clicks the Morph AI header control that previously opened Notes & TODOs or HybridContext as an overlay
- **THEN** the workspace shows the matching tab
- **AND** a full-screen overlay drawer is not required to use that content

### Requirement: Files tab opens a folder
The Files tab MUST let the user open a local folder as the session workspace and list the files in it (tree or equivalent). The user MUST be able to select files to pin into agent context. The product MUST NOT write or delete files in that folder as part of this capability.

#### Scenario: Open folder
- **WHEN** the user chooses Open folder and picks a directory the browser allows
- **THEN** the Files tab lists that folder’s files
- **AND** the folder remains the workspace until the user closes it or picks another

#### Scenario: Folder not available
- **WHEN** the browser cannot grant a directory (no picker or user cancels)
- **THEN** the Files tab shows a clear empty state and a way to retry
- **AND** chat still works without a folder

#### Scenario: Pin file for context
- **WHEN** the user pins a listed file
- **THEN** that file is marked as included for the agent until unpinned
- **AND** unpinning removes it from the include set

### Requirement: Notes & TODOs tab
The Notes & TODOs tab MUST show the same MorphData-synced notes and todos as the former Morph AI notes drawer, including AI assist already available there.

#### Scenario: Same data as MorphData
- **WHEN** the user opens the Notes & TODOs tab
- **THEN** they see the same items as MorphData notes/todos for that user
- **AND** create/edit/complete still persist to that store

### Requirement: Context & Knowledge tab
The Context & Knowledge tab MUST expose session HybridContext (files, paste, sources, attach/detach) and the Knowledge Library (list, upload, delete) that previously lived in the HybridContext drawer.

#### Scenario: Session context
- **WHEN** the user is on Context & Knowledge for the current chat session
- **THEN** they can view and manage that session’s HybridContext sources
- **AND** they can attach or detach the hybrid reference on the conversation

#### Scenario: Knowledge library
- **WHEN** the user uses Knowledge Library controls in that tab
- **THEN** they can list, upload, and delete knowledge files as today
