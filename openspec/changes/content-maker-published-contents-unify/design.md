## Context

See proposal.md — Why. Content Maker already stores saved HTML in `publish_drafts` and live pages in `published_pages`. `POST /publishes` always `Create`s and uniquifies colliding names (`name-2`). Published contents renders two cards in `ComposePublishRecordsPanel.svelte`; `table.grid td` uses `vertical-align: top`, so View/Delete/Open sit below the name/path baseline.

## Goals / Non-Goals

**Goals:**

- One Published contents table keyed by slugified name (published wins when both exist).
- Publish upserts by slug: same name keeps `/public/p/{slug}`; HTML/theme update in place.
- Action cells use middle alignment so buttons sit with the text.

**Non-Goals:**

- Unpublishing, custom slugs, or merging already-divergent `foo` vs `foo-2` history (leave extra rows; new publishes do not add more).
- Changing MorphUtils iframe auth, Compose keep-alive, or Event Logs / Data Access / MorphNotes.

## Decisions

1. **Client merge for the list; server upsert for uniqueness.** Keep `GET /publish-drafts` and `GET /publishes/history`. The panel merges by `slugify(name)` into one row (status Saved / Published; path only when published). Uniqueness is enforced on `POST /publishes` so a second Publish cannot race into a second path. Alternative: one new list API — extra surface for the same join.

2. **Upsert, not 409.** “Published only once” means one URL, not a hard lock that forces a new name. `Create` becomes find-by-slug then update HTML/theme/updated_at, else insert with the exact slug (no `nextUniqueSlug`). Alternative: 409 — worse for edit-and-publish.

3. **Alignment via row CSS.** `vertical-align: middle` on cells; action group `display: inline-flex; align-items: center`. Do not restyle unrelated grids.

4. **Drop the panel-meta lede** (“Saved HTML drafts and published page history.”). Heading stays Published contents.

## Risks / Trade-offs

- [Existing `name` and `name-2` both live] → New publishes do not create a third; list still shows both until an operator deletes. No silent merge of distinct slugs.
- [Two drafts with names that slugify the same] → One row; prefer newest `updated_at`. Rare if names are distinct in the UI.
- [Concurrent first publish] → Unique slug index (or equivalent) so only one insert wins; loser retries as update.

## Migration Plan

Deploy API upsert before relying on the UI. Rollback: restore uniquify-on-create and the two-table panel; existing upserted rows remain valid single paths.

## Open Questions

None. Republish updates the same path (not 409).
