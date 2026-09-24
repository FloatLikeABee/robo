# composerx-container Specification

## Purpose

Packages Content Maker's API and UI as one image so an operator can start that app, keep its data on a single volume, and supply secrets only through the environment. Health does not call Morph or MorphUtils.

## Requirements

### Requirement: One image serves the Content Maker API and UI
The repository MUST provide a container image that contains the Content Maker API binary and the production build of the Content Maker UI. That image MUST be the only application process for this capability. It MUST NOT include Event Logs, Data Access, Project, AI tools, Morph, MorphUtils, or Invite Signup. The process MUST listen on all interfaces. The listen port MUST be `COMPOSERX_PORT` when that variable is non-empty, otherwise `PORT`, otherwise `8043`. `GET /health` MUST return HTTP 200 and a JSON body whose `status` is `ok`, and MUST NOT call Morph or MorphUtils. When `COMPOSERX_UI_DIR` points at a directory that contains `index.html`, `GET /` MUST return that UI document without authentication. API routes other than `/health`, `/auth/`, and `/public/` MUST require a bearer token that Morph accepts at `USERS_PANEL_BASE_URL` (`GET /api/auth/user` returns HTTP 200). `X-User-Role`, `X-User-Roles`, and `X-User-Permissions` MUST NOT grant access. A local checkout that does not set `COMPOSERX_UI_DIR` MUST NOT serve that UI from the API process.

#### Scenario: Health does not depend on MorphUtils
- **WHEN** the image is running and Morph and MorphUtils are not reachable
- **THEN** `GET /health` returns HTTP 200
- **AND** the JSON body has `"status": "ok"`

#### Scenario: UI and API share the listen port
- **WHEN** the image is running with `COMPOSERX_UI_DIR` set to its built UI
- **THEN** `GET /` on that listen port returns the Content Maker UI document without a token
- **AND** `GET /templates` without a token is rejected

#### Scenario: Client role headers are not a session
- **WHEN** `GET /templates` includes `X-User-Role: admin` and no `Authorization` bearer
- **THEN** the response is HTTP 401

#### Scenario: Other apps are not in the image
- **WHEN** an operator inspects the image contents
- **THEN** the image does not contain the other application trees listed above

### Requirement: Local port selection prefers COMPOSERX_PORT
When both `COMPOSERX_PORT` and `PORT` are set, the API MUST listen on `COMPOSERX_PORT`. Local `start-all.sh` MUST keep the API on 8043 and the Vite UI on 8044.

#### Scenario: Root env PORT does not steal the API port
- **WHEN** `PORT` is `9090` and `COMPOSERX_PORT` is `8043`
- **THEN** the API listens on 8043

### Requirement: Data lives on one mount
The image MUST default `COMPOSERX_SQLITE_PATH`, `COMPOSERX_BADGER_PATH`, and `TRAN_FILE_STORAGE_PATH` under `/data`. A named volume mounted at `/data` MUST be enough for those stores to survive a container recreate. Local `start-all.sh` defaults (`./data` and `./storage` relative to the working directory) MUST stay unchanged when the image env defaults are not set.

#### Scenario: Local checkout paths stay relative
- **WHEN** Content Maker is started from a checkout without the image env defaults
- **THEN** SQLite and Badger still default under `./data`
- **AND** file storage still defaults to `./storage`

### Requirement: Secrets come from the environment only
The image build MUST NOT copy `.env` files or `ai.config.json`, and MUST NOT accept secrets as build arguments. The production UI build inside the image MUST set `VITE_API_BASE` empty so the UI calls the same origin. A committed example MUST list `USERS_PANEL_BASE_URL`, `MORPH_AI_API_KEY`, and `TRAN_OPENAI_API_KEY` by name and MUST NOT contain a usable key or a guessed public host.

#### Scenario: Example env has no real secret
- **WHEN** a reviewer reads the committed Content Maker production env example
- **THEN** `MORPH_AI_API_KEY` and `TRAN_OPENAI_API_KEY` are empty
- **AND** `USERS_PANEL_BASE_URL` is empty or a placeholder the runbook tells the operator to replace with `https://<morph public host>`

### Requirement: Operators have a short runbook
`composerx/backend/README.md` and `docs/agents/05-composerx.md` MUST describe how to build and run the image, which env vars the process reads, that `USERS_PANEL_BASE_URL` is the Morph auth base URL, and that `GET /health` does not call Morph. `docs/agents/12-build-deploy.md` MUST point at that image. The runbook MUST NOT tell the operator to follow `scripts/deploy.sh`.

#### Scenario: Reader can run the image without Render
- **WHEN** a reader follows the Content Maker image section
- **THEN** they can build and run the container locally
- **AND** the steps do not call Render
