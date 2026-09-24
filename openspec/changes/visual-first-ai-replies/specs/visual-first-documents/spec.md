## Purpose

Makes stored markdown and in-app HTML document previews show the same mermaid diagrams and charts that chats produce, instead of leaving them as dead code fences.

## ADDED Requirements

### Requirement: Markdown documents render mermaid

In-app markdown previews for MorphNotes Research and related notes, Project documents, Content Maker markdown, and Data Access analysis MUST draw mermaid fences as dark-theme diagrams.

#### Scenario: Research or Project markdown with mermaid

- **WHEN** the operator opens a markdown preview that contains a mermaid fence
- **THEN** they see a rendered diagram
- **AND** the preview stays dark (no light theme)

#### Scenario: Data Access analysis markdown

- **WHEN** a Data Access analysis markdown body includes a mermaid chart or diagram
- **THEN** the in-app preview draws it

### Requirement: HTML previews include live diagrams

In-app HTML document previews (MorphNotes HTML tabs, Project HTML, Content Maker HTML preview) MUST draw mermaid diagrams present in that HTML. Email clients and other non-preview HTML sinks are not required to run mermaid.

#### Scenario: HTML iframe preview

- **WHEN** the operator views generated HTML that contains mermaid markup
- **THEN** the in-app preview shows the diagram on the dark document theme

#### Scenario: Invalid mermaid in a document

- **WHEN** mermaid source in a document cannot be drawn
- **THEN** the preview still shows the source so the operator can edit it
