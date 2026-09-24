## Purpose

Presents Event Logs as Events & Info only, without an Info Sheets builder or tab for operators.

## ADDED Requirements

### Requirement: Info Sheets is gone from Event Logs

The Event Logs operator UI MUST NOT show an Info Sheets (or AI Surveys / Survey Bot) header tab or page. Opening Event Logs MUST land on Events & Info.

#### Scenario: Header has Events & Info only

- **WHEN** the operator opens Event Logs in MorphUtils or at the Event Logs app
- **THEN** the only section tab is Events & Info
- **AND** there is no Info Sheets tab

### Requirement: Old Info Sheets paths redirect

Operator routes that used to open Info Sheets MUST send the operator to Events & Info instead of a builder.

#### Scenario: survey-bot URL

- **WHEN** the operator opens `/survey-bot` (or equivalent Info Sheets path)
- **THEN** they are taken to Events & Info

### Requirement: MorphUtils copy matches

MorphUtils MUST describe Event Logs without mentioning Info Sheets.

#### Scenario: Utils module description

- **WHEN** the operator reads the Event Logs module in MorphUtils
- **THEN** the copy does not mention Info Sheets
