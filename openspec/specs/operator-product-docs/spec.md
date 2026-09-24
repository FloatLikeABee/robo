# operator-product-docs Specification

## Purpose

Keep the operator and agent docs that describe MorphNotes, Research, and Invite Signup aligned with the running product so planners do not revive removed screens or miss live ones.

## Requirements

### Requirement: MorphNotes nav matches the live drawer
`README.md`, `docs/agents/00-architecture-overview.md`, `docs/agents/03-morph.md`, and `morph/README.md` SHALL describe MorphNotes navigation as Tasks, Timelines, Big notes, Research, and Generic data. `docs/agents/03-morph.md` SHALL map those labels to `/morphdata/case-tasks`, `/morphdata/timelines`, `/morphdata/big-notes`, `/morphdata/research`, and `/morphdata/generic-data`. Those documents SHALL NOT present Settings → Users or `/morphdata/configuration/users` as a MorphNotes destination.

#### Scenario: Reader looks up MorphNotes navigation
- **WHEN** a reader opens `docs/agents/03-morph.md` or `morph/README.md` to see MorphNotes modules
- **THEN** they find Tasks, Timelines, Big notes, Research, and Generic data
- **AND** they do not find Settings → Users as a nav item or a Users page route

#### Scenario: Old settings URLs are not a destination
- **WHEN** a reader searches the listed docs for where to administer users inside MorphNotes
- **THEN** the docs state that `/morphdata/settings` and `/morphdata/configuration` redirect to `/morphdata/generic-data`
- **AND** they point new-account provisioning at Invite Signup

### Requirement: Invite Signup port and flows are documented
The listed docs SHALL document Invite Signup as the `invite-signup/` app, launcher service `invite-signup-ui`, with the dev UI at `http://localhost:3051`. `docs/agents/03-morph.md` SHALL describe the unauthenticated redeem flow at `/redeem` and the admin flow at `/admin`, including the Morph API routes those screens call.

#### Scenario: Reader starts Invite Signup
- **WHEN** a reader looks up how to run Invite Signup from the root README or `docs/agents/12-build-deploy.md`
- **THEN** they see service name `invite-signup-ui` and URL `http://localhost:3051`
- **AND** they see that the service has no separate API process and no `start-all.sh` alias

#### Scenario: Reader follows redeem and admin
- **WHEN** a reader opens the Invite Signup section of `docs/agents/03-morph.md`
- **THEN** they see `/redeem` submitting `POST /api/invite/redeem` and showing a username and password once
- **AND** they see `/admin` signing in with a Morph admin account and creating codes via `POST /api/admin/invite-codes`

### Requirement: Research module and publish paths are documented
`docs/agents/03-morph.md` SHALL document the Research module at `/morphdata/research`, including how a finished job is published and where the public HTML is served.

#### Scenario: Reader publishes research
- **WHEN** a reader looks up Research publish behavior in `docs/agents/03-morph.md`
- **THEN** they see `POST /api/tran/research/:id/publish`
- **AND** they see the public HTML path `GET /api/tran/public/research/:slug`
- **AND** they see that a new job runs five verified online rounds

### Requirement: Build-deploy service table includes Invite Signup
The service table in `docs/agents/12-build-deploy.md` SHALL include `invite-signup-ui`. The opening product description in that file SHALL name invite-signup among the local services.

#### Scenario: Reader scans the service table
- **WHEN** a reader reads the service table in `docs/agents/12-build-deploy.md`
- **THEN** they find a row whose UI column is `invite-signup-ui`
- **AND** the API column for that row is empty

### Requirement: Morph AI workspace is not a Files tab
Where `docs/agents/03-morph.md` describes the Morph AI agent workspace, it SHALL name the tabs Notes & TODOs and Context & Knowledge. It SHALL NOT describe a Files tab or a local-folder Files workspace as a current Morph AI surface. It MAY describe the header AI tools control as the drawer that opens AI tools (`bk`).

#### Scenario: Reader looks for a Files workspace
- **WHEN** a reader reads the Morph AI overview in `docs/agents/03-morph.md`
- **THEN** they see Notes & TODOs and Context & Knowledge as the workspace tabs
- **AND** they do not see a Files workspace tab listed as available
