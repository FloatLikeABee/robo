# morph-workspace-pane Specification

## Purpose

Restores the Morph AI Notes & TODOs and Context & Knowledge pane to the opening layout from before the pane was forced open.

## Requirements

### Requirement: Workspace pane opening matches the saved choice
On the Morph AI agent shell, the operator MUST be able to show and hide the right-hand workspace pane. With no saved choice, the pane MUST start open on a desktop width and closed on a phone. A saved open or closed choice MUST win on the next visit, including on a phone. While the pane is closed it MUST NOT take a column or a bottom band, and it MUST NOT shorten the transcript.

#### Scenario: First desktop visit
- **WHEN** an operator opens Morph AI chat on a desktop width with no saved workspace choice
- **THEN** the workspace pane is open

#### Scenario: First phone visit
- **WHEN** an operator opens Morph AI chat on a phone with no saved workspace choice
- **THEN** the workspace pane is closed
- **AND** the transcript is not shortened by a workspace band

#### Scenario: Saved closed choice
- **WHEN** the operator hides the workspace pane and later reloads Morph AI
- **THEN** the pane stays hidden until they show it again

### Requirement: Notes and knowledge tabs keep their previous order
The workspace pane MUST offer Notes & TODOs and then Context & Knowledge, in that order, with those English labels. It MUST NOT offer a Files tab. The Agent notes modal MUST still open over the current chat from the header menu.

#### Scenario: Pane tabs
- **WHEN** the workspace pane is open
- **THEN** the first tab is Notes & TODOs and the second is Context & Knowledge

#### Scenario: Agent notes modal
- **WHEN** the operator chooses Agent notes from the chat header menu
- **THEN** Agent notes opens as a modal over the current chat
