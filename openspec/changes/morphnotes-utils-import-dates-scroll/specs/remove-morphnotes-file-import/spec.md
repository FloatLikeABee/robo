## Purpose

Removes MorphNotes Settings File import so operators no longer see or call a dedicated import path that is no longer needed.

## ADDED Requirements

### Requirement: Settings has no File import surface
MorphNotes Settings MUST NOT show a **File import** nav item. Routes that previously rendered that page MUST NOT load a File import UI. The strings **File import** and the former configuration paths `configuration/file-import` and `configuration/data-import` MUST NOT appear as live Settings destinations.

#### Scenario: Settings nav without File import
- **WHEN** an operator opens MorphNotes Settings
- **THEN** they see remaining Settings items such as Users
- **AND** they do not see a File import item

#### Scenario: Old File import URL
- **WHEN** an operator opens `/admin/configuration/file-import` or `/admin/configuration/data-import`
- **THEN** they are not shown a File import page
- **AND** they are redirected to a remaining MorphNotes destination

### Requirement: Dedicated import API is gone
The MorphNotes API MUST NOT expose `POST /api/tran/generic-data/import`. Generic data list, create, extract, analyze, update, and delete MAY remain.

#### Scenario: Import POST is not registered
- **WHEN** a client sends `POST /api/tran/generic-data/import`
- **THEN** the server does not run the former Settings File import handler
- **AND** the response is not a successful import of a new generic-data row via that path

#### Scenario: Generic data extract still available
- **WHEN** an operator uses MorphNotes Generic data extract-for-review
- **THEN** `POST /api/tran/generic-data/extract` still works
