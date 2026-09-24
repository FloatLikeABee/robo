## Purpose

Gives Morph AI a permanently compact left session rail: wide enough to use, with a new-chat control and distinct colored session buttons instead of an expandable named list.

## ADDED Requirements

### Requirement: Session rail stays collapsed and wider

On the Morph AI main chat (not a single-session embed), the left sessions column MUST remain a vertical rail. It MUST NOT expand into a titled session list. The rail width MUST be at least 78px (30px wider than the current 48px collapsed strip).

#### Scenario: Main chat never shows the wide named list

- **WHEN** the operator opens Morph AI main chat on a desktop-width viewport
- **THEN** the left sessions column is a vertical rail at least 78px wide
- **AND** session titles are not shown as a stacked named list in that column
- **AND** no control expands that column to the previous ~220px titled list

#### Scenario: Single-session embed hides the rail

- **WHEN** Morph AI chat is shown in single-session mode
- **THEN** the left session rail is not shown

### Requirement: New session from the rail

The rail MUST include a **+** control that creates a new chat session.

#### Scenario: Plus creates a session

- **WHEN** the operator clicks **+** on the session rail
- **THEN** a new session is created and becomes the current session

### Requirement: Sessions as distinct colored buttons

Each chat session MUST appear as a colored button on the rail. Different sessions MUST use different colors. The current session MUST be visually distinct from the others. Activating a session button MUST switch to that session. Each button MUST expose the session title to assistive tech (and a tooltip).

#### Scenario: Switch by color button

- **WHEN** the operator has at least two sessions
- **THEN** each session is a colored button on the rail
- **AND** the two sessions do not share the same button color
- **WHEN** the operator activates a non-current session button
- **THEN** that session becomes current

#### Scenario: Current session is marked

- **WHEN** a session is current
- **THEN** its rail button is visually distinct from inactive session buttons
