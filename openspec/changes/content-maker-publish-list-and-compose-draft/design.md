## Context

See proposal.md for motivation. Content Maker (ComposerX, MorphUtils embed `VITE_COMPOSERX_URL` default `localhost:8044`) uses SQLite (`COMPOSERX_SQLITE_PATH`, `schema_sqlite.sql`): `published_pages.created_at` / `updated_at` are TEXT `CURRENT_TIMESTAMP`. `PublishedPageRepository.List` scans those columns into `time.Time`. SQLite drivers yield `string`; Scan fails; `listPublishedPages` returns 500 `"failed to list published pages"`. `publish_drafts` List/Get has the same Scan. MorphNotes already uses a `scanDestTime`-style dest for this.

Compose content is `{#if currentPage === compose}` → `<ComposePublishPanel>` in `App.svelte`. Published contents is a sibling `{:else if}`. Switching tabs **destroys** the panel; all `$state` (name, `pageHtml`, AI messages) is gone. On remount, `$effect` sets starter HTML when `pageHtml` is empty. Email composer autosave is a different page (`EMAIL_COMPOSER` / saved emails), not this HTML compose flow.

Product: Content Maker (not ComposerX) in user-facing copy. Dark UI only. Keep MorphUtils left nav; keep inner header tabs. Do not auto-publish on navigate.

## Goals / Non-Goals

**Goals:**

- History and drafts lists work on SQLite TEXT timestamps (TDD on List Scan).
- Compose HTML/name/AI chat survive Compose ↔ Published contents in one SPA session.
- Published contents still loads lists on enter.

**Non-Goals:**

- Dual-write MySQL; changing public slug URLs.
- Surviving a full browser refresh unless keep-alive makes it free (sessionStorage is optional backup, not required by spec).
- Auto-save to `/publish-drafts` on every keystroke (manual Save draft stays).

## Decisions

### 1. Flexible timestamp Scan (not schema rewrite)

**Choice:** Scan `created_at` / `updated_at` with a dest that accepts `[]byte`, `string`, and `time.Time` (same idea as MorphNotes `scanDestTime`). Apply to published list and publish-drafts list/get.

**Why:** `CREATE TABLE IF NOT EXISTS` will not convert existing TEXT columns. Parsing TEXT is the fix Morph already uses.

**Alternative:** Store INTEGER unix times — rejected (migration + existing rows still TEXT).

### 2. Keep Compose panel mounted (hidden), do not only lift state

**Choice:** Always mount `ComposePublishPanel` while the operator is in Content Maker; hide it with CSS when `currentPage` is Published contents (same pattern as MorphUtils iframes `is-hidden`). Published records panel can stay `{#if}` (stateless reload on enter is fine).

**Why:** Spec is “same session, click away and back.” Unmount is the wipe. Hidden keep-alive preserves AI chat DOM without a new store.

**Alternative A:** Lift all compose fields into `App.svelte` — more plumbing, same outcome.

**Alternative B:** `sessionStorage` only — survives remount/refresh but still flashes starter if `$effect` races; use as optional backup, not the only mechanism.

**Do not** reset to starter HTML when `pageHtml` is already non-empty (keep existing `$effect` guard).

### 3. Do not auto-POST drafts on tab switch

**Choice:** Navigation does not call `POST /publish-drafts` or `POST /publishes`.

**Why:** Spec: tab switch must not publish. Silent draft rows would clutter Published contents.

## Risks / Trade-offs

- [Hidden panel still runs `$effect` / path-resolve fetch] → Keep effects cheap; don’t re-fetch slug on hide unless name changed.
- [Memory of large HTML in hidden iframe] → Acceptable for one compose session.
- [List error was also auth 401] → UI already requires Morph session; this change is the 500 Scan path. Do not treat 401 as a list bug.

## Migration Plan

- No SQLite migration.
- Restart Content Maker API after Scan fix; reload UI after keep-alive.
- Rollback: revert Scan helper and `{#if}` mount.

## Open Questions

None. Refresh persistence is optional sessionStorage during apply if keep-alive is not enough for MorphUtils iframe remounts.
