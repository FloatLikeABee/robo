## Context

See proposal.md for motivation. MorphNotes Settings File import is `DataImport.js` at `configuration/file-import` (alias `configuration/data-import`), calling `POST /api/tran/generic-data/import`. Generic data still uses `POST /api/tran/generic-data/extract` for review-then-save. Tasks (`CaseTasks.js`) already have optional `start_at` / `end_at` datetime-local fields; the API stores them as nullable times. Timelines and Big Notes wrap an iframe in a Box with `overflow: auto`, so the pane and the iframe document both scroll. MorphUtils shell Sign out lives in `morph-utils/frontend/src/App.tsx`; modules still expose their own Out / Sign out / Logout. Project preview (`ProjectDocumentPanel.svelte`) uses `overflow-auto` with default scrollbar chrome.

Constraints that stay: MorphUtils left nav and module ids (`sheetx`, `/survey-bot`, `composerx`, `datax`, `projects`); dark-only product UIs; Morph JWT shared session; one repo-root `.env`; do not auto-save Generic data extract.

## Goals / Non-Goals

**Goals:**

- Delete the Settings File import surface and the import-only HTTP route without breaking Generic data extract/CRUD.
- Require Task start and end on create/update; default new drafts to today in the operator’s local calendar.
- Collapse Timelines HTML and Big Notes Preview/HTML to a single inner dark scrollbar.
- Remove MorphUtils and embedded-module logout chrome; keep Morph AI Sign out.
- Theme Project Markdown/HTML preview scrollbars dark.

**Non-Goals:**

- Removing Generic data as a MorphNotes module.
- Removing Morph AI Sign out or MorphNotes Users settings.
- Changing MorphUtils module ids, routes, or left-nav structure.
- Adding a light/dark theme switch.
- Migrating existing Task rows that already have null dates (edit/save then requires dates).

## Decisions

### 1. Remove File import page and import POST; keep extract

**Choice:** Delete `DataImport.js`, drawer item, routes, `tranEndpoints.genericDataImport`, `POST /api/tran/generic-data/import` (`ImportGenericData`), and chat-docs mention of that path. Keep extract + Generic data CRUD.

**Why:** User asked to remove Settings File import UI and backend. Extract is the Generic data review flow, not that Settings page. `genericDataImport` is only referenced from `DataImport.js`.

**Alternative:** Also delete Generic data. Rejected — Generic data is a separate MorphNotes list.

**Alternative:** Leave import API for scripts. Rejected — user asked backend gone.

### 2. Task dates stay datetime-local; required; default today 00:00 / 23:59 local

**Choice:** Keep `datetime-local`. New draft: start = today 00:00 local, end = today 23:59 local (same calendar day). Mark both `required`. Client blocks submit if either empty. Server create/update reject missing/unparseable `start_at` or `end_at` with 400. AI generate that omits dates: fill today in the form before the operator saves (do not persist AI output without dates).

**Why:** Current controls are datetime-local; “dates required, default current day” is satisfied without a schema migration. End-of-day end avoids a zero-length window when both default to midnight.

**Alternative:** Switch to `type="date"` only. Acceptable if apply prefers simpler UX; same spec (calendar day required, default today).

### 3. Preview pane: overflow hidden, iframe fills remaining height, dark scrollbars

**Choice:** Outer preview `Box` uses `overflow: hidden` (not `auto`). Iframe (or HTML host) is `flex: 1; height: 100%; minHeight: 0` so the document inside the iframe scrolls. Apply `color-scheme: dark` on the iframe and `scrollbar-color` / webkit scrollbar styles on the pane. For `srcDoc` HTML that does not already declare color-scheme, wrap or inject a small dark scrollbar style in the document head when rendering preview.

**Why:** Dual bars come from outer `overflow: auto` plus iframe internal scroll. Dark chrome needs `color-scheme` inside the iframe; CSS on the parent does not style the iframe’s document scrollbar.

**Applies to:** Timelines HTML tab; Big Notes Preview and HTML tabs. Markdown tabs may keep a single overflow; do not leave an extra outer bar around HTML.

### 4. Strip logout from MorphUtils shell and all four modules

**Choice:** Remove Sign out from `morph-utils` App header (stop calling `clearSharedToken` from UI). Remove Event Logs Layout logout, Content Maker header Out, Data Access UserMenu Sign out, Project AppLayout Sign out. Leave Morph AI `SkoolAiChat` Sign out.

**Why:** User asked all MorphUtils logout/sign out/out gone; they do not need a second logout. Modules are still reachable standalone, but operators use them through MorphUtils; removing the chrome in those apps is the only way the iframe does not show Out.

**Alternative:** Hide via `?embed=` query only. Rejected — user said all removed, not embed-only hide.

### 5. Project scrollbar theming on the preview container and iframe

**Choice:** On `ProjectDocumentPanel` preview `overflow-auto` region, set `scrollbar-width: thin`, `scrollbar-color` to track/thumb that match `--bg` / muted violet-gray, plus webkit thumb/track. HTML iframe: `color-scheme: dark` and inject the same scrollbar CSS into `srcdoc` when missing.

**Why:** Default macOS overlay/light bars read as “normal colored” on dark panels.

## Risks / Trade-offs

- [Standalone Event Logs / Content Maker / Data Access / Project have no in-app logout] → Mitigation: session still Morph JWT; Morph AI Sign out clears the shared token. Acceptable per product ask.
- [Existing Tasks with null dates fail on next save] → Mitigation: form requires dates; operator fills today or a real window. No bulk backfill.
- [Injecting CSS into HTML `srcDoc` could fight a light-themed published page] → Mitigation: only preview sandbox; published public pages unchanged. Big Notes can still store light HTML; preview scrollbar chrome is dark even if page background is light.
- [Removing import POST breaks old bookmarks/scripts] → Mitigation: 404 on that path; Generic data extract remains.

## Migration Plan

- Deploy Morph API + MorphNotes frontend together so File import nav and import POST disappear in one cut.
- No database migration.
- Rollback: restore `DataImport` route and import handler from git.

## Open Questions

None. Apply may pick `type="date"` vs datetime-local as long as both dates are required and default to today.
