## 1. Data Access (DataX) surface cut

- [x] 1.1 Remove Docs and Databases from SharpReport sidebar nav; keep Data tables, Data reports, Help
- [x] 1.2 Redirect legacy `/docs*` and `/databases*` routes to `/data-tables` (or Data reports)
- [x] 1.3 Strip Visual and SQL modes from Data reports builder; default and keep file-upload report flow only
- [x] 1.4 Update Data Access layout titles / help copy that still advertise Docs or Databases

## 2. Morph Utils shell

- [x] 2.1 Update `config.ts` Data Access and Project descriptions for the simplified scopes
- [x] 2.2 Remove Docs deep-link (`DATAX_DOCS_URL` / `isLegacyDocsPath`); legacy `docs`/`academi` open Data Access root only

## 3. Projects nav and Files-only surface

- [x] 3.1 Change Morph Engi `NAV` to Projects + Files only; remove People and Flow log tabs
- [x] 3.2 Remove People / Flow log page branches from `App.svelte` primary UI (leave APIs unused for now)
- [x] 3.3 Ensure Files list/upload empty states work as a simple retained-files library

## 4. Projects AI documents + publish

- [x] 4.1 Add project document persistence (markdown, html, publish fields) via migration or column adds on projects
- [x] 4.2 Implement create/generate API: multipart file(s) and/or paste; reject when both empty; AI → markdown; server HTML
- [x] 4.3 On paste source, save paste as a content file into the Files library and list it
- [x] 4.4 Register uploaded source files in the Files library when used for generation
- [x] 4.5 Implement publish + public HTML serve for project documents
- [x] 4.6 Replace Projects primary create UX with file/paste form, MD/HTML preview, publish (retire nested site-log/people/flow confirm as primary path)

## 5. Verification

- [x] 5.1 Data Access: no Docs/Databases in nav; Docs/DB legacy paths redirect; Data reports has no Visual/SQL
- [x] 5.2 Projects: only Projects + Files nav; create requires ≥1 source; paste-only and file-only work
- [x] 5.3 Paste creates a Files entry; publish serves public HTML; Morph Utils copy/deep-links no longer push Docs
