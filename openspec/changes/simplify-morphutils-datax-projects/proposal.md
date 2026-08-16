## Why

Morph Utils’ **Data Access** (DataX / SharpReport) still exposes Docs (broken / HTTP 500), Databases, and a Data reports builder with Visual and SQL modes that users do not want. **Projects** (Morph Engi) is still a multi-module job tracker (people, flow log, structured import) instead of a simple AI path: requirements/specs in → organized project markdown + HTML out, with a plain Files list. Simplifying both apps makes Morph Utils match the product job users actually need.

## What Changes

### Data Access (DataX / SharpReport, embedded as Morph Utils “Data Access”)

- **BREAKING:** Remove the **Docs** module from nav and primary UX (routes redirect or drop; stop deep-linking Morph Utils `/docs`/`/academi` into DataX Docs).
- **BREAKING:** Remove **Databases** (connections / linked DBs) from nav and primary UX — no user-facing database linking.
- **Data reports:** Keep **upload file → data report** only. **BREAKING:** Remove **Visual** and **SQL** builder modes from the Data reports UI.
- Keep **Data tables** (and Help if useful); update Data Access copy so it no longer advertises Docs/databases/queries.

### Projects (Morph Engi, embedded as Morph Utils “Project”)

- **BREAKING (nav):** Projects app becomes **two modules only**: **Projects** and **Files**. Remove **People** and **Flow log** (and related create/list UX) from primary nav.
- **Projects:** Redesign create/extract around AI reading **uploaded files** (requirements, specifications, descriptions) and/or **pasted content**, organizing into project document content.
  - Output **markdown** and **HTML**.
  - User can **publish** the result (public HTML, same idea as Composer / Morph Big Notes–style publish).
  - Source inputs are multi-select (file and/or paste) but **at least one is required**; otherwise show a clear warning/error.
- **Files:** A simple library that **stores and lists uploaded files** (and paste-origin content files). No project-management complexity beyond list/view/delete as needed.
- When the user **pastes** content during extract, the system MUST also **save that paste as a content file** so it appears in the Files list.
- Retire or hide the prior structured “import → site logs / people / flow entries” confirm flow as the primary Projects create path (replace with MD/HTML project documents).

### Morph Utils shell

- Update Data Access / Project module descriptions; remove Docs deep-link override (`DATAX_DOCS_URL` / `isLegacyDocsPath`).

## Capabilities

### New Capabilities
- `datax-simplified-surface`: Data Access without Docs and Databases; Data reports limited to file-upload mode (no Visual/SQL).
- `projects-ai-documents`: Projects module generates organized MD+HTML from file and/or paste (at least one required), with publish; paste also stored as a Files entry.
- `projects-files-library`: Files module lists and retains uploaded (and paste-saved) content files only; People and Flow log removed from primary surface.

### Modified Capabilities
- (none — `openspec/specs/` has no published baselines for these products)

## Impact

- **SharpReport (DataX):** `Sidebar.svelte`, docs routes, databases routes, `reports/builder/+page.svelte` (drop Visual/SQL modes), layout titles, assistant prompts that assume DBs/Docs; Morph Utils `config.ts` / `App.tsx` Docs deep-links.
- **Morph Engi (Projects):** `nav.ts`, `AppLayout.svelte`, `App.svelte`, `ProjectImportPanel.svelte` (replace with document-oriented generate/publish), resource-files APIs, new or extended AI generate + publish endpoints, schema for markdown/html/publish fields on projects (or a dedicated project-document table).
- **BREAKING (UX):** Docs, Databases, Visual/SQL reports, People, Flow log primary workflows go away for Morph Utils users of these embeds.
