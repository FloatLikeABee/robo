## Purpose

Keeps operator README, Cursor agent guides, and per-app READMEs aligned with the apps and frameworks that actually run from `start-all.sh`, so readers do not follow ghost products or missing files.

## ADDED Requirements

### Requirement: Root README matches the live stack
The root README MUST describe only remaining products and folders: Morph AI / MorphNotes (`morph/`), MorphUtils (`morph-utils/`), Event Logs (`formx/`), Content Maker (`composerx/`), Project (`morph-engi/`), Data Access (`SharpReport/`), AI tools (`bk/`), shared `pkg/`, and optional GraphRAG (`morphgraph-worker/`). It MUST use those user-facing names. It MUST NOT present Booki, Academi, UsersPanel, morph-broadcast, or `landing/` as live apps. It MUST NOT link files that are absent from the repo (`DEVELOPER_BASELINE.md`, `DEPLOY-README.md`, `AI_ASSISTANT_MORPHAI_CONTRACT.md`). Related-docs MUST point at existing paths (agent guides and per-app READMEs). Local config MUST remain one repo-root `.env`. Prerequisites MUST NOT require MySQL, MongoDB, or Redis for the default stack.

#### Scenario: README has no dead links or ghost products
- **WHEN** an operator opens the root README
- **THEN** the project-folders table does not list Booki, Academi, UsersPanel, morph-broadcast, or landing as live products
- **AND** every markdown link in that README resolves to a file or directory that exists
- **AND** the environment section still instructs `cp .env.example .env` at the repo root

### Requirement: Agent guides match the live stack
`docs/agents/` MUST describe the same remaining products, Morph JWT SSO, Morph AI as the system chat (no satellite `platform-chat` drawer), and local storage as SQLite + Badger (plus in-process cache) under each app’s `./data/`. Architecture and build guides MUST NOT list Booki or Academi as current apps, MUST NOT list MySQL/Mongo/Redis as required local infrastructure, and MUST NOT use Transfinder/school as the Morph product story. Ghost chapters for Booki and Academi MUST NOT remain as live agent guides. Folder and env ids (`formx`, `sheetx`, `USERS_PANEL_BASE_URL`) MAY appear when labeled as implementation ids, not as user-facing names.

#### Scenario: Architecture overview has no Booki or required MySQL
- **WHEN** an agent reads `docs/agents/00-architecture-overview.md`
- **THEN** the app map and stack table do not include Booki or Academi as live apps
- **AND** required local data stores are SQLite and Badger, not MySQL, MongoDB, or Redis
- **AND** user-facing module names are Event Logs, Content Maker, Data Access, and Project

### Requirement: Per-app READMEs use current names and storage
The READMEs for `morph/`, `morph-utils/`, `formx/`, `composerx/` (backend at least), `bk/`, `SharpReport/`, and `morph-engi/` MUST use the user-facing product names and MUST NOT tell operators to install MySQL, MongoDB, or Redis for local run. Morph MUST NOT be documented as a Transfinder Form/Report Assistant or school SQL-Server product. MorphUtils MUST list Event Logs (not Survey Maker). AI tools MUST be named AI tools (not Ground Control) in the operator-facing title. Data Access MUST be named Data Access (not DataPulse) and MUST NOT advertise a light/dark theme switch.

#### Scenario: Morph README is Morph AI and MorphNotes
- **WHEN** an operator opens `morph/README.md`
- **THEN** the title and overview describe Morph AI and MorphNotes
- **AND** they are not told the product is Transfinder or a student/SQL Server registration system

### Requirement: Deploy docs state what exists
Operator docs MUST present local development via `start-all.sh` as the supported path. Optional Project Vercel static preview MAY be documented if `morph-engi/README.md` still describes it. Docs MUST NOT claim a complete Render/Alibaba `deploy/` tree or `DEPLOY-README.md` while those files are absent. `scripts/deploy.sh` MAY be mentioned as leftover/incomplete. GraphRAG / Neo4j MUST be documented as optional. Dated plan docs under `docs/superpowers/` and `docs/*_PLAN.md` MUST NOT remain on the operator/agent path; they MUST live under `docs/archive/` if retained.

#### Scenario: README does not promise missing deploy markdown
- **WHEN** an operator follows the root README for production
- **THEN** they are not sent to `DEPLOY-README.md`
- **AND** they are not told they must run Neo4j to start the rest of the stack
