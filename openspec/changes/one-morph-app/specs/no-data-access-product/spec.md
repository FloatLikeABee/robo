## Purpose

Keeps Data Access out of the product so Morph does not offer a data-tables module.

## ADDED Requirements

### Requirement: Data Access is not offered
The product MUST NOT show a Data Access module in Morph AI, MorphNotes, or any other navigation. No screen in the product MUST embed the Data Access UI.

#### Scenario: No Data Access entry
- **WHEN** an operator looks through Morph AI and MorphNotes
- **THEN** there is no Data Access item
- **AND** no iframe loads the Data Access UI

### Requirement: The default start does not launch Data Access
`./start-all.sh` with no service argument, and `./start-all.sh start all`, MUST NOT start the Data Access API or UI. An explicit start of that service MAY still exist for the leftover code.

#### Scenario: Default start skips Data Access
- **WHEN** an operator runs `./start-all.sh` with no extra service name
- **THEN** the Data Access API and UI are not started
