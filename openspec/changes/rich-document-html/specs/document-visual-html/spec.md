## Purpose

Makes generated timelines, stories, research pages, and case-task documents readable as designed pages, with phase cards, callouts, and charts, instead of a flat list of bullets.

## ADDED Requirements

### Requirement: Sections are cards, not a flat column
Generated HTML for a timeline, story, research page, or case-task document MUST render each second-level section as its own card. A list of milestones under a section MUST render as a vertical rail with a marker on each item. The page MUST keep a dark background. The model MUST NOT be asked to return raw HTML.

#### Scenario: Phased timeline
- **WHEN** a timeline's markdown has two second-level headings and a bullet list under each
- **THEN** the HTML contains two section cards
- **AND** each bullet is a rail item rather than a plain list in one prose column

### Requirement: Callouts and stats are special blocks
A blockquote in the markdown MUST render as a callout. A stat block MUST render as a row of labeled values and as a bar chart. The chart MUST use only the values in that block. Labels and values MUST be text the generator was given, not invented quantities.

#### Scenario: Counts become a chart
- **WHEN** the markdown includes a stat block with Places 5 and Talks 20
- **THEN** the HTML shows those two labels with values 5 and 20
- **AND** the HTML includes a bar chart for those two values

#### Scenario: No quantities
- **WHEN** the markdown has sections and bullets but no stat block
- **THEN** the HTML still uses section cards and a rail
- **AND** the HTML does not include a chart

### Requirement: Generators ask for the structure
Timeline, story, and research generation MUST ask for a lede, second-level sections, and an optional stat block only when the source contains quantities. They MUST ask for a callout when the source contains a constraint or warning. They MUST tell the model not to invent numbers and not to return raw HTML.

#### Scenario: Timeline prompt
- **WHEN** the timeline generator builds its prompt
- **THEN** the prompt asks for second-level sections and an optional stat block
- **AND** the prompt says not to invent numbers and not to return raw HTML
