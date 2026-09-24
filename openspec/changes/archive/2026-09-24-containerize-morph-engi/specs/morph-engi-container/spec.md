## ADDED Requirements

### Requirement: One image serves Project
The repository MUST provide `morph-engi/Dockerfile` whose build context is the repo root. The image MUST serve the Project HTTP API and the built Svelte UI on the same origin. The process MUST listen on `PORT` (image default 9096) on all interfaces. `GET /health` MUST return HTTP 200. When `STATIC_DIR` points at the built UI, `GET /` MUST return that UI, and an unknown non-API path MUST fall back to `index.html`. The image MUST NOT set `MORPH_ENGI_PORT`. The UI build MUST set `VITE_API_BASE_URL` empty and MUST NOT take it as a build argument.

#### Scenario: Health responds on the configured port
- **WHEN** the image is running with `PORT` set
- **THEN** `GET http://127.0.0.1:$PORT/health` returns HTTP 200
- **AND** the server is not bound only to a loopback interface

#### Scenario: The UI is same-origin
- **WHEN** the image is built and started with `STATIC_DIR` pointing at the built frontend
- **THEN** `GET /` returns the Project HTML
- **AND** the built assets do not embed a non-empty `VITE_API_BASE_URL`

### Requirement: Healthcheck expands PORT and uses wget
The Dockerfile MUST define a shell-form `HEALTHCHECK` that requests `http://127.0.0.1:${PORT}/health` with `wget`. It MUST NOT contain `$$`. It MUST NOT set `USER`. The entrypoint MUST start as root only long enough to prepare `/data`, then exec the server as uid 65532.

#### Scenario: Healthcheck follows the runtime port
- **WHEN** a reviewer reads the Dockerfile `HEALTHCHECK`
- **THEN** it calls `wget` on `http://127.0.0.1:${PORT}/health`
- **AND** the file contains no `$$`

### Requirement: Secrets stay in the environment
The Dockerfile MUST NOT declare a build argument whose name contains `JWT`, `PASSWORD`, `SECRET`, `API_KEY`, or `TOKEN`. The build context MUST exclude `.env` files and `morph-engi/backend/uploads`. A committed example env MUST leave `JWT_SECRET`, `MORPH_AI_API_KEY`, and `USERS_PANEL_BASE_URL` empty and MUST NOT contain a secret value.

#### Scenario: No secret build args
- **WHEN** a reviewer reads `morph-engi/Dockerfile`
- **THEN** it has no build argument for a JWT, password, secret, API key, or token

#### Scenario: Example env is empty
- **WHEN** a reviewer reads the Project production env example
- **THEN** `JWT_SECRET`, `MORPH_AI_API_KEY`, and `USERS_PANEL_BASE_URL` are present with empty values
- **AND** the file has no API key, password, or development JWT

### Requirement: SQLite and uploads live under /data
`MORPH_ENGI_UPLOAD_DIR`, when unset or blank, MUST resolve to the relative directory `uploads`. When set, every upload write and `serve_upload` MUST use that directory. The image MUST default `MORPH_ENGI_DATABASE_URL` to `sqlite:///data/morph_engi.db` and `MORPH_ENGI_UPLOAD_DIR` to `/data/uploads`. The entrypoint MUST create `/data/uploads` and give `/data` to uid 65532 before the server runs.

#### Scenario: Local uploads stay relative
- **WHEN** `MORPH_ENGI_UPLOAD_DIR` is unset or blank
- **THEN** the upload directory is `uploads`

#### Scenario: Configured uploads are used for reads and writes
- **WHEN** `MORPH_ENGI_UPLOAD_DIR` is `/data/uploads`
- **THEN** new uploads are stored under `/data/uploads/<org>/`
- **AND** `GET /api/v1/uploads/<org>/<file>` reads that same tree

### Requirement: Production rejects a dev secret and a loopback Morph base
When `APP_ENV` is `production`, the process MUST exit before it listens if `JWT_SECRET` is empty, shorter than 32 characters, `dev-morph-engi-secret`, or `morph-dev-jwt-secret-change-me`; if `USERS_PANEL_BASE_URL` is empty or its host is `localhost`, `127.0.0.1`, `::1`, or `0.0.0.0`; or if the SQLite file or the upload directory is not an absolute path under `/data`. Any other `APP_ENV` MUST keep the current development defaults and MUST still listen.

#### Scenario: Production refuses the development JWT
- **WHEN** `APP_ENV` is `production` and `JWT_SECRET` is `dev-morph-engi-secret` or unset
- **THEN** startup fails before the process listens
- **AND** the error names `JWT_SECRET`

#### Scenario: Production refuses a loopback Morph API
- **WHEN** `APP_ENV` is `production` and `USERS_PANEL_BASE_URL` is `http://127.0.0.1:9090`
- **THEN** startup fails before the process listens
- **AND** the error names `USERS_PANEL_BASE_URL`

#### Scenario: Development still starts with defaults
- **WHEN** `APP_ENV` is unset or `development` and `JWT_SECRET` is unset
- **THEN** the configured secret is `dev-morph-engi-secret`
- **AND** the upload directory defaults to `uploads`
- **AND** startup is not rejected for those defaults

### Requirement: Operators can build Project without Render
`morph-engi/README.md` MUST describe the image build, the healthcheck, `PORT`, the Morph API base, and the storage paths. The documented build command MUST NOT require Render or `render.yaml`.

#### Scenario: Local image build is documented
- **WHEN** an operator follows the Project container section with this repository and a container engine
- **THEN** they can build the image and request `/health` without creating a Render service
