## Purpose

Data Access header navigation lists Data reports before Data tables so reports are the first section.

## ADDED Requirements

### Requirement: Data reports precedes Data tables

The Data Access header section nav MUST list **Data reports** before **Data tables**. Help MAY remain after both. Labels MUST stay “Data reports” and “Data tables”. The module name remains Data Access.

#### Scenario: Header order

- **WHEN** an authenticated user views Data Access
- **THEN** the section tabs appear in this order: Data reports, Data tables, then Help (if Help is shown)

#### Scenario: Labels unchanged

- **WHEN** the user reads the header tabs
- **THEN** they still say Data reports and Data tables (not DataX or other product names)
