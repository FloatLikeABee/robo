## Purpose

Removes the short introductory paragraph under Morph Utils module page titles so each screen starts with the title and actions only.

## ADDED Requirements

### Requirement: No intro copy under Morph Utils page titles

Morph Utils module pages (Event Logs, Info Sheets, Data Access, Content Maker, Project) SHALL NOT show a descriptive paragraph immediately under the page heading. The heading and primary actions MAY remain. Helper text inside dialogs, empty states, and form fields is allowed.

#### Scenario: Events & Info has no operational-notes lede

- **WHEN** a user opens Events & Info
- **THEN** the page heading Events & Info is shown
- **AND** the line about operational notes, import from files, URL, or paste is not shown

#### Scenario: Other Utils modules match

- **WHEN** a user opens Info Sheets, Data tables, Docs, or other Morph Utils module pages that currently have a one-line intro under the h1
- **THEN** that intro line is not shown
- **AND** the page title remains
