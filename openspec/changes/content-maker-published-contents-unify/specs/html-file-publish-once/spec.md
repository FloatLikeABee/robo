## Purpose

Each HTML file has at most one published public path; publishing again must not mint a second URL.

## ADDED Requirements

### Requirement: One public path per HTML file

The system SHALL allow an HTML file to be published at most once. Identity is the file’s publish name (normalized slug). A second Publish for the same identity MUST NOT create a second public path or a second published-history row.

#### Scenario: First publish creates one path

- **WHEN** an operator publishes an unpublished HTML file with a given name
- **THEN** the system creates one public path for that name and one published row for that file

#### Scenario: Publish again keeps the same path

- **WHEN** an operator publishes again using the same name as an existing published HTML file
- **THEN** the system does not mint a new slug or extra history row; the file still has one public path

#### Scenario: Distinct names stay distinct

- **WHEN** an operator publishes two HTML files with different names
- **THEN** each file has its own public path
