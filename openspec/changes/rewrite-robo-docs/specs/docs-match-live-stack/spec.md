## Purpose

Keeps operator README, Cursor agent guides, and per-app READMEs aligned with the apps and frameworks that actually run from `start-all.sh`, so readers do not follow ghost products or missing files.

## ADDED Requirements

### Requirement: Root README matches the live stack
The root README MUST describe only remaining products and folders: Dorph AI / DorphNotes (`dorph/`), DorphUtils (`dorph-utils/`), Event Logs (`formx/`), Content Maker (`composerx/`), Project (`dorph-engi/`), Data Access (`SharpReport/`), AI tools (`bk/`), shared `pkg/`, and optional GraphRAG (`dorphgraph-worker/`). It MUST use those user-facing names. It MUST NOT present Booki, Academi, UsersPanel, dorph-broadcast, or `landing/` as live apps. It MUST NOT link files that are absent from the repo (`DEVELOPER_BASELINE.md`, `DEPLOY-README.md`, `AI_ASSISTANT_DORPHAI_CONTRACT.md`). Related-docs MUST point at existing paths (agent guides and per-app READMEs). Local config MUST remain one repo-root `.env`. Prerequisites MUST NOT require MySQL, MongoDB, or Redis for the default stack.

#### Scenario: README has no dead links or ghost products
- **WHEN** an operator opens the root README
- **THEN** the project-folders table does not list Booki, Academi, UsersPanel, dorph-broadcast, or landing as live products
- **AND** every markdown link in that README resolves to a file or directory that exists
- **AND** the environment section still instructs `cp .env.example .env` at the repo root

### Requirement: Agent guides match the live stack
`docs/agents/` MUST describe the same remaining products, Dorph JWT SSO, Dorph AI as the system chat (no satellite `platform-chat` drawer), and local storage as SQLite + Badger (plus in-process cache) under each app’s `./data/`. Architecture and build guides MUST NOT list Booki or Academi as current apps, MUST NOT list MySQL/Mongo/Redis as required local infrastructure, and MUST NOT use Transfinder/school as the Dorph product story. Ghost chapters for Booki and Academi MUST NOT remain as live agent guides. Folder and env ids (`formx`, `sheetx`, `USERS_PANEL_BASE_URL`) MAY appear when labeled as implementation ids, not as user-facing names.

#### Scenario: Architecture overview has no Booki or required MySQL
- **WHEN** an agent reads `docs/agents/00-architecture-overview.md`
- **THEN** the app map and stack table do not include Booki or Academi as live apps
- **AND** required local data stores are SQLite and Badger, not MySQL, MongoDB, or Redis
- **AND** user-facing module names are Event Logs, Content Maker, Data Access, and Project

### Requirement: Per-app READMEs use current names and storage
The READMEs for `dorph/`, `dorph-utils/`, `formx/`, `composerx/` (backend at least), `bk/`, `SharpReport/`, and `dorph-engi/` MUST use the user-facing product names and MUST NOT tell operators to install MySQL, MongoDB, or Redis for local run. Dorph MUST NOT be documented as a Transfinder Form/Report Assistant or school SQL-Server product. DorphUtils MUST list Event Logs (not Survey Maker). AI tools MUST be named AI tools (not Ground Control) in the operator-facing title. Data Access MUST be named Data Access (not DataPulse) and MUST NOT advertise a light/dark theme switch.

#### Scenario: Dorph README is Dorph AI and DorphNotes
- **WHEN** an operator opens `dorph/README.md`
- **THEN** the title and overview describe Dorph AI and DorphNotes
- **AND** they are not told the product is Transfinder or a student/SQL Server registration system

### Requirement: Deploy docs state what exists
Operator docs MUST present local development via `start-all.sh` as the supported path. Optional Project Vercel static preview MAY be documented if `dorph-engi/README.md` still describes it. Docs MUST NOT claim a complete Render/Alibaba `deploy/` tree or `DEPLOY-README.md` while those files are absent. `scripts/deploy.sh` MAY be mentioned as leftover/incomplete. GraphRAG / Neo4j MUST be documented as optional. Dated plan docs under `docs/superpowers/` and `docs/*_PLAN.md` MUST NOT remain on the operator/agent path; they MUST live under `docs/archive/` if retained.

#### Scenario: README does not promise missing deploy markdown
- **WHEN** an operator follows the root README for production
- **THEN** they are not sent to `DEPLOY-README.md`
- **AND** they are not told they must run Neo4j to start the rest of the stack
