## Purpose

Presents the Morph Utils SheetX module as Event Logs, with Events & Info first and Info Sheets second, so operators land on the log instead of the sheet builder.

## ADDED Requirements

### Requirement: Module is labeled Event Logs
The Morph Utils left-nav module that currently reads Survey Maker SHALL be labeled **Event Logs**. The embedded app header, document title, Morph landing Utils list, and in-app assistant title MUST use Event Logs. They MUST NOT show Survey Maker, SurveyX, or SurveysX as the product name.

#### Scenario: Utils sidebar
- **WHEN** a user opens Morph Utils
- **THEN** the left-nav item for this module is Event Logs
- **AND** its tooltip/description refers to events and info sheets, not surveys as the product name

#### Scenario: Embedded app chrome
- **WHEN** the user opens Event Logs
- **THEN** the in-iframe header title is Event Logs
- **AND** the browser tab title for that app is Event Logs

### Requirement: Info Sheets tab label
The sheet builder that is labeled AI Surveys (or Survey Maker) SHALL be labeled **Info Sheets**. Publish, public collect links, and the answers list MUST remain available under that tab.

#### Scenario: Info Sheets tab
- **WHEN** the user is in Event Logs
- **THEN** a header tab named Info Sheets opens the sheet builder and collected answers
- **AND** the user can still publish a sheet and collect responses

### Requirement: Events & Info comes first
The Event Logs header tabs SHALL list **Events & Info** before **Info Sheets**. Opening Event Logs (Utils embed and in-app default route) MUST show Events & Info first.

#### Scenario: Tab order
- **WHEN** the user views Event Logs header tabs
- **THEN** Events & Info is the first tab
- **AND** Info Sheets is the next tab

#### Scenario: Default landing
- **WHEN** the user opens Event Logs without a deep link
- **THEN** Events & Info is shown
- **AND** they are not dropped onto Info Sheets as the home view
