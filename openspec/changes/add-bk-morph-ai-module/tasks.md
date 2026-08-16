## 1. Morph AI header — BK module

- [x] 1.1 Add `REACT_APP_BK_URL` to `morph/frontend` env example (default `http://localhost:3000`)
- [x] 1.2 Add `bk-icon.svg` under `morph/frontend/public/icons/`
- [x] 1.3 Extend `SkoolAiChat.js`: `APP_URLS.bk`, `HEADER_APP_ICONS.bk`, and third `headerAppLinks` entry after Morph Utils with `appHrefWithSession`
- [x] 1.4 Smoke: from Morph AI (`:3031`), BK link opens `:3000` with `userspanel_token` when signed in

## 2. start-all.sh — BK services

- [x] 2.1 Add `bk-api` and `bk-ui` to `ALL_SERVICES` (after morph-utils-ui block)
- [x] 2.2 Implement `start_one` cases: `bk-api` → `python main.py` in `bk/`; `bk-ui` → `npm start` in `bk/frontend`
- [x] 2.3 Add `bk` alias in `resolve_services`; `service_url` entries for both services
- [x] 2.4 Update `print_list` and `do_install` (`bk/frontend` npm install; note Python venv in comment or docs)
- [x] 2.5 Smoke: `./start-all.sh start bk` and `./start-all.sh status`

## 3. start-all.sh — Neo4j ensure

- [x] 3.1 Add `ensure_neo4j()` helper (check port 7687, `neo4j start` if needed, warn on failure)
- [x] 3.2 Call `ensure_neo4j` at start of `start_all`, `restart_all`, and full `start` paths
- [x] 3.3 Smoke: with Neo4j stopped and CLI installed, full start brings bolt port up or logs warning

## 4. Documentation

- [x] 4.1 Update `docs/agents/00-architecture-overview.md` and `12-build-deploy.md` with BK ports (8000/3000) and Neo4j note
- [x] 4.2 Update root `README.md` / `start-all.sh` header comment to mention `bk` alias and Neo4j
