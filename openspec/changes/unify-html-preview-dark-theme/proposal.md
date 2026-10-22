## Why

Generated and previewed HTML across the stack still reads as **green/teal** in places—especially the **Project** module (`dorph-engi`), whose published documents use a teal accent and emerald radial wash (`#2dd4bf`, `#134e4a`). DorphNotes Timelines, Big notes, Research, and Case/task already use a dark-blue base (`#0b1220`) with sky-blue accents. The mismatch makes Project (and some rendered Markdown previews) look like a different product when operators flip to the HTML tab or open a published page.

## What Changes

- Define one **dark blue + dark purple** HTML document palette (background, card, ink, link accent, subtle purple gradient wash) aligned with Dorph AI / DorphNotes / DorphUtils chrome.
- **Regenerate templates** for Project documents in `dorph-engi` so new and updated `html_content` matches that palette (no teal/green gradient).
- **Retheme Project UI** (dorph-engi frontend): shell tokens, Markdown prose links, and DorphUtils sidebar accent for the Project module—dark blue/grey like DorphNotes, not teal or pink/rose chrome.
- **Audit DorphNotes HTML previews** (Timelines, Big notes, Research, Case/task HTML tab, `darkPreviewSrcDoc` iframes) and align accent/gradient tokens to the shared palette (purple tint in gradient, consistent link color).
- **Preview wrapper**: when DorphNotes or Project shows stored HTML in an iframe, inject or normalize so legacy green/teal documents still preview on-brand until the record is re-saved.
- Dark-only. No light-mode switch. No API shape changes.

## Capabilities

### New Capabilities

- `html-preview-dark-theme`: Shared dark blue / dark purple styling for generated HTML documents and in-app HTML preview surfaces across DorphNotes and Project.

### Modified Capabilities

- (none — no archived baselines under `openspec/specs/`)

## Impact

- `dorph/handlers/tran_timelines.go`, `tran_big_notes.go`, `tran_case_task_views.go`, `tran_research.go` (HTML builders)
- `dorph/frontend/src/lib/darkPreviewSrcDoc.js` and admin pages with HTML iframes (Timelines, BigNotes, Research, CaseTasks)
- `dorph/frontend/src/pages/admin/caseTaskViewDocs.js` (client-side HTML export)
- `dorph-engi/backend/src/api/project_docs.rs` (`build_project_html`)
- `dorph-engi/frontend/src/app.css`, `ProjectDocumentPanel.svelte`, `lib/browserStore.ts` (preview HTML stub)
- `dorph-utils/frontend/src/config.ts` (Project module accent)
- Browser/CSS only for Content Maker unless its previews are found using off-palette greens (spot-check during apply).
