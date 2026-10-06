## Purpose

Renames MorphNotes Big notes and Tasks in the UI, and shows Tasks as a searchable grid of equal square stickers.

## ADDED Requirements

### Requirement: Big notes is labeled Stories
MorphNotes MUST show the Big notes module as Stories in navigation and on its page. The Timelines module MUST stay labeled Timelines. The Stories label MUST NOT take the Timelines address. Big notes and Tasks addresses and APIs MUST stay unchanged.

#### Scenario: Nav and page say Stories
- **WHEN** an operator opens MorphNotes navigation
- **THEN** the former Big notes item reads Stories
- **AND** Timelines still reads Timelines
- **AND** opening Stories still uses the existing Big notes address

### Requirement: Tasks is labeled Stick notes
MorphNotes MUST show the Tasks module as Stick notes in navigation, on the list page, and on the create and detail headings. The list MUST remain the same records as Tasks.

#### Scenario: Nav and page say Stick notes
- **WHEN** an operator opens the Tasks module
- **THEN** the navigation item and the page heading read Stick notes
- **AND** the address is still the existing Tasks address

### Requirement: Stick notes are equal square stickers
The Stick notes list MUST be a grid of square stickers that all share one size. The gap between stickers MUST be small, from 6px to 12px. Each sticker MUST show its title and more than one line of its body when the body has more than one line. Each sticker MUST use a color from a fixed palette of at least twelve visually distinct colors. Colors MUST be assigned from the full list ordered by record id, including records hidden by search. The first twelve records in that order MUST use different colors. Reloading or changing the search MUST NOT change a record's color. Clicking a sticker MUST open the existing detail view.

#### Scenario: Grid of equal stickers
- **WHEN** Stick notes has at least two records
- **THEN** they render as equal squares in a grid
- **AND** the space between them is between 6px and 12px
- **AND** a sticker with a long body shows more than one line of that body

#### Scenario: Colors stay put
- **WHEN** the operator reloads Stick notes or changes the search text
- **THEN** each record keeps the same color
- **AND** the first twelve records in id order use different colors

#### Scenario: Open detail
- **WHEN** the operator activates a sticker
- **THEN** the existing detail view for that record opens

### Requirement: Stick notes can be searched
The Stick notes page MUST offer a search field that filters the grid by title or body text, case-insensitively. A query that matches nothing MUST show an empty result, not the unfiltered grid. Clearing the query MUST show the full grid again.

#### Scenario: Filter by body text
- **WHEN** the operator types a word that appears only in one record's body
- **THEN** the grid shows that sticker and hides the others

#### Scenario: No match
- **WHEN** the operator types text that matches no title or body
- **THEN** the grid shows no stickers

#### Scenario: Clear search
- **WHEN** the operator clears the search field
- **THEN** every sticker is shown again
