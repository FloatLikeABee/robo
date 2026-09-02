## Why

MorphNotes Generic data file import writes a record immediately. Operators need Morph AI to turn the file into editable **JSON** and **Markdown** first, then **Save** after they review, and they still need to **run AI analysis** on the saved material (not auto-save, not auto-analyze).

## What Changes

- File import becomes extract → review/edit → Save, matching paste-text (cancel does not create a record).
- Morph AI produces both a JSON draft and a Markdown draft from the chosen file (CSV, Excel, JSON, PDF, Markdown).
- The import dialog has a **Save** action that persists the edited JSON and Markdown as Generic data content.
- After save, the operator can **Run AI analysis** on that record (existing analysis tab / action). Analysis is not run automatically on extract or save.
- Direct “Import” that saves without a review step is removed from the file tab (operators who only want raw parse can still Save without editing the drafts).

## Capabilities

### New Capabilities

- `generic-data-file-review-save`: Generic data file import extracts JSON and Markdown for review, saves only when the user confirms, and lets them run AI analysis on the saved record.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baselines)

## Impact

- **MorphNotes backend**: Generic data import/preview (likely `POST /api/tran/generic-data/import` preview or a sibling extract route) plus existing `POST /api/tran/generic-data` save and `POST /api/tran/generic-data/:id/analyze`.
- **MorphNotes frontend**: `GenericData.js` import dialog (file tab) and detail **AI analysis** tab.
- **Reuse**: `parseGenericDataImport`, `POST /api/tran/extract-json`, `AnalyzeGenericData`.
- **Out of scope**: Paste-text tab (already extract → Save); Morph AI chat workspace; OCR for scanned PDFs with no text layer.
