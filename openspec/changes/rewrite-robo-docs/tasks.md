## 1. Root README

- [x] 1.1 Remove `landing/` from the project-folders table and drop links to `DEVELOPER_BASELINE.md`, `DEPLOY-README.md`, and `AI_ASSISTANT_MORPHAI_CONTRACT.md`
- [x] 1.2 Point Related docs at `docs/agents/00-architecture-overview.md`, `docs/agents/12-build-deploy.md`, and existing per-app READMEs; keep one-root-`.env`, ports, and optional Neo4j as they already are
- [x] 1.3 Confirm every markdown link in `README.md` resolves to a path that exists

## 2. Agent guides — core map

- [x] 2.1 Rewrite `docs/agents/00-architecture-overview.md` (live app map, ports, SQLite+Badger, leftover `mysql`/`mongo` package names, Morph AI chat, no Booki/Academi, no required MySQL/Mongo/Redis)
- [x] 2.2 Rewrite `docs/agents/01-auth-flow.md` for Morph JWT SSO (`userspanel_session_token`, legacy `USERS_PANEL_BASE_URL` → Morph `:9090`)
- [x] 2.3 Rewrite `docs/agents/02-ai-integration.md` for Morph AI + `pkg/morphai` / `pkg/morphai-rs`; no satellite `platform-chat` drawers
- [x] 2.4 Rewrite `docs/agents/13-conventions.md` (user-facing names vs folder/env ids, dark-only, one `.env`, Morph as auth hub without MySQL `plat_users` as required infra)

## 3. Agent guides — per product

- [x] 3.1 Rewrite `docs/agents/03-morph.md` as Morph AI + MorphNotes (drop Transfinder/school)
- [x] 3.2 Rewrite `docs/agents/04-formx.md` as Event Logs (SQLite+Badger; keep `formx`/`sheetx` ids)
- [x] 3.3 Rewrite `docs/agents/05-composerx.md` as Content Maker (SQLite+Badger; drop TranMail MySQL/Mongo/Redis)
- [x] 3.4 Replace `docs/agents/06-booki.md` with `docs/agents/06-ai-tools.md` for `bk/` (Assistants, RAG, Documents, System)
- [x] 3.5 Rewrite `docs/agents/07-rust-apps.md` as Data Access + Project (Morph SSO, no platform-chat, `SHARPREPORT_PORT`)
- [x] 3.6 Rewrite `docs/agents/09-morph-utils.md` (Event Logs not Survey Maker; MorphUtils; iframe URLs)
- [x] 3.7 Rewrite `docs/agents/10-shared-libraries.md` (drop Booki as a consumer)
- [x] 3.8 Rewrite `docs/agents/12-build-deploy.md` (local `start-all.sh` only remaining services; no MySQL/Mongo/Redis; no Booki/Academi; note missing `deploy/` and leftover `scripts/deploy.sh`; optional Vercel Project preview)

## 4. Remove ghost agent chapters

- [x] 4.1 Delete `docs/agents/08-academi.md` and `docs/agents/11-platform-chat.md` after folding the platform-chat one-liner into 00/02
- [x] 4.2 Delete `docs/agents/06-booki.md` if it still exists after 3.4

## 5. Per-app READMEs

- [x] 5.1 Replace `morph/README.md` Transfinder/school overview with Morph AI / MorphNotes and a pointer to `docs/agents/03-morph.md`
- [x] 5.2 Update `morph-utils/README.md` to Event Logs (not Survey Maker) and MorphUtils
- [x] 5.3 Update `formx/README.md` user-facing name to Event Logs; keep formx/SheetX as ids; do not require MySQL/Mongo
- [x] 5.4 Update `composerx/backend/README.md` (and frontend README if it still says MergeEmailX/MySQL) for Content Maker + root `.env` + SQLite/Badger
- [x] 5.5 Retitle `bk/README.md` to AI tools; drop Ground Control as the product name; keep run/ports that are still true
- [x] 5.6 Update `SharpReport/README.md` to Data Access, dark-only, Morph SSO, UI `:5178` / `SHARPREPORT_PORT`
- [x] 5.7 Align `morph-engi/README.md` wording to MorphUtils (no space) if needed; keep Vercel preview section

## 6. Archive and GraphRAG ops

- [x] 6.1 Create `docs/archive/README.md` stating contents are historical, not current operator docs
- [x] 6.2 Move `docs/superpowers/`, `docs/MORPH_GRAPH_RAG_PLAN.md`, and `docs/SURVEY_BOT_AND_GRAPHRAG_PLAN.md` into `docs/archive/`
- [x] 6.3 Rewrite `docs/MORPH_GRAPH_OPS.md` as optional Neo4j + worker (`TRAN_MYSQL_DSN` is a worker requirement, not default stack)

## 7. OpenSpec context and verify

- [x] 7.1 Fill `openspec/config.yaml` `context` with remaining apps, user-facing names, ports, SQLite+Badger, Morph JWT, dark-only, one `.env`
- [x] 7.2 Grep `README.md`, `docs/agents/`, and the per-app READMEs in this change for Booki, Academi, Survey Maker, Transfinder, Ground Control, DataPulse, `DEPLOY-README`, `DEVELOPER_BASELINE`, `AI_ASSISTANT_MORPHAI_CONTRACT`, and required MySQL/Mongo/Redis; fix remaining hits (archive may keep historical wording)
