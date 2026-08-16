## Purpose

Defines Morph Utils branding and navigation for the Docs module (formerly Academi), so users see Docs consistently in the shell and embed chrome.

## ADDED Requirements

### Requirement: Docs module replaces Academi in Morph Utils
Morph Utils SHALL expose the former Academi embed as a module labeled **Docs**. User-facing shell labels, short labels, descriptions, and tooltip copy MUST use Docs (not Academi).

#### Scenario: Utils sidebar shows Docs
- **WHEN** a signed-in user views Morph Utils app navigation
- **THEN** the module formerly branded Academi is labeled Docs
- **AND** Academi is not shown as the primary module name

### Requirement: Docs embed chrome uses Docs branding
The Docs web/app chrome SHALL present the product name as **Docs** in the header brand area (and equivalent mobile chrome). Legacy “Academi” brand text MUST NOT appear as the primary product title.

#### Scenario: Docs header brand
- **WHEN** a user opens the Docs embed
- **THEN** the primary brand name shown is Docs
