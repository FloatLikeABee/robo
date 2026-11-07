# Spec Delta

## Purpose

Gives DorphNotes a single Contents & Data module so operators import files, run analysis, and publish that work as HTML without a separate Generic data nav item.

## ADDED Requirements

### Requirement: One Contents & Data nav item

DorphNotes SHALL show a single drawer item labeled Contents & Data. It SHALL NOT show a separate Generic data drawer item. The Contents & Data destination SHALL include the Generic data import/grid/analyze surface and the Content Maker publish surface.

#### Scenario: Drawer after merge

- **WHEN** an operator opens the DorphNotes drawer
- **THEN** they see Contents & Data and do not see Generic data as its own row
- **AND** choosing Contents & Data shows both data records and HTML publish

#### Scenario: Old Generic data URL

- **WHEN** an operator opens `/dorphdata/generic-data`
- **THEN** they reach Contents & Data (same module, not a dead page)

### Requirement: Publish imported analysis as HTML

From a Generic data record that has analysis (or after analysis is generated), the operator SHALL be able to publish HTML through the existing Content Maker publish API. The published page SHALL include the record title and the analysis body. A data excerpt MAY be included. The system SHALL NOT introduce a second HTML publishing product.

#### Scenario: Publish from a record

- **WHEN** the operator chooses publish as HTML on a record with AI analysis
- **THEN** Content Maker stores a published page whose HTML contains that title and analysis
- **AND** the operator can open that page from Contents & Data publish

#### Scenario: Publish unavailable without Content Maker

- **WHEN** Content Maker is not configured
- **THEN** the publish-as-HTML action reports that failure
- **AND** import and analyze still work
