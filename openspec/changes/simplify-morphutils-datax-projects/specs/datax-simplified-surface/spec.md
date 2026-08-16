## Purpose

Simplify Morph Utils Data Access (DataX) by removing Docs and Databases and limiting Data reports to file-upload creation only.

## ADDED Requirements

### Requirement: Docs module is not offered in Data Access
The Data Access product SHALL NOT present a Docs navigation item or Docs workspace as a primary surface. Requests to former Docs paths SHALL redirect to a remaining Data Access surface (for example Data tables or Data reports) or return a clear not-available experience — they MUST NOT remain a broken primary workflow.

#### Scenario: Nav has no Docs
- **WHEN** an authenticated user opens Data Access navigation
- **THEN** Docs is not listed as a nav item

#### Scenario: Legacy Docs path does not open Docs workspace
- **WHEN** a user navigates to a legacy Docs URL (including Morph Utils deep-links that previously opened DataX Docs)
- **THEN** they are redirected away from Docs (or otherwise do not land in a Docs module that errors)

### Requirement: Databases module is not offered in Data Access
The Data Access product SHALL NOT present a Databases / database-connection navigation item or primary UX for linking external databases.

#### Scenario: Nav has no Databases
- **WHEN** an authenticated user opens Data Access navigation
- **THEN** Databases is not listed as a nav item

#### Scenario: Legacy Databases path is retired from primary UX
- **WHEN** a user navigates to a former Databases path
- **THEN** they are redirected to a remaining surface or receive a clear not-available response rather than a connection-management workspace

### Requirement: Data reports is file-upload only
Data reports SHALL support creating a report from an uploaded file. The Data reports UI MUST NOT offer **Visual** or **SQL** builder modes.

#### Scenario: File mode is available
- **WHEN** the user opens Data reports
- **THEN** they can upload a file and produce a data report from that file

#### Scenario: Visual mode absent
- **WHEN** the user opens Data reports
- **THEN** there is no Visual mode control or Visual query-builder workspace

#### Scenario: SQL mode absent
- **WHEN** the user opens Data reports
- **THEN** there is no SQL mode control or SQL editor workspace for building reports

### Requirement: Morph Utils Data Access copy matches the simplified surface
Morph Utils SHALL describe Data Access without advertising Docs, databases, or query/dashboard linking that has been removed.

#### Scenario: Module description updated
- **WHEN** a user views the Morph Utils Data Access module label/description
- **THEN** it does not claim Docs or Databases as features of the module
