## 1. List completeness

- [x] 1.1 Add a failing test: insert 4 `generic_data` rows, `GET /api/tran/generic-data` returns 4 items including every id (newest first)
- [x] 1.2 Add a failing test: create via `POST /api/tran/generic-data` then list includes that id
- [x] 1.3 Fix list scanning so `Next` is not called twice per row (`ListGenericData` / `querySingleRowMapFromRows`)
- [x] 1.4 Confirm `querySingleRowMap` (single-row GET) still works

## 2. Verify

- [x] 2.1 `cd morph && go test ./handlers -count=1 -timeout 120s` covering the new list tests
- [x] 2.2 Browser: MorphNotes Generic data — count rows vs API; save an import and confirm the new row is in the grid and opens the saved content
