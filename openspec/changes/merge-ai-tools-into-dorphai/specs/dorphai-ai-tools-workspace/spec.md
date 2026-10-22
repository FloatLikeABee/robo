## Purpose

Surface the full AI Tools (bk) application inside DorphAI as a large right-drawer workspace, so users can access every AI Tools module without leaving Dorph AI.

## ADDED Requirements

### Requirement: AI Tools workspace opens as a DorphAI right-drawer

DorphAI SHALL provide a control that opens an AI Tools workspace as a large right-side drawer overlay, using the same drawer presentation as the Notes & TODOs panel. The drawer SHALL be dismissible via a close control and via clicking the overlay outside the panel.

#### Scenario: Opening the AI Tools workspace

- **WHEN** a user activates the AI Tools control in DorphAI
- **THEN** a large right-side drawer opens showing the AI Tools workspace

#### Scenario: Closing the AI Tools workspace

- **WHEN** the AI Tools drawer is open and the user clicks the close control or the overlay outside the panel
- **THEN** the drawer closes and returns focus to the DorphAI chat

### Requirement: Workspace embeds the full AI Tools application

The AI Tools workspace SHALL embed the AI Tools (bk) web application so that all of its modules — including Assistants, RAG, Video Stories, Documents, and System — are available inside DorphAI. The embedded application SHALL receive the current Dorph session so it is authenticated. A control SHALL be provided to open the AI Tools application in a separate browser tab.

#### Scenario: All modules available

- **WHEN** the AI Tools workspace is open
- **THEN** the user can reach every AI Tools module (Assistants, RAG, Video Stories, Documents, System) from within the drawer

#### Scenario: Session carried into the embed

- **WHEN** the AI Tools workspace loads with an active Dorph session
- **THEN** the embedded application is opened with the Dorph session token

#### Scenario: Open in a new tab

- **WHEN** the user activates the "open in tab" control
- **THEN** the AI Tools application opens in a separate browser tab carrying the Dorph session

### Requirement: Clear failure state when AI Tools is unavailable

If the AI Tools application cannot be embedded (for example, its UI is not running), the workspace SHALL present a clear message and a way to open the application directly, rather than failing silently.

#### Scenario: Embed cannot load

- **WHEN** the embedded AI Tools application fails to load
- **THEN** the workspace shows an explanatory message and a link to open AI Tools in a new tab
