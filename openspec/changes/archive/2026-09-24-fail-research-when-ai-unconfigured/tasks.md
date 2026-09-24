# Tasks

## 1. Research worker terminal status

- [x] 1.1 Add a Go test that runs a research job whose writing calls all fail and assert status `failed`, non-empty `error_text`, and empty `markdown_content`. Run it and confirm it fails on today’s `complete` status.
- [x] 1.2 Add a Go test for partial failure (some round writes fail, synthesis returns a thesis) asserting status `complete`, `error_text` naming the failed rounds and the count, and a conclusion that is the thesis. Run it and confirm it fails before the worker change.
- [x] 1.3 Add a Go test where at least one round write succeeds and synthesis returns nothing, asserting status `failed`, empty conclusion, and the successful piece still stored. Run it and confirm it fails before the worker change.
- [x] 1.4 Add a Go test that runs a job with a nil AI service and no hook, asserting status `failed`, `error_text` `AI is not configured (set MORPH_AI_API_KEY)`, and empty conclusion, with no network call. Run it and confirm it fails before the worker change.
- [x] 1.5 Implement the worker rules from design.md (ready check, not-configured abort, successful-piece synthesis, no concatenation fallback, `failed` vs `complete`) and re-run the research handler tests, including the existing success and legacy twenty-round tests, until they pass.

## 2. Create and publish guards

- [x] 2.1 Add a Go test that `POST /api/tran/research` with a valid prompt and no AI returns 503, mentions `MORPH_AI_API_KEY`, and inserts no row; empty prompt and a bad file type still return 400. Run it and confirm the valid-prompt case fails before the guard.
- [x] 2.2 Add a Go test that publishing a `failed` job returns 409 and leaves `published_slug` empty. Run it and confirm it fails before the guard.
- [x] 2.3 Implement the create-time ready check and the publish 409, point the existing prompt-only create test at the test hook, and re-run `go test` for the research handler tests until they pass.

## 3. Research UI

- [x] 3.1 Add a pure helper and a frontend unit test for failed vs complete-with-gaps presentation (list text, alert severity, empty thesis, publish allowed). Run the unit test and confirm it fails before `Research.js` uses the helper.
- [x] 3.2 Use the helper in `Research.js` so a failed job shows the cause and no thesis editor, a complete job with `error_text` shows a warning and the thesis, and publish is disabled when status is `failed`. Re-run the unit test until it passes.

## 4. Integration check

- [x] 4.1 Run `go test` for the Morph modules this change touches (`./handlers/` and `./ai/`) with `MORPH_AI_API_KEY` unset, and `CI=true npm run build` plus the Research unit test in `morph/frontend`. Confirm all of those commands exit 0.
