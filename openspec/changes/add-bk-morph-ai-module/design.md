## Context

Morph AI (`morph/frontend`) exposes two header shortcuts today: **Morph Data** (same-origin `/morphdata`) and **Morph Utils** (`REACT_APP_MORPH_UTILS_URL`, default `http://localhost:3040`) with `userspanel_token` appended via `appHrefWithSession()`.

BK (`bk/`) is a Python FastAPI backend (`main.py`, default **8000**) plus a Create React App frontend (`bk/frontend`, default **3000**, proxies to API). It is not in `start-all.sh`’s `ALL_SERVICES` list.

Neo4j is used by Morph’s GraphRAG path (`morph/handlers/knowledge.go`, env `NEO4J_*`). Ops docs recommend `brew install neo4j && neo4j start` (`docs/MORPH_GRAPH_OPS.md`). BK itself does not require Neo4j (ChromaDB), but the unified dev stack should start Neo4j for Morph graph features.

## Goals / Non-Goals

**Goals:**

- One-click navigation from Morph AI chat to BK with SSO token passthrough.
- `./start-all.sh` starts BK API + UI with the same pid/log patterns as other apps.
- `./start-all.sh` attempts to start Neo4j when bolt `127.0.0.1:7687` is not accepting connections.
- Consistent naming: header label **BK**, services `bk-api` / `bk-ui`, alias `bk`.

**Non-Goals:**

- Embedding BK inside Morph Utils iframes (header link only for now).
- BK auth refactor to native UsersPanel JWT validation (token query param is sufficient for v1 if BK already supports or ignores it).
- Docker-based Neo4j orchestration (Homebrew/native `neo4j` CLI only).
- Changing BK’s internal ports away from 8000/3000.

## Decisions

### 1. Header link placement and URL

- **Decision**: Add `{ id: 'bk', label: 'BK', href: appHrefWithSession(BK_URL) }` after `morphutils` in `headerAppLinks`.
- **Default URL**: `REACT_APP_BK_URL` env, fallback `http://localhost:3000`.
- **Rationale**: Matches Morph Utils pattern; BK UI is the user-facing entry (API is proxied by CRA).
- **Alternatives**: Link to API docs (`:8000/docs`) — rejected; not the product UI.

### 2. Icon

- **Decision**: Add `morph/frontend/public/icons/bk-icon.svg` (simple monogram or reuse a neutral agent icon); reference in `HEADER_APP_ICONS.bk`.
- **Alternatives**: Emoji fallback — rejected for visual consistency with Morph Data / Utils.

### 3. start-all.sh service wiring

- **bk-api**: `start_service bk-api "${ROOT}/bk" python main.py` with `load_app_env` on `bk/`.
- **bk-ui**: `start_service bk-ui "${ROOT}/bk/frontend" npm start` (CRA binds 3000).
- **Order**: Insert after `morph-utils-ui` in `ALL_SERVICES` (UI services grouped; API can start before UI — place `bk-api` before `bk-ui`, after `morph-engi-api` or near other APIs).
- **Alias**: `bk` → `bk-api bk-ui`.
- **Install**: `pip install -r requirements.txt` is heavy; `do_install` adds `cd bk/frontend && npm install` only (same as other frontends). Document optional venv for Python deps.

### 4. Neo4j ensure helper

- **Decision**: New function `ensure_neo4j()` called once at the beginning of `start_all`, `restart_all`, and `start` (all services):
  1. If `lsof -tiTCP:7687 -sTCP:LISTEN` succeeds → ok, return.
  2. Else if `command -v neo4j` → run `neo4j start`, sleep 2s, re-check port; warn on failure.
  3. Else warn: “Neo4j not installed; Morph graph features may be unavailable.”
- **Rationale**: Non-fatal — devs without graph enabled can still work.
- **Alternatives**: Docker compose — out of scope; fail hard if Neo4j missing — too strict for optional feature.

### 5. Session token for BK

- **Decision**: Pass `userspanel_token` on the BK URL; document that BK should accept or ignore until full SSO is wired.
- **If BK has no handler**: Still ship the link; follow-up task to add token consumption in BK (out of scope unless trivial).

## Risks / Trade-offs

- **[Port 3000 conflict]** BK UI defaults to 3000, which may clash with other CRA apps → Mitigation: document `PORT=3001` in `bk/frontend` if needed; `free_listening_port` optional for 3000 on bk-ui start.
- **[Neo4j start permissions]** `neo4j start` may require user approval on macOS → Mitigation: warn and continue.
- **[Python venv]** `python main.py` without venv may miss deps → Mitigation: tasks include checking `bk/requirements.txt` install note in README.
- **[CORS]** Opening BK from Morph AI origin → BK already allows `localhost:3000`; Morph AI is `:3031` — may need `CORS_ORIGINS` update if API called cross-origin (proxy usually handles).

## Migration Plan

1. Ship Morph AI header + env example.
2. Extend `start-all.sh`; verify `./start-all.sh start bk` and full stack.
3. Update `docs/agents/00-architecture-overview.md` and `12-build-deploy.md` port table.
4. Rollback: remove header entry and start-all cases (no data migration).

## Open Questions

- Does BK need explicit `userspanel_token` handling on load, or is v1 link-only acceptable? (Assume link-only unless BK already reads the param.)
