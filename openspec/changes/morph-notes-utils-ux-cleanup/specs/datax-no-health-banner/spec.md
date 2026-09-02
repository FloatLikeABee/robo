## Purpose

Stops Data Access from polling a health endpoint and showing a standing offline banner, so a 500 on that check no longer blocks or alarm the Morph Utils Data Access module.

## ADDED Requirements

### Requirement: No periodic Data Access health banner

The Data Access UI MUST NOT poll a health endpoint on a timer or on tab focus for the purpose of showing a global offline banner. The strings **DataX API offline** and **DataX API health check failed** MUST NOT appear in the Data Access UI.

#### Scenario: Opening Data Access with a failing health route

- **WHEN** a user opens Data Access in Morph Utils and `/api/v1/health` would return 500
- **THEN** the page does not show a DataX API offline banner
- **AND** the page does not show a health check failed message

#### Scenario: A real Data Access request fails

- **WHEN** a Data Access list or save request fails
- **THEN** that failure is shown as an error for that action
- **AND** a standing health-check banner is still not shown
