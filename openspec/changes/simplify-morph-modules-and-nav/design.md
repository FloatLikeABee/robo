## Context

See proposal.md for motivation. Today:

- Morph login sets a shared cookie with `SESSION_MAX_AGE_SECONDS` (48h) in `morph/frontend/src/auth/morphSession.js`.
- Morph Data drawer has Big notes, then a separate **Work data** group (Generic data, Places, People, Assets, Activities), plus Configuration including Display labels and Data import (`AppDrawer.js`).
- formx is embedded as SheetX with Forms + AI Sheets (`formx/frontend/src/components/Layout.tsx`); Morph Utils labels it SheetX (`morph-utils/frontend/src/config.ts`).
- Booki nav includes Data AI, Bookings, Accounting, Flow Log, Warehouse, Assets, Settings.
- Academi has Chat / Community / Profile surfaces in `academi/web`.
- Morph Utils has no shell-level Settings; modules carry their own.
- Morph AI mounts `MorphTaskChainsModal` and `pollDueTaskChains` from `SkoolAiChat.js`.

## Goals / Non-Goals

**Goals:**

- Align nav, labels, and import UX with the locked product IA across Morph AI, Morph Data, and Morph Utils embeds.
- Prefer UI/route removal and renames over large backend rewrites where APIs can remain compatible.
- One Morph Utils Settings shell entry that replaces per-module settings/profile chrome.

**Non-Goals:**

- Deleting Booki/SurveyX backend tables or Academi community data stores in this change (hide/remove UI and routes first).
- Renaming internal package folders (`formx`, `booki`) or database schemas unless required for branding strings.
- Changing Morph AI chat core, Big notes generation, or non-listed Morph Utils modules (ComposerX, DataX, Projects) beyond shared Settings policy.
- Migrating historical task-chain localStorage data.

## Decisions

### 1. Session persistence: cookie without Max-Age logout policy
- **Choice**: Treat “no session timeout” as: stop using a finite Max-Age that forces logout; persist token (localStorage + session cookie without expiry, or equivalent long-lived persistence) until explicit sign-out. Align UsersPanel/shared cookie behavior if Morph login shares it.
- **Alternatives**: Keep 48h Max-Age but auto-refresh — rejected; user asked for no timeout.
- **Note**: Server JWT expiry (if any) must be extended or made refreshable so the client change is effective.

### 2. Work data under Big notes
- **Choice**: Nest Generic data / People / Assets / Activities as children of Big notes (expandable or always-visible nested list). Remove the standalone Work data section header and Places item. Keep existing routes (`/generic-data`, `/people`, `/assets`, `/activities`) to minimize churn; members redirects can remain.
- **Alternatives**: Flat list after Big notes without nesting — rejected; user asked for under Big notes.

### 3. Hardcoded labels; delete display-labels feature
- **Choice**: Hardcode labels in `platformUiDefaults` / drawer; remove Display labels page, route, and any persist/load of custom nav labels used only for renaming.
- **Alternatives**: Keep API but hide UI — rejected; user said no more display name changes.

### 4. File import
- **Choice**: Rename Data import → File import in nav/UI. Entity picker: Generic data, People, Assets, Activities. Accept CSV, Excel, JSON (reuse/extend existing import handlers; Excel via existing or lightweight parser).
- **Alternatives**: Separate importers per entity page only — rejected; keep Configuration → File import as the hub.

### 5. SurveyX branding in Morph Utils + formx UI
- **Choice**: Morph Utils module id may stay `sheetx` internally for env URLs (`VITE_SHEETX_URL`) while **label** becomes SurveyX; update formx Layout to drop Forms nav and rename AI Sheets → AI Surveys. Default embed path stays events-info / AI Surveys entry. Redirect `/forms` away from product nav (optional soft redirect to AI Surveys).
- **Alternatives**: Rename repo `formx` → `surveyx` — deferred (non-goal).

### 6. Booki slim-down
- **Choice**: Keep routes for Booki AI (`/`), Bookings, Flow log; remove nav + routes for Accounting, Warehouse, Assets, Settings (or redirect to `/`). User’s “bookings, bookings” treated as a single Bookings item.
- **Alternatives**: Leave routes reachable by URL — prefer remove from router to avoid dead features.

### 7. Academi renames
- **Choice**: String/label changes in Academi web nav: Chat → Academi AI, Community → Board; remove Profile tab/section.
- **Alternatives**: New routes — unnecessary if only labels change.

### 8. Shared Morph Utils Settings
- **Choice**: Add shell-level Settings route/page in `morph-utils` (same level as module tabs / More menu). Strip Settings/Profile from Booki, Academi, SurveyX, and other Utils embeds’ primary nav. Settings content: account/session essentials and any cross-app prefs currently duplicated; deep module-specific prefs can be deferred into this page later.
- **Alternatives**: iframe a UsersPanel settings page — acceptable if already shared auth; prefer thin Morph Utils page first.

### 9. Remove Morph AI task chains
- **Choice**: Delete UI entry, modal usage, and `pollDueTaskChains` from Morph AI chat; leave unused files deletable in the same change to avoid dead code.
- **Alternatives**: Feature-flag off — rejected; product wants removal.

## Risks / Trade-offs

- [Shared cookie / UsersPanel TTL still expires] → Coordinate server and UsersPanel session policy with Morph client change.
- [Deep links to removed Booki/Forms/Profile routes] → Redirect to remaining home surfaces; document **BREAKING** in release notes.
- [Excel import complexity] → Ship CSV/JSON first-class; Excel via proven library already in repo if present, else add minimal dependency.
- [Nesting Work data under Big notes may confuse “Big notes” as notes-only] → Clear nested labels; Big notes parent still opens notes list when clicked.
- [Shared Settings initially thin] → Better empty shared Settings than leftover per-module settings.

## Migration Plan

1. Ship frontend nav/label/removals behind normal deploys (no DB migration required for IA).
2. Adjust Morph/UsersPanel session TTL or refresh so “no timeout” holds.
3. Soft-redirect removed paths (Forms, Booki accounting/warehouse/assets/settings, Academi profile, Morph Data display-labels, data-import → file-import).
4. Rollback: revert frontend deploys; session TTL can be restored independently.

## Open Questions

- Whether Morph Utils shared Settings should embed UsersPanel account UI or a Morph-only prefs page (does not change specs; decide at implement time).
