## 1. Backend extract (no persist)

- [x] 1.1 Add a failing test: `POST /api/tran/generic-data/extract` with a CSV (or Markdown) file returns JSON + markdown drafts and does **not** insert a `generic_data` row
- [x] 1.2 Add a failing test: extract with AI down still returns parse-seeded drafts for tabular/JSON/Markdown; scanned/empty PDF with AI down returns an error and no row
- [x] 1.3 Implement extract: parse via `parseGenericDataImport`, seed JSON/Markdown, optional Morph AI refine, JSON 200 with `{ json, markdown, title, source_type, filename }`, never INSERT
- [x] 1.4 Register the route next to import; keep `POST /api/tran/generic-data/import` for now but unused by the File tab

## 2. Save reviewed drafts

- [x] 2.1 Add a failing test: `POST /api/tran/generic-data` with extract-shaped `detail` (`payload` or columns/rows plus `content_markdown`) persists both JSON and Markdown
- [x] 2.2 Reject invalid JSON on Save (400, no row) if not already covered
- [x] 2.3 Confirm `AnalyzeGenericData` includes `content_markdown` in the excerpt so Run AI analysis works after file save

## 3. Import dialog File tab

- [x] 3.1 File tab: Extract (not Import) → editable JSON + Markdown → **Save** via create; Cancel/close without Save creates no row
- [x] 3.2 After Save, open the new record so Content shows JSON/Markdown and **AI analysis** / Run AI analysis is available (do not auto-run analyze)
- [x] 3.3 Invalid JSON blocks Save in the UI with a validation message

## 4. Verify

- [x] 4.1 `cd morph && go test ./handlers -count=1 -timeout 120s` covering extract + save
- [x] 4.2 Browser: MorphNotes Generic data → import file → extract JSON+MD → edit → Save → Run AI analysis; cancel after extract leaves the list unchanged
