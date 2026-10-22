## Purpose

Makes the Dorph AI chat shell fill the viewport so conversation and composer use the full screen width.

## ADDED Requirements

### Requirement: Dorph AI chat fills the viewport width
The standalone Dorph AI chat shell SHALL occupy the full viewport width. The chat panel MUST NOT be capped at a fixed max-width below the viewport, and the page MUST NOT leave unused horizontal gutters around the shell.

#### Scenario: Open Dorph AI on a wide screen
- **WHEN** an authenticated user opens standalone Dorph AI on a viewport wider than 1600px
- **THEN** the chat shell spans the full viewport width
- **AND** no empty page background gutters remain on the left or right of the shell

#### Scenario: Open Dorph AI on a typical laptop
- **WHEN** an authenticated user opens standalone Dorph AI on a typical laptop viewport
- **THEN** the chat shell still spans the full viewport width
- **AND** header, messages, and composer remain usable without horizontal clipping of primary controls

### Requirement: Embedded Dorph Data chat layout is unchanged
When Dorph AI chat is embedded inside Dorph Data, the compact embedded layout SHALL remain. Full-viewport width applies only to the standalone Dorph AI shell.

#### Scenario: Open chat drawer in Dorph Data
- **WHEN** a user opens the embedded Dorph AI chat from Dorph Data
- **THEN** the chat stays inside the existing drawer/panel bounds
- **AND** it does not expand to the full browser viewport
