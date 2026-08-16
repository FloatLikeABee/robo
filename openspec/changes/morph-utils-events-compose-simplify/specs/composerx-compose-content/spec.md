## Purpose

Defines ComposerX Content Maker navigation and compose UX: one Compose content flow with publish, Published contents list, and compose-time reference uploads instead of a standalone Reference docs module.

## ADDED Requirements

### Requirement: Single Compose content nav item
Content Maker navigation SHALL expose one primary authoring item labeled **Compose content** that replaces separate **Compose & Publish** and **Compose** entries. That surface MUST include content authoring and publish capability in the same flow.

#### Scenario: Open Content Maker nav
- **WHEN** a user views Content Maker navigation
- **THEN** they see **Compose content** as the unified authoring entry
- **AND** they do not see separate peer items labeled **Compose & Publish** and **Compose** for the same job

### Requirement: Published contents list
The former Publish Records list SHALL be labeled **Published contents** and remain available for browsing published pages.

#### Scenario: Browse published pages
- **WHEN** a user opens Published contents
- **THEN** they can list previously published content records

### Requirement: Reference upload inside Compose content
While composing content, users MUST be able to upload reference material for that compose session (or draft) without visiting a separate Reference docs module. The standalone **Reference docs** nav module MUST be removed after migration of needed upload behavior into Compose content.

#### Scenario: Attach reference while composing
- **WHEN** a user is in Compose content and uploads a reference file
- **THEN** that material is available to assist composing for that session/draft

#### Scenario: Reference docs module gone
- **WHEN** a user views Content Maker navigation after this change
- **THEN** there is no standalone **Reference docs** module item

### Requirement: Legacy paths redirect or land on Compose content
Bookmarks or deep links to the old Compose, Compose & Publish, or Reference docs pages SHALL land on Compose content (or Published contents when the link clearly targeted publish history), not a removed blank module.

#### Scenario: Old compose deep link
- **WHEN** a user opens a legacy Compose or Compose & Publish path
- **THEN** they are taken to Compose content
