## Why

Morph Utils and Morph Data still surface leftover chrome: a Data Access health banner that fires on a 500, intro blurbs under page titles, an Assets nav nobody needs, a bright-white light theme, and a User Settings shortcut that does not earn its place. Operators want quieter pages, MorphNotes naming, Events & Info they can search and delete, and a header that puts Morph AI next to Notes.

## What Changes

- **Stop Data Access health polling.** Do not ping `/api/v1/health` on an interval or show **DataX API offline / health check failed (500)** as a standing banner. Real request failures still show as the failing action’s error.
- **Remove intro ledes** under Morph Utils module titles (starting with Events & Info “Operational notes…”). Sweep Event Logs, Info Sheets, Data Access, Content Maker, and Project for the same “paragraph under the h1” pattern and delete those lines.
- **Events & Info list:** title search; delete one event from the list; batch-delete selected events (confirm first).
- **Morph Data → MorphNotes** in user-visible product name (header, drawer, landing, document title, Morph AI app link).
- **Remove the Assets module** from MorphNotes nav. Default landing becomes Generic data. `/assets` (and old people/resources aliases) redirect there.
- **Rename Configuration → Settings** in the MorphNotes drawer (path may stay `/configuration`).
- **Remove MorphNotes User Settings** page and header person button.
- **Header order:** Notes, Morph AI shortcut, then theme toggle (Morph AI immediately left of theme, beside Notes).
- **Soften light mode** across Morph AI, MorphNotes, and Morph Utils embeds: replace bright white page/paper/surfaces with a slightly darker, warmer gray so light theme is easier on the eyes.

## Capabilities

### New Capabilities

- `datax-no-health-banner`: Data Access does not run a periodic health check or show an offline banner from that check.
- `utils-no-page-intros`: Morph Utils module pages do not show introductory copy under the page title.
- `events-info-search-delete`: Events & Info list supports title search, single delete, and batch delete.
- `morphnotes-product-nav`: Product is MorphNotes; Assets and User Settings are gone; Settings replaces Configuration; header is Notes + Morph AI + theme.
- `softer-light-theme`: Light theme backgrounds are muted (not stark white) on Morph AI, MorphNotes, and Morph Utils.

### Modified Capabilities

- (none — no main specs under `openspec/specs/`)

## Impact

- **Data Access** (`SharpReport/frontend`): `backendHealth.ts`, `BackendStatusBanner.svelte`, `+layout.svelte`; keep `/api/v1/health` on the API if it exists, unused by the UI.
- **Event Logs** (`formx/`): `EventsInfo.tsx`, `SurveyBot.tsx`; list `q=` + batch-delete API in `events_info.go` / `event_repo.go`.
- **Data Access / Content Maker / Project frontends:** page-title intro paragraphs.
- **MorphNotes** (`morph/frontend`): `AppDrawer.js`, `AdminLayout.js`, `appRouter.js`, `DocumentBranding.js`, `platformUiDefaults.js`, `theme.js`, `App.css`; drop User Settings route; Morph AI header link order.
- **Landing / Morph AI header** labels that still say Morph Data.
