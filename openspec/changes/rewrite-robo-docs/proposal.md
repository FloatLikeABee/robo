## Why

The root README is closer to the live Morph stack than the rest of the docs, but it still links missing files (`DEVELOPER_BASELINE.md`, `DEPLOY-README.md`, `AI_ASSISTANT_MORPHAI_CONTRACT.md`) and lists a `landing/` folder that is gone. `docs/agents/` and several per-app READMEs still describe Booki, Academi, required MySQL/Mongo/Redis, Transfinder/school, Survey Maker, and satellite chat. Operators and agents cannot tell which documents are true.

## What Changes

- Rewrite root `README.md` so it matches `start-all.sh`: remaining products, user-facing names, stacks, ports, one root `.env`, Morph JWT SSO, Morph AI as system chat. Drop broken links. Do not present `landing/` or missing deploy markdown as live.
- Replace `docs/agents/` with a slim set that matches that map. Delete ghost product chapters (Booki, Academi). Fold the platform-chat stub into architecture. Drop “required MySQL · Mongo · Redis” and Transfinder/school copy.
- Slim per-app READMEs (`morph/`, `morph-utils/`, `formx/`, `composerx/`, others as needed) so they use the same names and storage story (SQLite + Badger locally; leftover `mysql`/`mongo` package names called out once).
- Document deploy as it actually is: local `start-all.sh`; optional Vercel static preview for Project; `scripts/deploy.sh` still mentions Render/Alibaba but the `deploy/` tree is missing — do not restore a fake `DEPLOY-README.md`. Data Access keeps `SharpReport/docs/` if it stays accurate.
- GraphRAG / Neo4j / `morphgraph-worker` stays **optional** (appendix + existing `docs/MORPH_GRAPH_OPS.md` rewritten to match). Dated plans (`docs/superpowers/`, `*_PLAN.md`) move to `docs/archive/` so they are not the operator path.
- Fill `openspec/config.yaml` `context` with the live stack so later changes do not reintroduce ghost apps.

## Capabilities

### New Capabilities

- `docs-match-live-stack`: README, agent guides, and per-app READMEs describe only remaining products and current frameworks; ghost apps and dead links are gone; GraphRAG is optional; historical plans are archived.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baseline for `current-stack-docs`)

## Impact

- Root `README.md`, `docs/agents/*`, selected per-app READMEs, `docs/MORPH_GRAPH_OPS.md`, `openspec/config.yaml`.
- Move `docs/superpowers/` and GraphRAG/survey *PLAN* markdown into `docs/archive/`.
- Delete `docs/agents/06-booki.md` and `docs/agents/08-academi.md` (and the platform-chat numbered chapter after folding).
- Docs only. No application code, no `start-all.sh` behavior, no restoring `deploy/` or deleted apps.
- Out of scope: rewriting `scripts/deploy.sh` logic, migrating GraphRAG off `TRAN_MYSQL_DSN`, committing leftover app deletions.
