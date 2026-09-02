## Purpose

Gives every Morph platform app one local environment file so shared secrets and URLs are set once and applied everywhere.

## ADDED Requirements

### Requirement: Single local environment file
Local development MUST read configuration from one gitignored file at the repository root. Operators MUST NOT need a separate `.env` inside each app directory for the stack to start.

#### Scenario: Copy once
- **WHEN** an operator copies the root environment example to a root `.env` and fills secrets
- **THEN** Morph, Event Logs / FormsX, Content Maker, Booki, Project / Morph Engi, Data Access, SharpReport, and other `start-all.sh` services can start without creating per-app `.env` files

#### Scenario: Shared AI key
- **WHEN** `MORPH_AI_API_KEY` is set in the root `.env`
- **THEN** every app that uses Morph AI MUST see that key
- **AND** an empty or missing per-app `.env` MUST NOT clear or replace that key

### Requirement: Launcher loads only the root file
`start-all.sh` MUST load the repository-root `.env` for each service. It MUST NOT load a nested app `.env` that would override shared keys.

#### Scenario: Restart after editing root file
- **WHEN** the operator changes a key in the root `.env` and restarts a service with `start-all.sh`
- **THEN** that service process receives the new value

#### Scenario: Nested empty file ignored
- **WHEN** a leftover `formx/backend/.env` (or similar) exists with `MORPH_AI_API_KEY` empty
- **AND** the root `.env` has a non-empty `MORPH_AI_API_KEY`
- **THEN** the FormsX process MUST use the root value

### Requirement: Direct process start still finds the root file
A backend or frontend started from its own directory without `start-all.sh` MUST still load the repository-root `.env`.

#### Scenario: Go run from app folder
- **WHEN** an operator runs a Go API from its backend directory
- **THEN** the process MUST load keys from the repository-root `.env`

#### Scenario: Frontend dev server
- **WHEN** an operator runs a Vite or CRA frontend from its frontend directory
- **THEN** `VITE_*` / `REACT_APP_*` values defined in the root `.env` MUST be available to that dev server

### Requirement: Production stays a separate shared file
Deployed environments MUST keep using the existing single production env file. The local root `.env` MUST NOT be the production secret store.

#### Scenario: Production unchanged
- **WHEN** an operator deploys with `deploy/.env.production`
- **THEN** production configuration continues to come from that file (or the host’s env), not from a developer’s local root `.env`

### Requirement: Template documents all keys
The committed root example file MUST list shared keys (`MORPH_AI_*`, Morph auth URL, JWT) and app-specific keys (ports, sqlite/badger paths, public URLs) in one place. Per-app example files MUST point operators to that root template instead of duplicating the full key list.

#### Scenario: New clone
- **WHEN** a developer opens the root environment example
- **THEN** they can see which keys are required for Morph AI and auth
- **AND** they can see per-app path and port keys without opening every app folder
