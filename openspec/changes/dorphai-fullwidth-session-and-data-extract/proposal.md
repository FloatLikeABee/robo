## Why

Dorph AI still uses a capped chat panel, sessions can disappear without Sign out, and Dorph Utils often needs a second login even after Dorph AI is already signed in. Generic Data cannot reliably ingest full PDFs or spreadsheet files, import sometimes fails to persist, and neither Generic Data nor Assets let a user paste a description and have AI fill structured JSON.

## What Changes

- **Dorph AI layout**: The chat shell fills the viewport width (full screen, no 1600px / side-padding cap). Embedded Dorph Data chat is unchanged.
- **Persistent session**: After a successful Dorph AI login, the session stays until the user explicitly Signs out. Remove the “Remember me” opt-out; closing the browser MUST NOT log the user out.
- **Dorph Utils auth**: Opening Dorph Utils after Dorph AI login MUST work like Dorph Data — no extra login, no expired-session bounce. The Dorph AI Apps link continues to hand off the token; Utils MUST keep that session and MUST NOT clear it on a failed `/api/auth/user` check when Dorph AI is still signed in.
- **Generic Data file import**: Read the full PDF text layer (not a partial string scrape). Accept CSV and Excel (`.xlsx` / `.xls` where supported). Persist imported content to the Generic Data record (SQL row + detail store). Surface a clear error instead of a silent empty save.
- **Generic Data from text**: Before or instead of a file, the user can paste plain text; AI extracts structured JSON and the user can save it as Generic Data.
- **Assets Detail (JSON) from text**: On Assets, the user can describe the asset in plain text; AI extracts JSON into the Detail (JSON) field for review and save.

## Capabilities

### New Capabilities
- `dorphai-fullwidth-layout`: Dorph AI chat uses the full viewport width.
- `dorphai-persistent-session`: Signed-in Dorph AI sessions last until explicit Sign out.
- `dorphutils-shared-session`: Dorph Utils reuses the Dorph AI session the same way Dorph Data does.
- `generic-data-ingest`: Generic Data import reads full PDFs, CSV, and Excel, persists content, and supports paste-text → AI JSON extract.
- `assets-detail-text-extract`: Assets Detail (JSON) can be filled from a plain-text description via AI extract.

### Modified Capabilities
- (none — `openspec/specs/` has no archived baselines)

## Impact

- **Dorph AI frontend**: `dorph/frontend/src/App.css`, `SkoolAiChat.js`, `LoginPage.js`, `auth/dorphSession.js`.
- **Dorph auth**: `dorph/auth/jwt.go` (keep long-lived JWT; no session-cookie TTL), login always persists.
- **Dorph Utils**: `dorph-utils/frontend/src/auth.ts`, `App.tsx`, Vite proxy / Dorph AI URL handoff.
- **Generic Data**: `dorph/handlers/tran_generic_data.go`, `generic_pdf_text.go`, schema `source_type`, `GenericData.js`; reuse `pkg/docextract` and existing xlsx parsers (`dorph/hybridcontext/xlsx_text.go`, `dorph/importcol/parse.go`).
- **Assets**: `AdminDataGrid.js` / `Vehicles.js` Detail (JSON) editor; new extract API shared with Generic Data.
- **APIs**: Generic Data import types expand; new extract-to-JSON endpoint(s) for pasted text (Generic Data create + Assets detail).
