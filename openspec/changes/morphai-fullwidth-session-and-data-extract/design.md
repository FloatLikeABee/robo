## Context

See proposal.md for motivation.

Current constraints that shape the approach:

- Standalone Morph AI chat is capped by `.app { max-width: min(1600px, 100%) }` and `.app-outer` horizontal padding; embedded chat already uses `max-width: none`.
- Morph AI JWT default expiry is ~100 years, and `setMorphToken(..., rememberMe)` can still write a session-only cookie when Remember me is off.
- Morph Data is same-origin (`/morphdata`) so it reads `userspanel_session_token` automatically. Morph Utils is a separate origin (`http://localhost:3040` by default) and only receives the session via `?userspanel_token=` plus a cookie on *its* origin. `ensureSharedSession()` currently **clears** the token when `GET /api/auth/user` is not OK, including proxy/network failures.
- Generic Data PDF import uses a regex scrape of PDF string literals (`extractPDFTextLocal`), not a real text-layer extractor. Import accepts `.csv/.json/.pdf/.md` only — not Excel. `savePoppedDetail` returns `nil` when no entity-detail store is configured, so import can report success with empty `detail`.
- MySQL `generic_data.source_type` is `ENUM('csv','json','pdf')`. Assets Detail (JSON) is an `AdminDataGrid` `json_detail` field with no text-to-JSON helper.

## Goals / Non-Goals

**Goals:**

- Full-bleed standalone Morph AI shell; keep embedded chat compact.
- Login always persists on the device; Sign out is the only user-facing way to drop the session.
- Morph Utils keeps and reuses a Morph AI session after first handoff, without a second login and without clearing on transient auth-check failures.
- Generic Data import uses real PDF text-layer extraction, CSV + Excel tables, and fails loudly if content cannot be stored.
- Shared AI extract-to-JSON for Generic Data paste and Assets Detail (JSON).

**Non-Goals:**

- OCR / vision for scanned image-only PDFs.
- Merging Morph Utils into the Morph AI SPA (it stays a separate shell).
- Changing UsersPanel-dependent apps (formx, booki, etc.) beyond consuming the shared cookie/token they already accept.
- Auto-saving Assets or Generic Data on extract; user always reviews first.
- Redesigning Morph AI chrome beyond width / full-screen shell.

## Decisions

### 1. Full-width standalone shell via CSS only
- **Choice**: For the non-embedded shell, set `.app-outer` padding to 0 and `.app` to `width: 100%; max-width: none; height: 100dvh; border-radius: 0` so the panel is edge-to-edge. Leave `.app.app--embedded` rules as they are.
- **Rationale**: The cap is purely CSS; no layout rewrite. Embedded Morph Data chat must stay in the drawer.
- **Alternatives**: Keep 1600px cap with reduced padding — rejected; user asked for max-to-screen. New layout component — unnecessary.

### 2. Always-persist Morph AI session
- **Choice**: `setMorphToken(token)` always uses localStorage + cookie `Max-Age` (~100 years). Remove the Remember me checkbox. `getMorphToken` prefers cookie then localStorage; do not use sessionStorage for the auth token. Keep JWT default expiry (~100 years); Sign out calls `clearMorphSession()`.
- **Rationale**: Matches “permanent until logout”. JWT and cookie already support this; the opt-out is what makes sessions vanish on browser close.
- **Alternatives**: Refresh-token rotation — overkill. Server-side session store — not needed.

### 3. Morph Utils session: persist locally, clear only on real auth failure
- **Choice**:
  1. Keep Morph AI Apps `?userspanel_token=` handoff.
  2. On consume, Morph Utils stores the token in **localStorage and a long-lived cookie** (same pattern as Morph AI).
  3. `getSharedToken()` reads cookie then localStorage; skip clearing on parse/network errors.
  4. `ensureSharedSession()`: **401/403 from Morph AI → clear**; network/5xx/unreachable → **keep** the token and stay authed.
  5. Default Morph Utils Apps URL may stay `:3040` in local dev; once the token is stored on that origin, later visits work without the query param, like Morph Data after login.
- **Rationale**: Cross-origin cookies cannot be shared between Morph AI (`:3031`) and Utils (`:3040`). Morph Data works because it is same-origin. Local persistence after handoff plus “don’t clear on blips” is the Utils equivalent. Same-origin `/morphutils` proxy is a follow-up if we later unify hosts.
- **Alternatives**: Reverse-proxy Utils under Morph AI origin now — better cookie sharing, but a deploy/routing change beyond this fix. postMessage between windows — fragile.

### 4. PDF text via `pkg/docextract`
- **Choice**: Replace `extractPDFTextLocal` in Generic Data import with `docextract.ExtractPDFBytes` (already used elsewhere). Keep wrapping as markdown (`content_markdown`) for the existing PDF viewer. Map `docextract.ErrNoText` to a 400 with a clear message. Do not create a successful empty record.
- **Rationale**: The regex scraper misses most real PDF text layers and never walks pages. The shared extractor already handles temp-file + panic recovery.
- **Alternatives**: Call an external OCR API — out of scope. Keep regex as fallback — skip; it produces garbage.

### 5. Spreadsheets: parse Excel, store as tabular detail
- **Choice**: Accept `.xlsx` (and `.xls` only if the existing importer can parse it; otherwise 400 asking for `.xlsx`/`.csv`). Reuse `morph/importcol` or `hybridcontext` xlsx parsing. Store rows/columns like CSV. Keep `source_type` as `csv` for tabular files **or** extend MySQL ENUM with `xlsx` — prefer **extend ENUM / sqlite TEXT** so the UI can chip “Excel”. If ALTER is painful, store `csv` + `import_meta.format = "xlsx"` and still show a table. First worksheet only.
- **Rationale**: User asked to read CSV/Excel; Generic Data already has a CSV table viewer. First sheet is predictable.
- **Alternatives**: All sheets as nested JSON — more UI work than needed.

### 6. Import must fail if detail cannot be stored
- **Choice**: `ImportGenericData` requires a configured entity-detail store. If `savePoppedDetail` would no-op (`store == nil`) or returns an error, respond 503/500 and do not present success. Prefer deleting the SQL row on detail-write failure so the list does not show an empty import. Frontend already shows `error` from the API — keep that, and allow Excel in the file picker.
- **Rationale**: Today a missing store looks like “import worked but nothing is in the data.”
- **Alternatives**: Fall back to stuffing JSON into a SQL column — diverges from the entity-detail pattern used by Assets.

### 7. Shared extract-to-JSON API
- **Choice**: One authenticated endpoint, e.g. `POST /api/tran/extract-json` with `{ "text": "...", "purpose": "generic_data" | "asset_detail" }`. Prompt the model to return a single JSON object; parse with `morphai.ExtractJSONObject`. Limits: reuse a generous text cap (on the order of existing `text-assist` / analysis runes). Generic Data: import dialog gains a **Paste text** path (file remains available); extract → review JSON → save via existing create/import. Assets: Detail (JSON) toolbar gets **Extract from text**; if detail is non-empty, confirm before replace. Neither path saves the record until the user confirms/saves.
- **Rationale**: One prompt/parser for both surfaces. Existing `/api/tran/text-assist` is prose-only and the wrong contract.
- **Alternatives**: Two endpoints — duplicate. Client-side-only extract — still needs the model; keep it on the server with other Morph Data AI helpers.

## Risks / Trade-offs

- [Full-bleed chat may feel tight on ultrawide] → Mitigation: messages already cap their own bubble width; only the shell goes full width.
- [Long-lived JWT if a device is shared] → Mitigation: Sign out still clears cookie + localStorage; this is an explicit product request.
- [Utils on another origin still needs one Apps-menu (or URL) handoff] → Mitigation: persist token after that handoff; do not clear on network errors. Document that Utils login is Morph AI, not a Utils form.
- [Excel `.xls` / huge workbooks] → Mitigation: reject legacy `.xls` if unsupported; cap rows like CSV (`genericDataMaxCSVRows`); first sheet only.
- [PDF extractor misses scans] → Mitigation: explicit no-text error; OCR out of scope.
- [AI JSON may be messy] → Mitigation: review-before-save; user can edit JSON; invalid model output returns an extract error, not a half-saved record.

## Migration Plan

1. Ship CSS full-width and session persistence together (frontend-only; no data migration).
2. Ship Morph Utils auth.ts / App.tsx changes with Morph AI Apps handoff unchanged.
3. Deploy Generic Data import (PDF extractor, Excel, fail-on-missing-detail) with schema ENUM update if `xlsx` is added.
4. Add extract-json API, then Generic Data paste UI and Assets Detail extract UI.
5. Rollback: revert frontend CSS/auth and the new endpoint; leftover Generic Data rows stay valid.

## Open Questions

- None that block implementation. First-sheet-only Excel and no OCR are accepted limits.
