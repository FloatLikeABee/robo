## Context

See proposal.md — Why. The Skills dialog is a dark split: upload form left, catalog right. The left form uses `overflow: auto`, Instructions `rows={12}` and `min-height: 16rem`, plus a header lede. That stack is taller than the pane, so the left column scrolls. Catalog scroll on the right stays.

## Goals / Non-Goals

**Goals:**

- Compact Instructions (~4 rows / ~5–6rem) so the left form fits the pane.
- Left form `overflow: hidden` (no pane scrollbar). Long text: `overflow-y: auto` on the textarea only.
- Drop the Skills under-title hint to reclaim height.

**Non-Goals:**

- Changing skills APIs, Improve with AI, markdown parse, catalog behavior, or the chat skill picker.
- Making the dialog taller or stacking the form above the catalog.

## Decisions

1. **Shrink the field, don’t grow the dialog.** The overflow is the 16rem textarea, not a missing max-height on the modal. Alternative: keep a tall editor and hide the pane scrollbar — that clips Upload/Improve. Rejected.

2. **Textarea-internal scroll only when the body is long.** Empty/short drafts show no left-pane bar. Alternative: forbid textarea scroll too — then long markdown is uneditable. Rejected.

3. **Remove the header lede.** Session chrome: no under-title copy. Saves two lines and matches Morph AI headings.

## Risks / Trade-offs

- [Long skills harder to read in a short box] → Textarea still scrolls; operators can enlarge with `resize: vertical` only if it does not force the left pane to scroll (prefer `resize: none` in the dialog so the pane stays fit).
- [Narrow/short viewports] → Spec is typical desktop dialog (`min(80dvh, 820px)`). Very short windows may still clip; do not add a left-pane scrollbar as the default fix.

## Migration Plan

CSS + textarea rows in Morph AI frontend. Rollback: restore `min-height: 16rem`, `rows={12}`, form `overflow: auto`, and the hint.

## Open Questions

None.
