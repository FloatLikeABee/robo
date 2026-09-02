## Context

See proposal.md for why. The root README already lists the remaining products; `docs/agents/` and several per-app READMEs do not. `start-all.sh` `ALL_SERVICES` is the source of truth for what runs: morph, formx, composerx, morph-engi, bk, SharpReport, morph-utils. Local data is SQLite + Badger under `./data/` even where Go packages are still named `mysql` / `mongo`. `deploy/` and `landing/` are absent; `scripts/deploy.sh` still lists Booki and points at missing `DEPLOY-README.md`. This change is documentation only.

User-facing names (do not “fix” URL ids): Morph AI, MorphNotes, MorphUtils, Event Logs (`formx`, embed id `sheetx`, route `/survey-bot`), Content Maker (`composerx`), Data Access (`SharpReport`, iframe `/datax`), Project (`morph-engi`), AI tools (`bk`). Auth: Morph JWT; env still `USERS_PANEL_BASE_URL` → Morph `:9090`. Chat: Morph AI only. Product UIs: dark-only.

## Goals / Non-Goals

**Goals:**

- One operator front door (root README) with no broken links.
- A short `docs/agents/` set Cursor can trust.
- Per-app READMEs that match those names and the SQLite/Badger story.
- Historical plans off the default reading path.
- `openspec/config.yaml` context so later proposals do not reintroduce ghost apps.

**Non-Goals:**

- Changing application code, `start-all.sh`, or `scripts/deploy.sh` behavior.
- Restoring `deploy/`, `landing/`, `DEVELOPER_BASELINE.md`, or `AI_ASSISTANT_MORPHAI_CONTRACT.md`.
- Migrating GraphRAG off `TRAN_MYSQL_DSN`.
- Rewriting SharpReport API/deployment internals beyond naming if those files stay accurate.
- Deleting unused product folders (UsersPanel/academi/booki) — that is separate WIP.

## Decisions

### 1. Keep `docs/agents/` numbered files; rewrite in place

**Choice:** Keep the path prefix `docs/agents/NN-*.md` so existing skills/rules that mention those files still resolve. Rewrite contents. Delete ghost chapters. Reuse the Booki slot for AI tools.

**Target set:**

| File | After |
|------|--------|
| `00-architecture-overview.md` | Live map, ports, SQLite+Badger, Morph AI chat |
| `01-auth-flow.md` | Morph JWT, cookie `userspanel_session_token`, `USERS_PANEL_BASE_URL` as legacy name |
| `02-ai-integration.md` | Morph AI + `pkg/morphai` / `pkg/morphai-rs`; no satellite drawers |
| `03-morph.md` | Morph AI + MorphNotes (Tasks, Timelines, Big notes, Generic data, Settings) |
| `04-formx.md` | Event Logs; SQLite+Badger; leftover package names once |
| `05-composerx.md` | Content Maker; same storage story |
| `06-ai-tools.md` | Replaces Booki chapter; `bk/` Assistants, RAG, Documents, System |
| `07-rust-apps.md` | Data Access + Project |
| `09-morph-utils.md` | MorphUtils iframe shell |
| `10-shared-libraries.md` | Drop Booki consumers |
| `12-build-deploy.md` | Local `start-all`; honest “no deploy/ tree”; Vercel Project preview pointer |
| `13-conventions.md` | Names, dark-only, one `.env`, ids vs labels |
| **Delete** | `06-booki.md`, `08-academi.md`, `11-platform-chat.md` (fold into 00/02) |

**Alternatives considered:** Flatten to `docs/architecture.md` + `docs/apps/*.md`. Cleaner, but breaks links and OpenSpec/agent habits. In-place rewrite is less churn.

### 2. README related-docs point at 00 + 12, not missing contracts

**Choice:** Do not recreate `DEVELOPER_BASELINE.md` or `AI_ASSISTANT_MORPHAI_CONTRACT.md`. Auth and AI live in `01` and `02`. Deploy lives in `12` plus `morph-engi` Vercel notes and `SharpReport/docs/DEPLOYMENT.md` if still accurate.

**Alternatives considered:** Restore the three missing files as stubs. Rejected — stubs would rot like the current dead links.

### 3. GraphRAG is an appendix, not required infra

**Choice:** README already says Neo4j is optional; keep that. Rewrite `docs/MORPH_GRAPH_OPS.md` so it is clearly optional and still documents `TRAN_MYSQL_DSN` as a **worker** requirement (code truth), not a platform default. Move `docs/MORPH_GRAPH_RAG_PLAN.md`, `docs/SURVEY_BOT_AND_GRAPHRAG_PLAN.md`, and `docs/superpowers/` to `docs/archive/`.

**Alternatives considered:** Delete plans outright. Archive keeps git-navigable history for people who want the old design without polluting `docs/agents/`.

### 4. Per-app README depth

**Choice:** Replace lying overviews; do not rewrite every API table. Targets:

- `morph/README.md` — replace Transfinder novel with Morph AI / MorphNotes + pointer to `docs/agents/03`.
- `morph-utils/README.md` — Event Logs not Survey Maker.
- `formx/README.md` — user-facing Event Logs; keep folder/env as formx/SheetX ids.
- `composerx/backend/README.md` — drop MySQL/Mongo/Redis prerequisites; Content Maker + root `.env`.
- `bk/README.md` — retitle AI tools; keep useful run/API bits, drop Ground Control as the product name.
- `SharpReport/README.md` — Data Access, dark-only, Morph SSO, `SHARPREPORT_PORT` / UI `:5178`.
- `morph-engi/README.md` — already close; MorphUtils (no space), keep Vercel section.

**Alternatives considered:** Delete per-app READMEs and point only at agents. Operators still clone into a single folder; keep a short README there.

### 5. Leftover `mysql` / `mongo` package names

**Choice:** One paragraph in `00` and `04`/`05`: packages may still be named `mysql`/`mongo` while `main` opens SQLite/Badger. Do not document installing those servers.

### 6. `openspec/config.yaml` context

**Choice:** Add a short `context:` block: remaining apps, names, ports, SQLite+Badger, Morph JWT, dark-only, one `.env`. Future `/opsx-propose` runs will pick this up.

## Risks / Trade-offs

- **[Risk]** Docs describe the intended working tree while some ghost folders may still exist on `origin/main` until other WIP is committed → **Mitigation:** Document the **launcher** stack (`start-all.sh`), not every directory that might still be on disk. Do not claim UsersPanel/academi/booki are gone from git if they are only deleted locally; say they are not part of the supported stack.
- **[Risk]** Archive path is still discoverable and can confuse search → **Mitigation:** Add `docs/archive/README.md` stating these are historical and not current.
- **[Risk]** Slimming `bk/README.md` and `morph/README.md` drops still-true API detail → **Mitigation:** Keep run commands and ports; move or drop only Transfinder/Ground Control product fiction.
- **[Trade-off]** Not fixing `scripts/deploy.sh` leaves a script that lists Booki. Accepted: out of scope; `12` will say the script is leftover until `deploy/` exists.

## Migration Plan

1. Rewrite README and agent guides (no runtime).
2. Slim per-app READMEs.
3. Move plans to `docs/archive/`; delete ghost agent chapters.
4. Update `openspec/config.yaml` context.
5. Grep the remaining markdown under `README.md`, `docs/agents/`, and per-app READMEs for Booki, Academi, Survey Maker, Transfinder, Ground Control, DataPulse, required MySQL/Mongo/Redis, and the three missing filenames.
6. Rollback: revert the doc commits; no data migration.

## Open Questions

None that change specs or tasks. Deploy cloud topology stays “not currently documented” until someone restores `deploy/`.
