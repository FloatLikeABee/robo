# Design

## Context

See proposal.md for motivation. The live stack is still Dorph, DorphUtils, Event Logs, Content Maker, Data Access, Project, and AI tools. Booki, Academi, UsersPanel, and satellite `platform-chat` are already out of that stack. Graphic documents is already out of AI tools Documents. This checkout has leftover files for those drops, untracked Morph rename copies, and a working-tree deletion of tracked Project (`dorph-engi/`). Data Access source (`SharpReport/`) is absent from git in this checkout; its Blueprint and specs still describe it.

## Goals / Non-Goals

**Goals:**

- Delete an explicit leftover inventory, then restore live Project from git.
- Keep Graphic document HTTP API and DorphUtils.
- Commit remaining local work and push `main` to GitHub `origin` without force-push.

**Non-Goals:**

- Repo-wide unused-file or dead-import sweep.
- Removing Data Access as a product, rewriting its specs, or inventing a `SharpReport/` tree.
- Dropping Project, GraphRAG worker, OpenSpec archives, or DorphUtils.
- Restoring a Files workspace or Graphic documents UI.

## Decisions

### Explicit leftover list, not a dead-file detector
A generic unused-file scan would delete tests, docs, skills, and optional workers. Cleanup is a named list: `platform-chat/`, Graphic documents page + unused frontend helper, unused Booki/Academi/UsersPanel icons, untracked `morph/`, `morph-utils/`, `morphgraph-worker/`.

**Alternative considered:** Automated unused-export tooling. Rejected — false positives on live products.

### Restore Project instead of committing the working-tree deletion
`dorph-engi/` is still a live product and still in git. The 61 deleted working-tree files are restored with `git restore dorph-engi`. Missing from disk is not a product drop.

**Alternative considered:** Commit the deletion and strip launcher/Blueprint entries. Rejected — that would drop Project without an owner decision.

### Leave Data Access Blueprint and specs alone
`SharpReport/` is already gone from this git tree. Inventing source or deleting the `sharpreport` Blueprint would either fake a product or silently drop Data Access. Out of scope.

**Alternative considered:** Remove `render.yaml` / `start-all.sh` Data Access entries. Rejected — specs still require them.

### Keep Graphic document API
The Documents UI already dropped Graphic documents. Delete the page and `generateGraphicDocument` frontend helper only. Keep `bk` `POST /graphic-document/generate`.

### Untracked Morph copies are disk leftovers
`morph/`, `morph-utils/`, and `morphgraph-worker/` are untracked rename leftovers (including nested `node_modules` and `.git.nested-backup`). Delete from disk. Do not git-rm trees that were never tracked.

### Push after cleanup, no force
`origin` is `https://github.com/FloatLikeABee/robo.git`. After leftover deletion and Project restore, commit remaining work (including this change) and `git push origin HEAD`. No `--force`. Do not skip hooks. Do not amend a published commit.

## Risks / Trade-offs

- [Project restore surprises an operator who meant to drop it] → Restore is the default because it is still in the stack; drop Project only in a later change if asked.
- [Data Access stays specified without source in this checkout] → Do not invent or delete the product here.
- [Deleting `platform-chat` breaks a forgotten import] → Grep already shows no remaining package.json dependents; apply re-checks before delete.
- [Push includes unrelated session work] → That is requested (“push all codes”); keep secrets (`.env`) untracked.

## Migration Plan

1. Restore `dorph-engi/` from git.
2. Delete the leftover inventory and Morph disk copies.
3. Drop leftover ignore/docs needles that still treat `platform-chat` as a remaining app, without rewriting Data Access or Project docs.
4. Run the leftover check (tree gone, page gone, API present, Project present).
5. Commit remaining work and push `main` to `origin`.

Rollback: revert the cleanup commit. GitHub history still has `platform-chat` and the Graphic documents page.
