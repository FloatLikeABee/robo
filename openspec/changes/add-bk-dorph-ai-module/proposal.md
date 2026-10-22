## Why

BK (Ground Control RAG / agent workspace) lives in the monorepo but is not reachable from Dorph AI’s app header or the shared `start-all.sh` dev launcher. Developers must start BK and Neo4j manually, and Dorph graph features that depend on Neo4j fail silently when the database is down.

## What Changes

- Add **BK** as a third header app link in Dorph AI, positioned **after Dorph Utils** (order: Dorph Data → Dorph Utils → BK).
- Pass the shared UsersPanel session token to BK the same way Dorph Utils does (`?userspanel_token=…`), so SSO works from Dorph AI.
- Register **bk-api** (FastAPI, default port 8000) and **bk-ui** (React, default port 3000) in `start-all.sh` with aliases `bk`.
- Before starting the dev stack, **ensure Neo4j is running** (start via `neo4j start` when the bolt port is not listening and the CLI is available); log a clear warning if Neo4j cannot be started.
- Document BK ports and env vars in agent docs / README where other apps are listed.

## Capabilities

### New Capabilities

- `dorph-ai-bk-module`: Dorph AI header exposes BK after Dorph Utils with session-aware deep link.
- `robo-dev-bk-services`: `start-all.sh` starts and manages BK API + UI alongside other robo apps.
- `robo-dev-neo4j-start`: Dev launcher ensures Neo4j is up (or warns) before dependent services start.

### Modified Capabilities

- (none — no existing specs under `openspec/specs/`)

## Impact

- **dorph/frontend**: `SkoolAiChat.js` header app links, optional `REACT_APP_BK_URL`, BK icon under `public/icons/`.
- **start-all.sh**: `ALL_SERVICES`, `start_one`, `resolve_services`, `service_url`, `do_install`, `print_list`; new `ensure_neo4j` helper.
- **bk/**: No product changes required for v1; may need env note for CORS/origin if Dorph AI opens BK on a different host.
- **docs/agents**: Port table and architecture notes for BK + Neo4j.
- **Neo4j**: Local dev dependency when `DORPH_GRAPH_ENABLED` or graph sync is used (Dorph backend).
