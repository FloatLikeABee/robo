## Purpose

Keeps the Morph AI Skills upload column fully visible: compact instructions, no left-pane scrollbar.

## ADDED Requirements

### Requirement: Left Skills pane fits without a scrollbar

In the Morph AI Skills dialog, the left upload column SHALL show Name, Description, Instructions, and the upload/improve actions without a vertical scrollbar on that column. The Instructions field SHALL be compact (not a tall multi-block editor). The right catalog MAY still scroll.

#### Scenario: Empty form fits the left pane

- **WHEN** an operator opens Skills from the Morph AI header at a typical desktop dialog size
- **THEN** the left column has no vertical scrollbar
- **AND** Name, Description, Instructions, and action buttons are all visible without scrolling that column

#### Scenario: Instructions is short in the layout

- **WHEN** the operator looks at the Instructions field in the left column
- **THEN** the field occupies a small block (about four visible lines), not a tall pane that forces the column to scroll

### Requirement: Long instruction text stays inside the field

If instruction text is longer than the compact field, that text SHALL scroll inside the Instructions field. It MUST NOT make the left column itself scroll.

#### Scenario: Long body does not grow the left pane

- **WHEN** the operator pastes a long skill body into Instructions
- **THEN** extra lines scroll inside Instructions
- **AND** the left column still has no vertical scrollbar

### Requirement: Skills chrome stays compact and dark

The dialog SHALL stay dark-only. The heading SHALL remain **Skills**. There SHALL NOT be an under-title lede under that heading.

#### Scenario: No subtitle under Skills

- **WHEN** an operator opens Skills
- **THEN** there is no sentence under the heading describing upload or Neo4j
