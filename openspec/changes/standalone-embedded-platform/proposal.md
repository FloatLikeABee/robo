## Why

Morph already moved toward embedded storage, but Utils apps (ComposerX, FormsX/SheetX), Booki, and other AI tools still hard-require networked **MySQL**, **MongoDB**, and **Redis**. That undoes the standalone goal and recently forced reinstalling those servers just to clear Morph Utils 500s. The platform should run from clone → build → run with on-disk stores only, keep Neo4j as the AI graph of record (async sync for uploads and durable content), and give Morph AI a first-class **skills store** so users can upload and reuse many skills.

## What Changes

- **BREAKING (ops)**: **Delete / stop and decommission** local MySQL, MongoDB, and Redis for this project (first implementation step). No Morph AI, Morph Data, Morph Utils, or AI-tool service may require them to boot.
- Extend the Morph embedded pattern **project-wide**: every service that still uses MySQL/Mongo/Redis MUST switch to:
  - **Relational**: embedded SQL in Go (`modernc.org/sqlite` preferred, consistent with Morph)
  - **Documents / blobs metadata**: **BadgerDB**
  - **Ephemeral cache / sessions**: **patrickmn/go-cache** (or thin wrapper)
- Cover **Morph AI**, **Morph Data**, **Morph Utils** embeds (ComposerX, SheetX/FormsX, DataX/SharpReport as needed, Morph Engi/Projects), and shared AI tool backends that currently open MySQL/Mongo/Redis.
- Provide migrations from legacy MySQL/Mongo where data still exists; empty/fresh embeds are acceptable when sources were already wiped.
- **Neo4j async sync**: durable uploads and store writes (files, published pages, survey assets, knowledge docs, skills, etc.) enqueue an async path that upserts/mirrors content into Neo4j for AI retrieval—without blocking the HTTP request on Neo4j availability when possible.
- **AI skills store**: platform capability for many skills (bundled + user-uploaded), discoverable by Morph AI assistants, with upload/list/get/delete (and optional enable/disable) APIs and UI entry points.
- Update `start-all.sh`, `.env.example` files, and docs so the default path is standalone (no brew MySQL/Mongo/Redis).

### Out of scope

- Replacing Neo4j itself (Neo4j remains the graph/AI index).
- Multi-node shared-file HA for SQLite/Badger (single-process / file-local by design).
- Rewriting every product UI; prefer stable HTTP APIs with storage swapped underneath.

## Capabilities

### New Capabilities

- `platform-embedded-storage`: Project-wide rule that Morph AI / Morph Data / Utils / AI tools persist with SQLite + Badger + go-cache only; MySQL/Mongo/Redis are removed as runtime dependencies.
- `neo4j-async-ingest`: Async mirroring of uploads and durable content into Neo4j for AI use, with retries and non-blocking API writes.
- `ai-skills-store`: Morph AI skills library—many skills, user upload, list/retrieve, and assistant consumption.

### Modified Capabilities

- (none — Morph embedded storage remains defined by the prior `morph-embedded-dbs` change; this change’s `platform-embedded-storage` extends the standalone guarantee across Utils and AI tools.)

## Impact

- **Ops**: Uninstall/stop Homebrew MySQL/Mongo/Redis used for this monorepo; remove DSNs from app `.env` / `.env.example`.
- **composerx/backend**, **formx/backend**, **booki/backend** (and any other Utils/AI tools still on MySQL/Mongo/Redis): storage rewrite + schema init on SQLite/Badger; drop networked drivers when unused.
- **morph/**: Finish cleanup of remaining MySQL/Mongo/Redis deps and env defaults; keep SQLite + Badger + go-cache as source of truth.
- **pkg/morphgraph / Neo4j workers**: New or extended async outbox/ingest for uploads and skills.
- **Morph AI frontend / assistants**: Skills store APIs and upload UX; assistants load skills from the store.
- **start-all.sh / docs**: Standalone boot instructions; no requirement to start MySQL/Mongo/Redis.
