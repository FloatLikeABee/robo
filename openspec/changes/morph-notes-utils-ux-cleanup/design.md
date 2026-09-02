## Context

See proposal.md for motivation.

Today Data Access (`SharpReport`) starts `startBackendHealthMonitor` in `+layout.svelte`, which GETs `/api/v1/health` every 45s and on visibility; a non-OK status sets `backendStatus` to offline and `BackendStatusBanner` shows “DataX API offline” plus “health check failed (500)”. Real API errors already throw in `api.ts`.

Events & Info (`formx`) already has `DELETE /api/v1/events-info/:id` and a drawer-only `remove`. The list has no search, checkboxes, or row delete. `ListEventInfo` is paginated with no `q`. Morph Utils page intros live as `<p>` under `<h1>` in `EventsInfo.tsx`, `SurveyBot.tsx`, SharpReport `data-tables` and `docs` layouts.

MorphNotes is still labeled Morph Data / MorphData (`platformUiDefaults`, `DocumentBranding`, `SkoolAiChat` header, landing). Nav includes Assets (`AppDrawer` + `appRouter` default `assets`). Header is Notes → theme → Morph AI → User Settings.

Light tokens use `#ffffff` / `#f9f8fb` (MorphNotes `theme.js`), `#ffffff` chat surfaces (`App.css`, `platform-chat/chat-tokens.css`), and similar in formx/SharpReport/composerx.

## Goals / Non-Goals

**Goals:**

- Remove health-driven banner and polling; leave per-request errors.
- Drop Utils title ledes; add Events search + list/batch delete.
- MorphNotes naming and nav/header as specified; muted light backgrounds.

**Non-Goals:**

- Deleting the Data Access `/health` route if it exists.
- Deleting MorphNotes asset *data* or import APIs; only the Assets nav/module.
- Renaming URL `/morphdata` or `/configuration` (labels only unless a redirect is required).
- Dark-theme redesign.
- Public Events submit page copy (not a Morph Utils admin title lede).

## Decisions

### D1: Stop the monitor; do not keep a silent health ping

**Choice:** Remove `startBackendHealthMonitor` from layout and stop rendering `BackendStatusBanner` from health status. Do not call `pingBackend` on an interval. Optional: leave `backendHealth.ts` unused or delete the banner component if nothing else imports it.

**Why:** The 500 is the health route, which the user does not want in the normal path.

**Alternatives:** Fix the 500 and keep the banner. Rejected — user said we do not need the health check.

### D2: Title search via `q` on list, plus client filter if the current page is the full set

**Choice:** Add optional `q` to `GET /events-info` and filter titles in `EventInfoRepo.List` (case-insensitive contains). The UI search box debounces and reloads. If limit already returns the working set, filtering in the repo still keeps pagination honest.

**Why:** List is paginated (default 50); client-only filter would miss later pages.

**Alternatives:** Client-only filter. Rejected for pagination.

### D3: Batch delete as `POST /events-info/batch-delete` with `{ ids: string[] }`

**Choice:** New authenticated POST (or DELETE with JSON body). Loop existing `Delete` per id; skip missing. UI: checkboxes, “Delete selected”, confirm. Row-level delete uses existing DELETE.

**Why:** Avoids a new storage model; reuses `EventInfoRepo.Delete`.

**Alternatives:** Only UI looping DELETE. Acceptable fallback if POST is extra; prefer one round-trip for many ids.

### D4: MorphNotes is a display name; keep `/morphdata` paths

**Choice:** Change `product_name` default and hardcoded “Morph Data” strings in chrome. Keep `ADMIN_BASE_PATH` `/morphdata` and internal ids.

**Why:** Avoids breaking bookmarks and embeds.

### D5: Assets routes redirect to generic-data; keep Resources.js unlinked

**Choice:** Default index → `generic-data`. Redirect `assets`, `resources`, `people`, etc. to generic-data. Remove Assets `ListItemButton`. Do not delete backend vehicle/asset handlers in this change.

**Why:** Generic data is the remaining dump; asset records can still exist for import/API.

### D6: Light canvas ≈ `#e8e4dc` / `#ece8e1` family, paper ≈ `#f0ece4`

**Choice:** Shift MorphNotes `background.default`/`paper`, Morph AI `--chat-page-bg` / `--chat-surface` / `--chat-msg-bg` / `--chat-input-bg`, platform-chat light tokens, Event Logs light `bg-[#f5f8ff]`, Data Access light CSS variables, Content Maker/Project page backgrounds off pure white toward a dimmer warm gray. Contrast: keep text near `#1a1423` / slate-800.

**Why:** User asked for darker light, not a new accent.

**Alternatives:** Dark-gray light mode (too close to dark theme). Rejected.

### D7: Header order Notes | Morph AI | theme

**Choice:** Reorder `AdminLayout.js` icon buttons; delete User Settings `IconButton` and route (`Navigate` to generic-data).

## Risks / Trade-offs

- **[Health banner hid a truly down API]** → Per-request errors still appear; operators restart from start-all as today.
- **[Bookmarks to /assets]** → Redirect to generic-data; document in tasks.
- **[product_name override in DB]** → If Display names stored MorphData, update default and the Display names page default; stored custom names still win until edited.
- **[Light theme too dark / low contrast]** → Keep text dark; only lift canvas off #fff by ~8–12%.

## Migration Plan

- Frontend-heavy; FormsX adds list `q` + batch-delete. Deploy morph-ui, formx-ui/api, SharpReport UI, landing together for labels.
- Rollback: revert those frontends; Assets URLs still redirect until revert.

## Open Questions

None. MorphNotes spelling, header order, and dropping health checks are fixed by the request.
