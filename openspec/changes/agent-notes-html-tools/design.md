# Design

## Context

See proposal.md. Agent Notes (`AgentNotesModal.js`) loads the list and leaves `selectedId` null until a click. DorphNotes HTML builders (`htmldoc.DarkDocumentCSS`, `contentsDataPublish.js`, case-task CSS, preview override) still stamp page `#2c2640`. DorphTools Documents (`PublishHub.js`) has a Graphic documents tab in front of `GraphicDocumentGenerator`. The header title uses `WebkitTextFillColor: transparent` on a purple gradient, the same failure the Dorph AI title had.

## Goals / Non-Goals

**Goals:**

- Select the first note only when the load finishes and nothing is selected yet.
- Point new DorphNotes HTML at the bento hexes already used by the operator UI.
- Remove the Graphic documents tab. Make the title solid ink.

**Non-Goals:**

- Deleting `/graphic-document/generate`.
- Rewriting stored HTML.
- A second HTML renderer.

## Decisions

### 1. Select inside the load, not on every empty selection

**Choice:** when the notes response arrives and the list is non-empty, set `selectedId` to the first id. The back button still sets `selectedId` to null and does not refetch, so the list stays. An effect that selects whenever `selectedId` is null would undo Back.

### 2. One palette in the shared CSS

**Choice:** change `darkRootVars` in `htmldoc` to `--bg:#221c34` and `--card:#322a4a`, and keep accents `#8b5cf6` `#a78bfa` `#f97316`. Point `contentsDataPublish.js`, `darkPreviewSrcDoc.js`, `caseTaskViewDocs.js`, and the form-HTML prompt at those same hexes. Update the tests that still require `#2c2640`.

### 3. Remove the tab, keep the API

**Choice:** Documents tabs are Articles and ScholarForge. Drop the Graphic documents import and the tab index. `/graphic-document` keeps redirecting to `/documents`. Leave `generateGraphicDocument` on the API.

Rejected: deleting the backend route in the same change. The UI is what the operator sees.

### 4. Solid title

**Choice:** header title `color: #f4f0fa`, no background-clip and no transparent fill.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Back immediately reselects | Set the id only in the load callback |
| Old tests expect `#2c2640` | Update those assertions with the builders |
| A bookmark to Graphic documents | The existing redirect lands on Documents |

## Migration Plan

Frontend and HTML template colors only. No data migration.

## Open Questions

None.
