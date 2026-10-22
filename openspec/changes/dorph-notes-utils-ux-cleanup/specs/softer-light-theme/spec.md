## Purpose

Makes light theme less harsh by using muted gray page and paper colors instead of bright white on Dorph AI, DorphNotes, and Dorph Utils.

## ADDED Requirements

### Requirement: Light surfaces are muted, not stark white

In light mode, primary page background and main paper/surfaces for Dorph AI chat, DorphNotes admin, and Dorph Utils embedded modules SHALL use a slightly darker, warm or cool gray rather than pure `#ffffff` / `#fff` as the full-page canvas. Text MUST remain readable against those surfaces.

#### Scenario: DorphNotes light mode

- **WHEN** DorphNotes is in light theme
- **THEN** the page background is not pure white
- **AND** primary text remains readable

#### Scenario: Dorph AI light mode

- **WHEN** Dorph AI chat is in light theme
- **THEN** the page and main panels are not stark white

#### Scenario: Dorph Utils light mode

- **WHEN** Event Logs, Data Access, Content Maker, or Project is in light theme
- **THEN** the page canvas is muted relative to pure white
