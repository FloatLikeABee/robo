# Spec Delta

## Purpose

Drops Graphic documents from DorphTools Documents and makes the header title readable on the dark bar.

## ADDED Requirements

### Requirement: Documents has no Graphic documents tab
The Documents area SHALL offer Articles and ScholarForge. It SHALL NOT offer a Graphic documents tab or mount that generator.

#### Scenario: Open Documents
- **WHEN** an operator opens DorphTools Documents
- **THEN** they see Articles and ScholarForge
- **AND** they do not see Graphic documents

### Requirement: Header title is light ink
The DORPHTOOLS title SHALL use solid light ink `#f4f0fa`. It SHALL NOT use a clipped gradient fill.

#### Scenario: Header on the dark bar
- **WHEN** an operator looks at the DorphTools header
- **THEN** the title text is `#f4f0fa`
