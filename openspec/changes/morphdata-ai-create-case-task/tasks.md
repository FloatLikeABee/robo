## 1. Draft API

- [x] 1.1 Add `POST /api/tran/case-tasks/ai-draft` (JSON `{prompt}` and multipart `prompt` + `file`) that requires at least one source, extracts file text (txt/md/pdf/csv/xlsx), concatenates prompt + file, and returns `{title, description, start_at, end_at, location, detail}` without inserting a CaseTask
- [x] 1.2 Prompt Morph AI for that object: required title + JSON object `detail`; dates and `location.area` only when present in the source; drop invalid coordinates; 400 empty extract / unsupported type / no text PDF; 503 if AI is missing; 502 if title+detail cannot be parsed
- [x] 1.3 Register the route, add `tranEndpoints.caseTaskAiDraft`, and list it in `management_chat.go`

## 2. Handler tests

- [x] 2.1 Tests: reject neither-source; reject unsupported file; reject empty PDF extract; 503 when AI unset; parse a mocked model JSON into title/description/dates/location/detail without writing SQL

## 3. Tasks drawer UI

- [x] 3.1 On Create/Edit case/task, add Generate from material (prompt, file picker, Generate) at the top of the drawer
- [x] 3.2 Apply a successful draft into title, description, start/end, location (map chip/preview), and Details JSON; confirm before replacing a non-empty form; do not save until existing Save; keep manual create

## 4. Verify

- [x] 4.1 `go test` for the new handler tests; smoke Create case/task: prompt-only, file-only, both, cancel-without-save, manual save still works
