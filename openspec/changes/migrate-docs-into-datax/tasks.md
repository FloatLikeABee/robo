## 1. DataX Docs workspace shell

- [x] 1.1 Add Docs nav item and routes under SharpReport Data access (`/docs` and child routes)
- [x] 1.2 Scaffold Docs library page (list/create/open) wired to Academi docs APIs via DataX proxy or direct base URL
- [x] 1.3 Add DataX backend proxy/config for Academi Docs endpoints (`ACADEMI_API_BASE_URL` / `/api/v1/docs/*`)

## 2. Docs AI two-pane sessions

- [x] 2.1 Build Docs AI page with left session history and right chat pane
- [x] 2.2 Wire session create/list/select (and delete if already supported) to Docs chat-session APIs
- [x] 2.3 Wire message send + streaming/response into the right pane for the active session
- [x] 2.4 Attach content uploads (PDF/TXT+) on Docs AI; reject or clearly block tabular-only import UX on this path

## 3. Content vs data upload separation

- [x] 3.1 Label Docs uploads as content documents (PDF/TXT) in library and AI UI
- [x] 3.2 Keep DataX data-table import accept-list as CSV/JSON (unchanged); ensure Docs does not use that import endpoint for PDFs
- [x] 3.3 Return clear errors when wrong file type is used on each pipeline

## 4. Docs publish (Board replacement)

- [x] 4.1 Add publish UI for a Docs document (ComposerX-like): preview, optional analysis prompt, publish action
- [x] 4.2 Implement publish API that stores/serves HTML displaying the document
- [x] 4.3 When analysis prompt is present, run AI analysis and include it in the published HTML; when absent, publish display only
- [x] 4.4 Remove or redirect Docs Board/community feed primary entry to the publish flow

## 5. Morph Utils cleanup

- [x] 5.1 Remove `docs` from Morph Utils module list; update login/settings copy
- [x] 5.2 Redirect legacy Utils `/docs` and `/academi` into DataX Docs
- [x] 5.3 Point any Docs embed env/docs to DataX Docs path

## 6. Verification

- [x] 6.1 Smoke DataX: Docs library, AI two-pane sessions, PDF/TXT attach; data import still CSV/JSON
- [x] 6.2 Smoke publish HTML with and without analysis prompt; Board not primary
- [x] 6.3 Smoke Morph Utils: no Docs module; legacy Docs path lands in DataX Docs
