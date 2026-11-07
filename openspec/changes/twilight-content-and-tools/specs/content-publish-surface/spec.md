# Spec Delta

## Purpose

Makes the Publish pane inside Contents & Data read as Content, in twilight glass, with an enlargeable preview and HTML generated from a chosen data record.

## ADDED Requirements

### Requirement: Publish module is named Content
The Publish pane SHALL label its product chrome Content. It SHALL NOT label that chrome Content Maker. The DorphNotes drawer item SHALL stay Contents & Data.

#### Scenario: Open the Publish pane
- **WHEN** an operator opens Contents & Data on the Publish tab
- **THEN** the embedded module title is Content
- **AND** the title is not Content Maker

#### Scenario: Drawer label
- **WHEN** an operator opens the DorphNotes drawer
- **THEN** the item is still Contents & Data

### Requirement: Content shell uses twilight glass
The Content shell SHALL use a dark purple page, translucent panels, light text, and a purple accent matching Dorph AI twilight glass. It SHALL NOT use a blue primary chrome.

#### Scenario: Publish chrome
- **WHEN** an operator views the Content shell
- **THEN** panels are translucent on a purple page
- **AND** the accent is purple rather than blue

### Requirement: HTML preview enlarges when it has content
When the HTML preview is a real page, the operator SHALL be able to open that same preview in a large modal. The empty starter preview SHALL NOT offer enlarge.

#### Scenario: Preview has a page
- **WHEN** the preview document is not the empty starter
- **THEN** the operator can open it in a near-viewport modal
- **AND** closing the modal returns to the inline preview

#### Scenario: Starter preview
- **WHEN** the preview is still the empty starter
- **THEN** enlarge is not offered

### Requirement: Generate from Data uses a chosen record
The control labeled Generate from Data SHALL open a picker of existing generic-data records. A PDF or other prose record SHALL generate an article page. A CSV, Excel, or JSON record SHALL generate a page with a data table and summary figures. The system SHALL NOT invent a second HTML publisher.

#### Scenario: PDF record
- **WHEN** the operator generates from a PDF record
- **THEN** the page is an article built from that record
- **AND** it is not a dashboard of rows

#### Scenario: Table record
- **WHEN** the operator generates from a CSV, Excel, or JSON record
- **THEN** the page includes a data table and summary figures from that record

#### Scenario: Record list unavailable
- **WHEN** the record list cannot be loaded
- **THEN** the operator sees an error
- **AND** generating from pasted source text still works
