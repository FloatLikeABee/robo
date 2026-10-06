## Purpose

Keeps the Morph AI agent workspace visible and makes Context & Knowledge the first, default tab.

## ADDED Requirements

### Requirement: Workspace pane stays visible
The Morph AI agent shell MUST show the workspace pane on wide layouts and on phone layouts. The shell MUST NOT offer a control whose action hides or shows that pane. A previously saved closed preference MUST NOT hide the pane.

#### Scenario: Wide layout has no hide control
- **WHEN** an operator opens Morph AI on a layout wider than a phone
- **THEN** the workspace pane is visible beside the chat
- **AND** no Hide workspace or Show workspace control is present

#### Scenario: Saved closed preference is ignored
- **WHEN** the browser has a saved workspace-closed preference and the operator reloads Morph AI
- **THEN** the workspace pane is still visible

### Requirement: Context and Knowledge is the leading tab
The workspace tab list MUST show Context & Knowledge before Notes & TODOs. When the session has no stored tab, the selected tab MUST be Context & Knowledge. A stored Notes & TODOs choice for that session MUST still open Notes & TODOs.

#### Scenario: First visit selects Context and Knowledge
- **WHEN** an operator opens a chat session that has no stored workspace tab
- **THEN** Context & Knowledge is the selected tab
- **AND** it appears before Notes & TODOs

#### Scenario: Stored notes choice is kept
- **WHEN** the session has a stored Notes & TODOs tab and the operator reloads
- **THEN** Notes & TODOs is selected
- **AND** Context & Knowledge still appears first in the tab list
