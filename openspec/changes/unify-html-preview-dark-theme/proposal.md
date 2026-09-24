## Why

Generated and previewed HTML across the stack still reads as **green/teal** in places—especially the **Project** module (`morph-engi`), whose published documents use a teal accent and emerald radial wash (`#2dd4bf`, `#134e4a`). MorphNotes Timelines, Big notes, Research, and Case/task already use a dark-blue base (`#0b1220`) with sky-blue accents. The mismatch makes Project (and some rendered Markdown previews) look like a different product when operators flip to the HTML tab or open a published page.

## What Changes

- Define one **dark blue + dark purple** HTML document palette (background, card, ink, link accent, subtle purple gradient wash) aligned with Morph AI / MorphNotes / MorphUtils chrome.
- **Regenerate templates** for Project documents in `morph-engi` so new and updated `html_content` matches that palette (no teal/green gradient).
- **Retheme Project UI** (morph-engi frontend): shell tokens, Markdown prose links, and MorphUtils sidebar accent for the Project module—dark blue/grey like MorphNotes, not teal or pink/rose chrome.
- **Audit MorphNotes HTML previews** (Timelines, Big notes, Research, Case/task HTML tab, `darkPreviewSrcDoc` iframes) and align accent/gradient tokens to the shared palette (purple tint in gradient, consistent link color).
- **Preview wrapper**: when MorphNotes or Project shows stored HTML in an iframe, inject or normalize so legacy green/teal documents still preview on-brand until the record is re-saved.
- Dark-only. No light-mode switch. No API shape changes.

## Capabilities

### New Capabilities

- `html-preview-dark-theme`: Shared dark blue / dark purple styling for generated HTML documents and in-app HTML preview surfaces across MorphNotes and Project.

### Modified Capabilities

- (none — no archived baselines under `openspec/specs/`)

## Impact

- `morph/handlers/tran_timelines.go`, `tran_big_notes.go`, `tran_case_task_views.go`, `tran_research.go` (HTML builders)
- `morph/frontend/src/lib/darkPreviewSrcDoc.js` and admin pages with HTML iframes (Timelines, BigNotes, Research, CaseTasks)
- `morph/frontend/src/pages/admin/caseTaskViewDocs.js` (client-side HTML export)
- `morph-engi/backend/src/api/project_docs.rs` (`build_project_html`)
- `morph-engi/frontend/src/app.css`, `ProjectDocumentPanel.svelte`, `lib/browserStore.ts` (preview HTML stub)
- `morph-utils/frontend/src/config.ts` (Project module accent)
- Browser/CSS only for Content Maker unless its previews are found using off-palette greens (spot-check during apply).
