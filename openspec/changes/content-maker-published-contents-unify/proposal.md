## Why

Content Maker **Published contents** splits the same HTML work across two tables (Saved HTML files and Published history). Operators see duplicates, can publish the same file again as a new public path, and row action buttons sit off the text baseline.

## What Changes

- One list on Published contents: each HTML file is a single row (saved, published, or both).
- An HTML file can be published only once (one public path). Publishing the same file again MUST NOT mint a second slug or history row.
- Row operation buttons (View, Delete, Open) MUST sit vertically aligned with the name/path text in that row.
- Labels stay Published contents / Content Maker. Dark-only. No under-title lede. MorphUtils iframe and Morph AI SSO unchanged.

## Capabilities

### New Capabilities

- `published-contents-unified-list`: One table for saved HTML and published pages; actions aligned with row text.
- `html-file-publish-once`: At most one published URL per HTML file.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baselines)

## Impact

- Content Maker UI: `ComposePublishRecordsPanel.svelte` (and related compose publish actions if uniqueness is enforced at Publish).
- Content Maker API: `POST /publishes` uniqueness / upsert; list join of drafts + published pages if the UI cannot merge safely client-side.
- MorphUtils Content Maker iframe only. No Event Logs, Data Access, MorphNotes, or Morph AI chrome.
