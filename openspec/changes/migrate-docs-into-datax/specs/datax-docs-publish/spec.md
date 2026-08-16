## Purpose

Replaces Docs Board with ComposerX-like publishing: shareable HTML that displays a document, with optional AI analysis driven by a user prompt.

## ADDED Requirements

### Requirement: Publish document as HTML
The system SHALL allow a user to publish a Docs document as an HTML page that displays the document content for readers with a shareable link.

#### Scenario: Successful publish
- **WHEN** a user publishes a Docs document
- **THEN** the system produces an HTML page (or public URL) that renders the document for outside readers

### Requirement: Optional AI analysis prompt on publish
When publishing, the user SHALL be able to supply an optional natural-language prompt. If provided, the published page SHALL include AI analysis of the document guided by that prompt. If omitted, the page SHALL still publish the document display without requiring analysis.

#### Scenario: Publish with analysis prompt
- **WHEN** a user publishes with a non-empty analysis prompt
- **THEN** the published HTML includes document display plus AI analysis produced for that prompt

#### Scenario: Publish without analysis prompt
- **WHEN** a user publishes with no analysis prompt
- **THEN** the published HTML displays the document and does not require an analysis section

### Requirement: Board community feed removed from Docs
The Docs Board community-style feed SHALL be removed or replaced by the publish flow so Docs no longer presents Board as a social post feed.

#### Scenario: Board entry points
- **WHEN** a user opens Docs Board entry points after migration
- **THEN** they reach document publish (or an empty/retired notice that points to publish), not the legacy community feed as the primary experience
