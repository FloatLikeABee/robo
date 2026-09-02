## Purpose

Makes light theme less harsh by using muted gray page and paper colors instead of bright white on Morph AI, MorphNotes, and Morph Utils.

## ADDED Requirements

### Requirement: Light surfaces are muted, not stark white

In light mode, primary page background and main paper/surfaces for Morph AI chat, MorphNotes admin, and Morph Utils embedded modules SHALL use a slightly darker, warm or cool gray rather than pure `#ffffff` / `#fff` as the full-page canvas. Text MUST remain readable against those surfaces.

#### Scenario: MorphNotes light mode

- **WHEN** MorphNotes is in light theme
- **THEN** the page background is not pure white
- **AND** primary text remains readable

#### Scenario: Morph AI light mode

- **WHEN** Morph AI chat is in light theme
- **THEN** the page and main panels are not stark white

#### Scenario: Morph Utils light mode

- **WHEN** Event Logs, Data Access, Content Maker, or Project is in light theme
- **THEN** the page canvas is muted relative to pure white
