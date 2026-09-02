## Why

Morph AI still uses a capped chat panel, sessions can disappear without Sign out, and Morph Utils often needs a second login even after Morph AI is already signed in. Generic Data cannot reliably ingest full PDFs or spreadsheet files, import sometimes fails to persist, and neither Generic Data nor Assets let a user paste a description and have AI fill structured JSON.

## What Changes

- **Morph AI layout**: The chat shell fills the viewport width (full screen, no 1600px / side-padding cap). Embedded Morph Data chat is unchanged.
- **Persistent session**: After a successful Morph AI login, the session stays until the user explicitly Signs out. Remove the “Remember me” opt-out; closing the browser MUST NOT log the user out.
- **Morph Utils auth**: Opening Morph Utils after Morph AI login MUST work like Morph Data — no extra login, no expired-session bounce. The Morph AI Apps link continues to hand off the token; Utils MUST keep that session and MUST NOT clear it on a failed `/api/auth/user` check when Morph AI is still signed in.
- **Generic Data file import**: Read the full PDF text layer (not a partial string scrape). Accept CSV and Excel (`.xlsx` / `.xls` where supported). Persist imported content to the Generic Data record (SQL row + detail store). Surface a clear error instead of a silent empty save.
- **Generic Data from text**: Before or instead of a file, the user can paste plain text; AI extracts structured JSON and the user can save it as Generic Data.
- **Assets Detail (JSON) from text**: On Assets, the user can describe the asset in plain text; AI extracts JSON into the Detail (JSON) field for review and save.

## Capabilities

### New Capabilities
- `morphai-fullwidth-layout`: Morph AI chat uses the full viewport width.
- `morphai-persistent-session`: Signed-in Morph AI sessions last until explicit Sign out.
- `morphutils-shared-session`: Morph Utils reuses the Morph AI session the same way Morph Data does.
- `generic-data-ingest`: Generic Data import reads full PDFs, CSV, and Excel, persists content, and supports paste-text → AI JSON extract.
- `assets-detail-text-extract`: Assets Detail (JSON) can be filled from a plain-text description via AI extract.

### Modified Capabilities
- (none — `openspec/specs/` has no archived baselines)

## Impact

- **Morph AI frontend**: `morph/frontend/src/App.css`, `SkoolAiChat.js`, `LoginPage.js`, `auth/morphSession.js`.
- **Morph auth**: `morph/auth/jwt.go` (keep long-lived JWT; no session-cookie TTL), login always persists.
- **Morph Utils**: `morph-utils/frontend/src/auth.ts`, `App.tsx`, Vite proxy / Morph AI URL handoff.
- **Generic Data**: `morph/handlers/tran_generic_data.go`, `generic_pdf_text.go`, schema `source_type`, `GenericData.js`; reuse `pkg/docextract` and existing xlsx parsers (`morph/hybridcontext/xlsx_text.go`, `morph/importcol/parse.go`).
- **Assets**: `AdminDataGrid.js` / `Vehicles.js` Detail (JSON) editor; new extract API shared with Generic Data.
- **APIs**: Generic Data import types expand; new extract-to-JSON endpoint(s) for pasted text (Generic Data create + Assets detail).
