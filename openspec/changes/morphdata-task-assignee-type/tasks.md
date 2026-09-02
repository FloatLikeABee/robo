## 1. Create insert defaults

- [x] 1.1 In `CreateCaseTask`, always include `assignee_type` and `assignee_id` on INSERT (`''` and `0` when unassigned; first parsed assignee when present)
- [x] 1.2 Keep stripping client `assignee_type` / `assignee_id` so JSON null never binds as SQL NULL

## 2. Tests

- [x] 2.1 Add a test that creates a title-only task against SQLite `CaseTask` with `assignee_type TEXT NOT NULL` (no DEFAULT) and expects 200 + an id
- [x] 2.2 Assert blank title still returns a validation error and inserts no row

## 3. Verify

- [x] 3.1 Save a new Morph Data Tasks row with no assignee in the UI; confirm it stores without the 1299 constraint error
