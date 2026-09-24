## Purpose

Gives MorphNotes and Project a single dark blue and dark purple look for generated HTML documents and in-app HTML preview panes, so operators never see off-brand green or teal washes when reading or publishing content.

## ADDED Requirements

### Requirement: Shared dark document palette

Generated HTML documents and in-app HTML preview surfaces for MorphNotes (Timelines, Big notes, Research, Case/task) and Project MUST use a consistent dark palette: deep blue page background (`#0b1220` family), blue-grey cards/lines, sky-blue or indigo link accents, and a subtle dark-purple tint in the page gradient wash. They MUST NOT use teal, emerald, or green gradient washes as the primary document theme.

#### Scenario: Project document HTML generation

- **WHEN** Project creates or updates `html_content` from markdown
- **THEN** the embedded stylesheet uses the shared dark blue / purple palette
- **AND** link accents are blue or indigo, not teal (`#2dd4bf` or similar)

#### Scenario: MorphNotes timeline HTML generation

- **WHEN** a Timeline record has `html_content` built or regenerated
- **THEN** its stylesheet matches the same palette family as Project documents
- **AND** the gradient wash reads as blue-purple, not green

### Requirement: HTML tab preview matches generated documents

Where MorphNotes or Project shows an **HTML** tab or iframe preview of stored `html_content`, the visible page MUST appear on the shared dark palette. If stored HTML still carries a legacy green/teal stylesheet, the preview wrapper MUST normalize colors for display until the record is re-saved.

#### Scenario: Project HTML tab

- **WHEN** the operator opens the HTML tab for a Project document in MorphUtils / Project
- **THEN** the preview background and links read as dark blue / purple-grey
- **AND** the preview does not show a green radial wash

#### Scenario: MorphNotes HTML iframe

- **WHEN** the operator opens the HTML tab on Timelines, Big notes, Research, or Case/task
- **THEN** the iframe preview uses the shared palette (directly or via preview normalization)

### Requirement: Rendered Markdown previews align with HTML theme

In-app **rendered Markdown** panes for Project and MorphNotes document modules MUST use link and heading colors from the same token family as generated HTML (blue/indigo links, light headings on dark blue-grey surfaces). They MUST NOT use teal or bright green link colors.

#### Scenario: Project rendered Markdown

- **WHEN** the operator views Rendered markdown for a Project document
- **THEN** prose links and accents match the dark blue theme
- **AND** the panel does not read as teal-on-grey

### Requirement: Project module chrome matches MorphUtils blue-grey

The Project app shell (sidebar, cards, module accent in MorphUtils launcher) MUST use dark blue / dark grey tokens consistent with Morph AI and MorphNotes, not teal module accents or pink/rose as the primary chrome color. Destructive actions MAY keep a muted red affordance.

#### Scenario: MorphUtils Project module accent

- **WHEN** the operator selects Project in MorphUtils
- **THEN** the module accent color is blue-grey aligned with other MorphUtils modules
- **AND** it is not teal (`#2dd4bf`)

#### Scenario: Project app surfaces

- **WHEN** the operator uses the Project document workspace
- **THEN** page background and cards match the MorphUtils / MorphNotes dark blue family
- **AND** informational text uses muted blue-grey, not teal highlights
