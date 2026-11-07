# Tasks

## 1. Restore live Project

- [x] 1.1 Restore `dorph-engi/` from git (`git restore dorph-engi`) and verify `dorph-engi/README.md` exists on disk and `git status --short dorph-engi` is empty of deletions.

## 2. Delete dropped leftovers

- [x] 2.1 Delete the `platform-chat/` tree and leftover ignore/docs needles that still treat it as a remaining app. Verify no top-level `platform-chat/` directory and no `@robo/platform-chat` package.json dependency.
- [x] 2.2 Delete `bk/frontend/src/pages/GraphicDocumentGenerator.js` and the unused `generateGraphicDocument` frontend helper. Verify Documents still mounts Articles and ScholarForge, no UI import of that page, and `bk/src/api.py` still serves `POST /graphic-document/generate`.
- [x] 2.3 Delete unused `booki-icon.svg`, `academi-icon.svg`, and `userspanel-icon.svg` under Dorph and DorphUtils public icons. Verify those filenames are gone from those icon directories.
- [x] 2.4 Delete untracked leftover Morph rename copies `morph/`, `morph-utils/`, and `morphgraph-worker/` from disk. Verify those directories are gone and `dorph-utils/` remains.

## 3. Leftover check

- [x] 3.1 Add a small leftover check (assert-based test or one script) that fails if `platform-chat/` exists, Graphic documents page exists, Graphic document API is missing, `dorph-utils/` is missing, or `dorph-engi/README.md` is missing. Verify that check passes.

## 4. Push to GitHub

- [ ] 4.1 Stage remaining project work plus this cleanup (never `.env` or secrets). Commit on `main` with a message covering leftover deletion, Project restore, and the unpushed session work. Verify `git status` is clean except ignored files.
- [ ] 4.2 Push `main` to `origin` (`https://github.com/FloatLikeABee/robo.git`) without `--force`. Verify `git status -sb` shows `main` in sync with `origin/main`.

## Workflow follow-up

- Archive the change after review with `/opsx:archive`.
