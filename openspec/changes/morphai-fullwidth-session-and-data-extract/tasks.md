## 1. Morph AI full-width layout

- [x] 1.1 In `morph/frontend/src/App.css`, make standalone `.app-outer` / `.app` edge-to-edge (no 1600px cap, no side gutters, full viewport height)
- [x] 1.2 Confirm `.app.app--embedded` still fills only the Morph Data drawer and is unchanged

## 2. Persistent Morph AI session

- [x] 2.1 Update `morphSession.js` so login always writes localStorage + long-lived cookie; stop using sessionStorage for the auth token
- [x] 2.2 Remove the Remember me checkbox from `LoginPage.js` and always persist on successful login
- [x] 2.3 Confirm Sign out still calls `clearMorphSession()` and the next visit requires login

## 3. Morph Utils shared session

- [x] 3.1 In `morph-utils/frontend/src/auth.ts`, persist the handoff token in localStorage + long-lived cookie (cookie then localStorage on read)
- [x] 3.2 Change `ensureSharedSession()` to clear the token only on Morph AI 401/403; keep the session on network/5xx failures
- [x] 3.3 Keep Morph AI Apps `userspanel_token` handoff; smoke-check Utils is authed after Morph AI login without a second login form

## 4. Generic Data file import

- [x] 4.1 Replace Generic Data PDF scrape with `pkg/docextract.ExtractPDFBytes`; map no-text PDFs to a clear 400
- [x] 4.2 Accept `.xlsx` (and `.xls` only if parseable) in import; parse first sheet into columns/rows like CSV
- [x] 4.3 Fail import if the entity-detail store is missing or `savePoppedDetail` fails; do not leave a successful empty record
- [x] 4.4 Update `GenericData.js` file picker/copy to allow CSV, Excel, JSON, PDF, Markdown and show API errors

## 5. Extract text to JSON

- [x] 5.1 Add `POST /api/tran/extract-json` (`text` + `purpose`: `generic_data` | `asset_detail`) that returns a JSON object via Morph AI
- [x] 5.2 In Generic Data import, add paste-text → extract → review JSON → save; cancel must not create a record
- [x] 5.3 On Assets Detail (JSON), add extract-from-text; confirm before replacing existing detail; do not auto-save the asset

## 6. Verify

- [x] 6.1 Smoke-check full-width Morph AI, persistent session across browser restart, Morph Utils after Morph AI login, Generic Data PDF/CSV/Excel import + paste extract, and Assets Detail extract
