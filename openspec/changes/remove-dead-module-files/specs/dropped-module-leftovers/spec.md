# Spec Delta

## Purpose

Keeps the checkout free of files that exist only for products or UI modules the live stack already dropped, without treating a missing working-tree copy of a live product as permission to delete it.

## ADDED Requirements

### Requirement: Shared chat package is gone
The repository SHALL NOT contain a `platform-chat` application tree. Dorph AI SHALL remain the only product chat.

#### Scenario: No platform-chat tree
- **WHEN** a reviewer lists top-level application directories
- **THEN** there is no `platform-chat` directory
- **AND** no remaining product depends on `@robo/platform-chat`

### Requirement: Graphic documents page is gone
AI tools SHALL NOT ship a Graphic documents page or a frontend helper used only by that page. The Graphic document HTTP API SHALL remain.

#### Scenario: Documents has no generator page
- **WHEN** an operator opens AI tools Documents
- **THEN** they see Articles and ScholarForge
- **AND** they do not load a Graphic documents page

#### Scenario: Graphic document API stays
- **WHEN** a reviewer inspects the AI tools HTTP API
- **THEN** `POST /graphic-document/generate` still exists

### Requirement: Dropped-product icons are gone
The repository SHALL NOT keep unused Booki, Academi, or UsersPanel icon files under Dorph or DorphUtils public icons.

#### Scenario: Icons for gone products
- **WHEN** a reviewer lists public icon files under Dorph and DorphUtils
- **THEN** they do not find `booki-icon.svg`, `academi-icon.svg`, or `userspanel-icon.svg`

### Requirement: Live products stay
The repository SHALL keep DorphUtils, Project (`dorph-engi`), the GraphRAG worker, OpenSpec history, and Data Access Blueprint entries. A missing working-tree copy of a live product SHALL be restored from git rather than treated as a dropped module.

#### Scenario: Project is restored
- **WHEN** Project files are missing from the working tree but still tracked in git
- **THEN** those files are restored
- **AND** they are not committed as a deletion

#### Scenario: DorphUtils remains
- **WHEN** a reviewer lists top-level application directories
- **THEN** `dorph-utils/` is present

### Requirement: Unused means dropped, not unreferenced
Cleanup SHALL delete only leftovers of products or UI modules already dropped from the live stack. It SHALL NOT delete tests, skills, docs, or source whose only fault is no current import.

#### Scenario: Tests and skills stay
- **WHEN** cleanup finishes
- **THEN** OpenSpec history, skills, and existing tests remain
- **AND** no file is deleted solely because a search found no import
