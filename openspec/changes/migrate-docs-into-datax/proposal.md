## Why

Docs (former Academi) is still a separate Morph Utils module, while DataX already owns “Data access” for structured files and AI over data. Users need one place for both **data** (CSV/JSON tables) and **content documents** (PDF/TXT), with Docs AI and Board-style publishing living inside DataX instead of a second Utils app.

## What Changes

- **Migrate Docs into DataX (Data access)**: Move the Docs product surface (AI study/docs chat, document library, Board) into SharpReport / DataX as a dedicated Docs area under Data access. (**BREAKING** for standalone Docs Utils embed URLs and the Academi/Docs Utils module.)
- **Remove standalone Docs Utils module**: Drop `docs` from Morph Utils nav/config once DataX hosts Docs; legacy `/docs` / `/academi` Utils paths redirect into DataX Docs (or DataX root with Docs deep-link).
- **Docs AI layout**: Session history list on the **left**; active chat on the **right** (two-pane chat, not overhead tabs alone).
- **Docs uploads ≠ data uploads**: Docs attachment/upload accepts **content** files (PDF, TXT, and existing docs types as needed). DataX data import / data tables remain **CSV / JSON** (and related tabular formats). UI and API clearly separate the two pipelines.
- **Board → published HTML docs**: Replace Docs Board (community-style feed) with a **ComposerX-like publish** flow: publish an HTML page that displays the document, with **optional user prompt** for AI analysis included in or alongside the published page.

## Capabilities

### New Capabilities

- `datax-docs-workspace`: Docs lives inside DataX Data access (nav, routes, document library) instead of a separate Utils module.
- `datax-docs-ai-sessions`: Docs AI two-pane UX (session list left, chat right) with content-file uploads distinct from DataX tabular imports.
- `datax-docs-publish`: Publish document as public/shareable HTML with optional AI analysis prompt (Board successor).

### Modified Capabilities

- (none archived under `openspec/specs/`; prior `morph-utils-docs` lived only under change `docs-projects-unify-morph-data` and is superseded by folding Docs into DataX.)

## Impact

- **SharpReport / DataX** (`SharpReport/`): New Docs nav section/routes; AI sessions UI; content upload APIs vs existing data-table import; publish HTML + optional analysis endpoints/UI.
- **academi**: Source of Docs AI / documents / Board behavior to port or proxy; standalone Academi UI may be retired from Utils embeds.
- **morph-utils**: Remove or redirect `docs` module; login/settings copy; embed URL env for Docs → DataX.
- **ComposerX**: Reference pattern for publish HTML + optional AI prompt (not a hard dependency; may reuse patterns/APIs).
- **start-all / docs**: Service map updated so Docs is reached via DataX, not a separate Utils iframe target.
