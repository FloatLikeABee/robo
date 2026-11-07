# Spec Delta

## Purpose

Makes the Event Logs operator app look like Dorph AI twilight glass without frosting the public form a respondent fills in.

## ADDED Requirements

### Requirement: Event Logs operator chrome is twilight glass
The signed-in Event Logs shell SHALL use the twilight glass page, translucent panels, light text, and purple accent. It SHALL NOT use a blue or slate primary chrome.

#### Scenario: Operator opens Event Logs
- **WHEN** an operator opens Event Logs inside DorphNotes
- **THEN** the shell page is dark purple with translucent panels
- **AND** the accent is purple

### Requirement: Public forms stay solid and readable
A public respondent form SHALL use the same light ink and purple accent on a solid dark background. It SHALL NOT use a frosted mesh or animated glass backdrop.

#### Scenario: Respondent opens a public form
- **WHEN** a respondent opens a published Event Logs form
- **THEN** the page background is solid
- **AND** body text stays light on that background
