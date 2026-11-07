# Proposal

## Why

The checkout still carries files from products and UI modules the stack already dropped (`platform-chat`, Graphic documents page, Booki/Academi/UsersPanel icons, leftover Morph rename copies). Those leftovers invite revival and bloat the tree we push to GitHub. A repo-wide “delete anything unused” sweep would also take tests, docs, optional workers, and live apps that are only missing from this working tree.

## What Changes

- **BREAKING (leftovers only):** Delete the unused `platform-chat/` package and remaining ignore/docs mentions of it as a shippable tree. Dorph AI stays the only chat.
- **BREAKING (AI tools UI):** Delete the Graphic documents page and its unused frontend client helper. Keep the Graphic document HTTP API.
- Delete unused Booki, Academi, and UsersPanel icons that no live UI references.
- Delete untracked leftover Morph rename copies (`morph/`, `morph-utils/`, `morphgraph-worker/`) from disk. They are not git-tracked.
- Restore Project (`dorph-engi/`) from git. It is still a live product; the working tree is missing it.
- Leave DorphUtils, GraphRAG worker, OpenSpec history, skills, tests, and Data Access Blueprint/docs in place. Data Access source (`SharpReport/`) is already absent from this checkout; this change does not invent it or drop the product.
- Do not delete a file only because no import was found this week.
- After the cleanup, commit the remaining local work and push `main` to GitHub `origin`.

## Capabilities

### New Capabilities

- `dropped-module-leftovers`: Files that exist only for a product or UI module already dropped from the live stack are gone. Live products stay, including trees that are only missing from the working copy.

### Modified Capabilities

- (none)

## Impact

- Delete `platform-chat/` and drop leftover ignore/docs needles that treat it as a remaining app.
- Delete `bk/frontend/src/pages/GraphicDocumentGenerator.js` and unused `generateGraphicDocument` frontend helper. Keep `bk` `/graphic-document/generate`.
- Delete unused `booki-icon.svg`, `academi-icon.svg`, and `userspanel-icon.svg` under Dorph and DorphUtils public icons.
- Restore `dorph-engi/` from git. Do not delete `dorph-utils/`, `dorphgraph-worker/`, or OpenSpec archives.
- Remove untracked Morph rename leftover directories from disk.
- GitHub: `origin` is `https://github.com/FloatLikeABee/robo.git`. Push `main` after cleanup. No force-push.
