## Purpose

Hosts the Docs product inside DataX Data access so users reach documents, Docs AI, and doc publishing without a separate Morph Utils Docs module.

## ADDED Requirements

### Requirement: Docs area inside DataX
The system SHALL expose a first-class Docs area within DataX (Data access) that authenticated users can open from DataX navigation.

#### Scenario: User opens Docs from DataX
- **WHEN** an authenticated user selects Docs in DataX navigation
- **THEN** the Docs workspace loads inside DataX without requiring a separate Morph Utils Docs module iframe

### Requirement: Document library for content files
The system SHALL provide a Docs document library where users can create, list, open, and upload content documents (PDF and TXT at minimum).

#### Scenario: Upload a PDF into Docs
- **WHEN** a user uploads a PDF through the Docs library or Docs AI attach flow
- **THEN** the file is stored as a Docs content document and is available for Docs AI and publish flows

### Requirement: Standalone Docs Utils module removed
Morph Utils SHALL NOT list a primary Docs (or Academi) module once Docs is available in DataX. Legacy Utils paths for Docs/Academi SHALL redirect users into DataX Docs.

#### Scenario: Legacy Utils Docs path
- **WHEN** a user opens a legacy Morph Utils Docs or Academi module path
- **THEN** they are redirected into DataX Docs (or DataX with a Docs deep link)
