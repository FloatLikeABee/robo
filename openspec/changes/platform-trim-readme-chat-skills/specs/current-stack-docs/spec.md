## Purpose

Keeps the root README and launcher aligned with the apps that still exist, and documents one repo-root env file as the only local config.

## ADDED Requirements

### Requirement: README lists only remaining products
The root README MUST describe only folders and products that exist in the repo: Morph AI / MorphNotes (`morph/`), MorphUtils (`morph-utils/` with Event Logs, Content Maker, Data Access, Project), FormsX/Event Logs (`formx/`), Content Maker (`composerx/`), Project (`morph-engi/`), Data Access (`SharpReport/`), AI tools (`bk/`). It MUST NOT present Booki, Academi, UsersPanel, or `morph-broadcast` as current apps. User-facing names MUST stay MorphNotes, MorphUtils, Event Logs, Content Maker, Data Access, Project, AI tools.

#### Scenario: README has no removed-app table rows
- **WHEN** an operator opens the root README project-folders table
- **THEN** they do not see Booki, Academi, UsersPanel, or morph-broadcast as live folders
- **AND** they see Morph AI / MorphNotes and MorphUtils with the current module names

### Requirement: Launcher names match remaining services
`start-all.sh` MUST NOT default-start services whose app folders are missing (Booki, Academi). Status, list, and help MUST not advertise those as part of the default stack.

#### Scenario: start-all list has no Booki or Academi
- **WHEN** the operator runs `./start-all.sh list`
- **THEN** the output does not include `booki-api`, `booki-ui`, `academi-api`, or `academi-ui` as default services

### Requirement: One local config file
The README MUST state that local/dev configuration is a single repo-root `.env` copied from `.env.example`. Nested leftover `.env` files MUST be described as ignored. Production MUST remain `deploy/.env.production`.

#### Scenario: README tells operators to copy only the root template
- **WHEN** an operator follows the README environment section
- **THEN** they are instructed to `cp .env.example .env` at the repo root
- **AND** they are not told to copy a per-app `.env` to run the stack
