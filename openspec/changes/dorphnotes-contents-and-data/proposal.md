# Proposal

## Why

Two separate DorphNotes nav items (Generic data and Content Maker) hide that imported files, AI analysis, and public HTML are one job. The Dorph AI header title also clips a purple gradient onto a purple glass bar, so “DORPH AI” reads as dark ink. The owner wants one **Contents & Data** place and a lighter title.

## What Changes

- Rename the DorphNotes nav label **Content Maker** to **Contents & Data**. Remove the separate **Generic data** drawer row.
- One DorphNotes destination hosts (1) the existing Generic data import/grid/analyze UI and (2) the existing Content Maker (composerx) publish UI.
- From a Generic data record, the operator can **publish analysis (and a data excerpt) as HTML** through the existing Content Maker publish API. No second HTML product.
- Redirect `/dorphdata/generic-data` and keep `/dorphdata/content-maker` as aliases of the combined page.
- Dorph AI `.app-title` uses light text (`--chat-text`) instead of clipped accent-on-accent. The 4px accent bar stays.
- **Not in scope**: renaming the `composerx` folder, deleting `/api/tran/generic-data`, a new HTML renderer, Event Logs, a theme picker.

## Capabilities

### New Capabilities

- `dorphnotes-contents-and-data`: One DorphNotes module for import/analyze plus HTML publish, including publish-from-analysis.

### Modified Capabilities

- `dorph-ai-chat-visual-theme`: header product title must stay high-contrast light on twilight glass (in-flight from `dorphai-futuristic-glass-ui` / `dorphai-twilight-glass-lift`).

## Impact

- `dorph/frontend` drawer, router, messages, embed labels, Generic data page (publish action), small `App.css` title rule.
- Existing `POST /api/composerx/publishes` (already proxied) for HTML create.
- Tests: `dorphNotesEmbeds.test.js`, drawer/router labels, one publish-from-analysis check.
- Docs: `docs/agents/03-dorph.md`, user-facing names.
