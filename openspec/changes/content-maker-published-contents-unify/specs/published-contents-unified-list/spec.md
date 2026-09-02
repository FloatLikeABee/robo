## Purpose

Published contents shows each HTML file once, with row actions aligned to the text in that row.

## ADDED Requirements

### Requirement: Single Published contents table

The Content Maker Published contents view SHALL present saved HTML files and published pages in one table. Each HTML file SHALL appear as at most one row.

#### Scenario: Saved and published are not split

- **WHEN** an operator opens Published contents and there are saved HTML files and published pages
- **THEN** the page shows one table, not separate Saved HTML files and Published history tables

#### Scenario: Same file is not listed twice

- **WHEN** an HTML file exists as a saved draft and as a published page with the same identity (name/slug)
- **THEN** Published contents shows that file as one row, not two

### Requirement: Row actions align with row text

In the Published contents table, operation controls (View, Delete, Open, and any equivalent actions) SHALL be vertically centered with the name and path text in the same row.

#### Scenario: Buttons sit on the text midline

- **WHEN** an operator views a Published contents row that has a name, optional path, and action buttons
- **THEN** the buttons share the same vertical center as that row’s text, not a top-aligned cell vs a lower control

### Requirement: Published contents chrome

The view SHALL keep the heading **Published contents**. It SHALL NOT show an under-title lede. The app remains dark-only. Product names SHALL stay Content Maker and MorphUtils.

#### Scenario: No subtitle under the heading

- **WHEN** an operator opens Published contents
- **THEN** there is no sentence under the heading describing drafts and history
