## Context

See proposal.md for motivation. Spec: `generic-data-file-review-save`.

Today MorphNotes Generic data **File** tab uploads via `POST /api/tran/generic-data/import`, which parses and **inserts immediately**. **Paste text** already extracts JSON (`POST /api/tran/extract-json`), lets the user edit, then **Save** (`POST /api/tran/generic-data`). Saved records have **Run AI analysis** (`POST /api/tran/generic-data/:id/analyze`) on the detail drawer. `parseGenericDataImport` already yields tables, JSON payload, or `content_markdown` for PDFs.

## Goals / Non-Goals

**Goals:**
- File path matches paste: no row until Save.
- One extract action fills editable JSON and Markdown from the file (parse + Morph AI).
- Save writes both into the existing Generic data detail document (JSON payload + `content_markdown`).
- Analysis stays on the saved record; do not auto-run on extract or save.

**Non-Goals:**
- Changing paste-text (already review-then-save).
- OCR for image-only PDFs.
- New tables; keep `generic_data` + entity detail store.

## Decisions

### 1. Preview extract, then create — do not reuse import-as-save
- **Choice:** Add `POST /api/tran/generic-data/extract` (multipart `file`) that parses the file, asks Morph AI for a JSON object and a Markdown document, returns `{ json, markdown, title, source_type, filename }` with **no INSERT**. Save uses existing `POST /api/tran/generic-data` with `detail` containing `payload`/`columns`+`rows` as appropriate plus `content_markdown`. Keep `POST .../import` unused by the File tab (or make it call extract internally only if a client still needs it — UI must not save on extract).
- **Rationale:** Current import cannot show drafts before persist. Same pattern as paste + Event Logs review-then-save.
- **Alternatives:** Query flag `preview=1` on import — easy to miss and still easy to persist by accident.

### 2. JSON + Markdown from parse seed + one AI pass
- **Choice:** Parse first (`parseGenericDataImport`). Seed Markdown from PDF/MD text or a markdown table of CSV/Excel rows; seed JSON from parsed payload/rows. Then one Morph AI call to refine both (or JSON via existing extract-json prompt plus a short MD rewrite). If AI fails after a successful parse, still show parse-based drafts so the user can Save without AI.
- **Rationale:** User asked for AI extract of JSON and MD; parse-only fallback avoids blocking Save when AI is down for tabular files. Spec “extract unavailable” applies when there is nothing to draft (AI required and parse empty, or AI down **and** no parse seed). Prefer: AI down + parse seed → still show drafts; AI down + PDF with no text → error.
- **Alternatives:** Always require AI — worse for CSV when the key is already structured.

Clarify for implementers: **AI unavailable** with a parseable CSV/xlsx/json/md/pdf-text-layer still returns parse-seeded drafts. **AI unavailable** with nothing extractable (empty/scanned PDF) errors and creates no record.

### 3. Analysis stays on the record
- **Choice:** After Save, open the new row’s detail (Content + AI analysis tabs). Do not call analyze during extract/save. Reuse `AnalyzeGenericData` and include `content_markdown` in the analysis excerpt if missing today.
- **Rationale:** Spec forbids auto-analysis; existing UI already has the button.

## Risks / Trade-offs

- [Large files / AI latency] → Keep 20MB import cap; truncate tables as today; timeout already 5 minutes on upload.
- [Detail shape vs Content tab] → Save `content_markdown` plus existing `payload`/`columns`/`rows` so the current Content renderer still works.
- [Old clients posting `/import`] → Leave the route for one release; File tab stops using it.

## Migration Plan

1. Ship extract + File dialog Save; stop calling `/import` from the dialog.
2. Rollback: point File tab at `/import` again (loses review).

## Open Questions

None.
