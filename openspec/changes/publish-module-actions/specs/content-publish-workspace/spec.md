# Spec Delta

## Purpose

Makes the publish workspace use its column for the source text and makes every action read as a button, with generate-from-data as the main source action.

## ADDED Requirements

### Requirement: The source field fills the column
The source text area SHALL grow into the unused height of the source column. It SHALL remain usable when that column is short.

#### Scenario: Tall column
- **WHEN** the publish workspace has leftover height under the source controls
- **THEN** the source text area takes that height

#### Scenario: Short column
- **WHEN** the source column is shorter than a comfortable writing area
- **THEN** the source text area keeps a usable minimum height

### Requirement: Source actions stack with data last
The source actions SHALL stack full width, one under another. Generate from data SHALL be the last action and the only primary action in that column.

#### Scenario: Action order
- **WHEN** an operator looks at the source column
- **THEN** Upload is above From text
- **AND** From text is above From data
- **AND** From data is the primary action

### Requirement: Source labels stay short
The source actions SHALL be labeled Upload, From text, and From data.

#### Scenario: Visible labels
- **WHEN** the source column is idle
- **THEN** the three actions read Upload, From text, and From data

### Requirement: Publish actions are filled buttons
An action in the publish module SHALL be a filled control. The pale accent stroke used on the panel edge SHALL NOT be what marks it as a button.

#### Scenario: Secondary action
- **WHEN** an operator sees Save, Upload, From text, or Apply
- **THEN** each is a filled button whose fill is distinct from the panel behind it

#### Scenario: Primary actions
- **WHEN** an operator sees Publish or From data
- **THEN** each uses the solid theme accent fill

### Requirement: Quiet actions stay quiet
Close, Clear, and Enlarge SHALL stay text actions without a filled button body.

#### Scenario: Preview enlarge
- **WHEN** a preview can be enlarged
- **THEN** Enlarge is a text action, not a filled button
