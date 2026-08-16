## Context

See proposal.md for motivation. Docs today is Academi (`academi/`), embedded as Morph Utils module `docs`. DataX is SharpReport (`SharpReport/`) with Data access (databases, data tables, CSV/JSON import) and a global assistant drawer. ComposerX already publishes HTML pages with optional AI over sources. Docs Board is a community feed (`community` screen) with optional file analysis — product wants ComposerX-like publish instead.

Constraints: shared UsersPanel auth; Morph Utils iframe embeds; keep DataX tabular import unchanged; prefer reusing Academi APIs via proxy first if faster than full rewrite.

## Goals / Non-Goals

**Goals:**
- Docs workspace + AI + publish live inside DataX navigation.
- Two-pane Docs AI (sessions left, chat right).
- Clear split: Docs content files (PDF/TXT) vs DataX data files (CSV/JSON).
- Board → HTML publish with optional analysis prompt.
- Remove standalone Docs from Morph Utils primary nav.

**Non-Goals:**
- Merging Docs content documents into DataX data-table rows.
- Rewriting ComposerX or replacing DataX Metabase/SQL features.
- Full deletion of Academi service in one release (proxy/retire later is OK).
- Keeping Board as a social feed.

## Decisions

### 1. Host Docs UI in SharpReport (DataX), not Morph Utils iframe of Academi
- **Choice:** Add DataX routes/nav for Docs (e.g. `/docs`, `/docs/ai`, `/docs/library`, `/docs/publish`) implemented in SharpReport frontend; call Academi or new DataX backends as needed.
- **Why:** One Data access product surface; Utils drops a module.
- **Alternatives:** Keep Academi iframe inside DataX (weaker IA); only redirect Utils Docs URL to Academi (fails “into Data access”).

### 2. Backend: proxy Academi Docs APIs from DataX first
- **Choice:** DataX API proxies Academi chat-sessions, documents, and related endpoints under `/api/v1/docs/*` (or similar), with UsersPanel/session forwarding. Follow-up can nativize storage in DataX DB.
- **Why:** Fastest migration; preserves existing Docs AI quality.
- **Alternatives:** Port all Academi persistence into DataX immediately (higher risk).

### 3. Two-pane AI chrome
- **Choice:** Dedicated Docs AI page: left `sessions` list (create/select/delete), right chat transcript + composer; content attach accepts PDF/TXT (and existing Docs types). Do not reuse DataX data-table attach for this path.
- **Why:** Matches requested layout; avoids mixing with tabular “attach table” assistant context.

### 4. Upload pipelines stay separate
- **Choice:** Docs content upload endpoints validate PDF/TXT (content). Existing `/api/v1/data-tables/import` (CSV/JSON) unchanged. UI copy labels “Documents” vs “Data files”.
- **Why:** Prevents users uploading PDFs into table import and CSVs into Docs content by accident.

### 5. Publish modeled on ComposerX
- **Choice:** Publish flow stores HTML (document rendering + optional analysis block) and public slug/URL, patterned after ComposerX published pages. Optional `analysis_prompt` triggers AI analysis at publish time; empty prompt skips analysis.
- **Why:** User asked for ComposerX-like publishing; Board feed is the wrong model.
- **Alternatives:** Keep Board posts with a “publish” flag (still feed-oriented).

### 6. Utils cleanup
- **Choice:** Remove `docs` from `UTILS_MODULES`; map legacy `docs`/`academi` to DataX embed with Docs path (e.g. `VITE_DATAX_URL/docs`). Update login/settings copy.
- **Why:** Single entry via DataX.

## Risks / Trade-offs

- [Academi still required while proxied] → Mitigate with clear DataX env `ACADEMI_API_BASE_URL`; plan native cutover.
- [Auth mismatch Academi JWT vs DataX] → Mitigate with platform-session exchange already used by Academi/Utils.
- [Publish HTML fidelity for PDF] → Mitigate: embed extracted text/markdown view + download original; optional PDF viewer iframe later.
- [Users confuse Docs vs data import] → Mitigate: separate nav labels and accept-lists; reject wrong types with clear errors.

## Migration Plan

1. Ship DataX Docs nav + proxied library/AI two-pane UI.
2. Ship publish HTML + optional analysis; retire Board as primary UX.
3. Remove Docs from Morph Utils; redirect legacy paths.
4. Follow-up: native Docs storage in DataX; archive Academi service when unused.

## Open Questions

- Public publish URL host: DataX origin vs ComposerX-style dedicated public path (default: DataX `/p/docs/:slug` or equivalent).
- Whether MD/images remain allowed in Docs uploads alongside PDF/TXT (default: allow existing Academi content types, but product messaging emphasizes PDF/TXT vs CSV/JSON).
