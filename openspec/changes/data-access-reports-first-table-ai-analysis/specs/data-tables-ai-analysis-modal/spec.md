## Purpose

Data Access lets the user run AI analysis on a data table, read it as markdown in a modal, and download that markdown.

## ADDED Requirements

### Requirement: AI analysis opens a modal

From Data tables, the user MUST be able to run AI analysis for a chosen table. The result MUST open in a modal (not a new page and not only the Data reports builder). The UI stays dark; there is no light/dark switch.

#### Scenario: Analyze from table detail

- **WHEN** the user is on a data table detail page and chooses AI analysis
- **THEN** a modal opens for that table
- **AND** the analysis is generated for that table’s data

#### Scenario: Analyze from the Data tables list

- **WHEN** the user is on the Data tables list and chooses AI analysis for a table
- **THEN** the same modal opens for that table

### Requirement: Analysis is markdown and downloadable

The modal MUST show the analysis as rendered markdown (headings, lists, tables if present). The user MUST be able to download the analysis as a markdown file. Download MUST use the generated markdown, not a screenshot.

#### Scenario: Read as markdown

- **WHEN** analysis has finished
- **THEN** the modal body is rendered markdown, not a single unformatted blob

#### Scenario: Download

- **WHEN** the user downloads from the modal
- **THEN** they get a `.md` file containing that analysis

### Requirement: Failures stay in the modal

If Morph AI is not configured or the request fails, the modal MUST show an error. It MUST NOT pretend a successful report was generated.

#### Scenario: AI unavailable

- **WHEN** analysis cannot run (for example AI not configured)
- **THEN** the modal shows a failure message
- **AND** download is not offered as a successful report
