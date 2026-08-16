## Context

Morph Utils embeds **DataX** (`SharpReport`, `VITE_DATAX_URL`) and **Projects** (`morph-engi`, `VITE_PROJECTS_URL`). DataX nav today: Docs, Data tables, Databases, Data reports (Visual | SQL | file). Docs is failing for users. Projects (Morph Engi) nav today: Projects, Files, People, Flow log — plus an AI import that proposes nested site logs / people / money rows. See `proposal.md` for why that no longer matches the product.

Prior art: Morph Data Timelines / Big Notes (MD+HTML+publish), Morph Engi `project-imports` analyze/confirm, SharpReport file report builder mode.

## Goals / Non-Goals

**Goals:**
- Strip Data Access to useful surfaces without Docs/Databases; Data reports = file upload only.
- Collapse Morph Engi to Projects + Files; Projects = AI MD/HTML docs from file/paste with publish; paste also lands in Files.
- Update Morph Utils shell copy and remove Docs deep-links.

**Non-Goals:**
- Deleting Data tables from DataX.
- Physically deleting all Docs/Databases backend tables in v1 (hide/redirect UX is enough; hard delete APIs can follow).
- Keeping People/Flow log CRUD behind a hidden flag in v1 primary UX.
- Migrating old Engi job rows into MD documents automatically.
- Requiring ComposerX to host publish (in-app public HTML like Big Notes/Timelines is preferred).

## Decisions

### 1. DataX: hide and redirect, default Data reports to file mode
- **Choice:** Remove Docs and Databases from `Sidebar.svelte`; redirect `/docs*`, `/databases*` to `/data-tables` or `/reports/builder`. In `reports/builder/+page.svelte`, remove Visual/SQL mode toggles and UI branches; keep only file mode (default).
- **Why:** Fastest UX fix for Morph Utils; Docs is already broken; Databases contradict “no linking.”
- **Alternatives:** Soft-gate with feature flags — unnecessary complexity for an explicit removal.

### 2. Morph Utils shell Docs deep-link removal
- **Choice:** Drop `DATAX_DOCS_URL` / `isLegacyDocsPath` special case; map legacy `docs`/`academi` to `datax` embed root only. Update Data Access and Project descriptions in `config.ts`.
- **Why:** Prevents Utils from sending users into a removed Docs surface.

### 3. Projects: document-centric model instead of nested job import
- **Choice:** New (or evolved) project document fields: `title`, `markdown_content`, `html_content`, publish slug/path, source summary; create API accepts multipart `file`(s) + `paste` with ≥1 required. AI returns organized markdown; server builds HTML; publish endpoint serves public HTML.
- **Why:** Matches Timelines/Big Notes pattern users already understand; matches “requirements → organized project contents.”
- **Alternatives:** Keep structured site_logs/people/flow confirm — rejected by product direction.

### 4. Files library reuses resource-files
- **Choice:** Keep/extend `resource_files` (or equivalent) as the Files list. On paste-source create: write a `.md`/`.txt` content file into resource_files (name like `paste-YYYYMMDD-HHMMSS.md`) and link it to the project if linkage exists. Uploaded files also register in the same list.
- **Why:** Spec requires paste to appear in Files; Engi already has a Files tab and upload API.
- **Alternatives:** Separate paste-blob table — more surfaces for the same list job.

### 5. People and Flow log cutover
- **Choice:** Remove from `NAV` and `App.svelte` page branches; leave backend routes unused for now (or return 410 later). Empty-state copy on Projects explains document workflow.
- **Why:** Spec says two modules only; soft retention of APIs reduces migration risk.

### 6. Source file types for Projects AI
- **Choice:** Prefer `.txt`, `.md`, `.pdf` (and existing `.csv` if already supported) for uploads; paste as UTF-8 text. Reuse Engi’s PDF extract path from prior AI import work.
- **Why:** Spec/requirements docs are typically those types; aligns with other Morph ingest flows.

## Risks / Trade-offs

- **[Risk] Existing Engi projects/people/flow data orphaned in UI** → Mitigation: no auto-delete; document that primary UX is new; optional later export.
- **[Risk] DataX Docs/DB APIs still callable** → Mitigation: acceptable for v1; remove nav first; harden later if needed.
- **[Risk] Data reports Visual/SQL users lose workflows** → Mitigation: intentional; file mode remains.
- **[Trade-off] Multi-file upload vs single file** → Allow multiple files in one create if cheap (like current import max), but validate ≥1 source overall.

## Migration Plan

1. Ship DataX nav + reports UI cut + Morph Utils shell copy/deep-link fixes.
2. Ship Morph Engi nav cut to Projects/Files.
3. Ship new project-document generate + publish + paste→Files persistence.
4. Smoke: Data Access has no Docs/Databases; reports file-only; Projects create from file/paste; Files lists paste; publish HTML works.
5. Rollback: revert frontends; unused tables/APIs remain.

## Open Questions

- Exact max file count/size for Projects create (lean toward existing import limits unless product specifies otherwise).
- Whether Data tables remains the default landing route after Docs removal (lean yes).
