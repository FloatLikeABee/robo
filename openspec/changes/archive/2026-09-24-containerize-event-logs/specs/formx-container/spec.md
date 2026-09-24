## Purpose

Packages Event Logs (API and UI) as one image so an operator can run it, keep its data on a single volume, and supply secrets only through the environment. Health does not depend on MorphUtils or the other embeds.

## ADDED Requirements

### Requirement: One image serves the Event Logs API and UI
The repository MUST provide a container image that contains the Event Logs API binary and the production build of the Event Logs UI. That image MUST be the only application process for this capability. The process MUST listen on all interfaces on `SERVER_PORT`, which the container entrypoint MUST set from `PORT` (default 29909). `GET /health` MUST return HTTP 200 and a JSON body with `"status": "healthy"` without calling Morph, MorphUtils, or any sibling embed. When the UI build is present, `GET /events-info` MUST return the UI document. Requests under `/api/` MUST NOT be answered with that document.

#### Scenario: Health does not call MorphUtils
- **WHEN** the image is running and MorphUtils is not running
- **THEN** `GET /health` returns HTTP 200
- **AND** the body reports status healthy

#### Scenario: Embed path is the UI
- **WHEN** the image is running with the UI build inside it
- **THEN** `GET /events-info` returns the Event Logs UI document
- **AND** `GET /api/v1/events-info` is not that document

### Requirement: Data lives on one mount
The image MUST default the SQLite file, the Badger directory, and the uploads directory under `/data`. A volume mounted at `/data` MUST be enough for those stores to survive a container recreate. Local `start-all.sh` defaults (`./data` and `./uploads` relative to the working directory, `SERVER_PORT` 29909) MUST stay unchanged when the image env defaults are not set. A root `.env` value of `PORT=9090` MUST NOT move the Event Logs listener off `SERVER_PORT`.

#### Scenario: Local checkout paths stay relative
- **WHEN** Event Logs is started from a checkout without the image env defaults
- **THEN** the database paths still default under `./data` and uploads under `./uploads`
- **AND** the listener stays on `SERVER_PORT` (default 29909) even if `PORT` is 9090

### Requirement: Secrets come from the environment only
The image build MUST NOT copy `.env` files or accept secrets as build arguments. A committed example MUST list variable names and MUST NOT contain a usable JWT, password, or API key. The root Morph image context MUST still exclude the `formx` tree.

#### Scenario: Example env has no real secret
- **WHEN** a reviewer reads the committed Event Logs env example
- **THEN** secret keys are empty or absent
- **AND** the Dockerfile has no secret build argument
