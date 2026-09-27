# Design

## Context

See proposal.md for why a Research job must not finish `complete` when AI never produced a thesis. The worker is `runResearchJob` in `morph/handlers/tran_research.go`. Each round calls `researchLLM`. A write error is stored on the piece as status `error` and the text `Round N failed: …`. After the loop the worker always sets the job to `complete`. `synthesizeResearchConclusion` falls back to `assembleResearchMarkdown`, which concatenates those failure strings into `markdown_content`. `error_text` already exists on the row and in the JSON payload and is never written.

`researchLLM` returns “AI service not configured” only when `aiService` is nil. In the process, `ai.New` still returns a service when the key is empty, and `pkg/morphai` then fails each call with “MORPH_AI_API_KEY is not configured”. Tests inject `researchLLMHook` and do not need a key. `pkg/morphai` is owned by in-flight provider work and MUST NOT be edited.

The Research UI (`morph/frontend/src/pages/admin/Research.js`) renders `status` and `markdown_content` and does not read `error_text`.

## Goals / Non-Goals

**Goals:**

- Terminal status `failed` when AI is not configured, every round write fails, or synthesis produces no conclusion.
- No fake thesis in `markdown_content` on those paths.
- Partial success (some round writes succeed and synthesis returns a document) stays `complete`, with `error_text` naming the failed rounds, and the UI shows a warning plus the thesis.
- Create fails fast with HTTP 503 when AI is not configured, before a row is inserted.
- Go tests simulate total and partial model failure without a real key.

**Non-Goals:**

- Thesis quality when a key is present and rounds succeed.
- Rewriting historical `complete` rows that already stored failure text.
- New job-status values beyond `failed`.
- Changes under `pkg/morphai`.
- Watermarking the published HTML page with the round-failure warning.

## Decisions

### 1. Job status `failed`, reason in `error_text`

The stored status is the string `failed`. The cause is `error_text`, already returned as JSON. Piece rows keep their own status (`ok`, `unverified`, `error`).

A round write is successful when its piece status is `ok` or `unverified`. Verification failure alone does not fail the round or the job. A round is failed when the writing call returns an error (piece status `error`).

**Alternatives rejected:**

- Status `error` for the job. Piece rows already use `error`. A different job word matches `cancelled` and reads clearly on the status chip.
- Status `partial` when some rounds fail. A third terminal status forces resume, publish, and chip rules for a case that still has a thesis. The list and a warning can show the gaps without a new state.
- Leave status `complete` whenever the loop finishes, and only fill `error_text`. That is the current bug: operators trust the status chip.

### 2. What “done” means

| Outcome | Status | Conclusion | `error_text` |
| --- | --- | --- | --- |
| AI not configured at create | no row (HTTP 503) | — | response `error` |
| AI not configured when the job runs, including after an earlier round succeeded | `failed` | empty; pieces already stored stay | `AI is not configured (set MORPH_AI_API_KEY)` |
| Every round write failed | `failed` | empty | rounds failed, with the model error |
| Some writes failed, synthesis returned a thesis | `complete` | that thesis | how many rounds failed, and which |
| Some writes succeeded, synthesis returned nothing | `failed` | empty | thesis could not be synthesized; successful pieces kept |
| Every write succeeded and synthesis returned a thesis | `complete` | that thesis | empty |
| Operator cancelled | `cancelled` | unchanged by this change | unchanged |

Synthesis receives only successful pieces. It MUST NOT fall back to concatenating pieces. A not-configured error stops the remaining rounds instead of repeating the same failure five times. Any other write error continues, so a later round can still succeed.

Cancellation wins: a job already marked `cancelled` is not overwritten with `failed`.

### 3. Where “AI is configured” is decided

`researchAIReady` is true when `researchLLMHook` is set (tests) or when `aiService.Configured()` is true. `Configured` on `morph/ai.AIService` delegates to the existing `morphai.Client.Configured()` method. No new `pkg/morphai` API.

Create checks this after prompt and file validation and before insert. Invalid prompt and bad file types stay HTTP 400. `runResearchJob` checks again before doing model work, so a resumed `running` job cannot finish `complete` after the key disappears.

A model error whose text contains “not configured” is treated as the same not-configured failure so a key that the client still considers set cannot spin all five rounds. Transient HTTP failures do not match that phrase and do not abort the rest of the run.

### 4. UI

- Failed list rows show `failed` and a shortened `error_text`.
- Detail shows an error alert for `failed`, and a warning alert when `complete` still has `error_text`.
- A failed job with an empty conclusion shows “No thesis was produced.” in place of the editor.
- Publish is disabled in the UI and `POST …/publish` returns HTTP 409 while status is `failed`, so an empty page cannot be published by bypassing the button. The not-configured sentence shared by the 503 body and `error_text` is `AI is not configured (set MORPH_AI_API_KEY)`. Partial success copy includes the counts, for example `2 of 5 rounds failed (2, 4)`.

Presentation rules live in a small pure helper so the UI contract is unit-tested without a browser.

### 5. Design review

Proposer and reviewer walked the handler, the existing tests, and the Research page before coding.

Challenges and resolutions:

- “`complete` plus a warning hides partial failure in the list.” The list secondary line MUST include the failure count, not only the word `complete`. Detail uses a warning alert. A `partial` status was rejected in decision 1.
- “One successful round out of five should still be `failed`.” A numeric cutoff is arbitrary and would hide a real thesis. The warning states the count (`N of M rounds failed`).
- “Failing only at create misses invalid keys and resumed jobs.” Create returns 503 when the client has no key. The worker still classifies total call failure and a not-configured run as `failed`.
- “Matching the words ‘not configured’ is brittle if provider errors change.” The primary check is `Client.Configured()` before any call. The phrase match only aborts leftover rounds. If the phrase changes, five failed writes still end `failed` because every write failed.
- “The concatenation fallback is a useful draft if synthesis fails.” That draft is the fake thesis this change removes. Successful pieces stay on the Rounds tab.
- “HTTP 503 will break the prompt-only create test, which uses a nil AI service.” That test’s intent is a stored five-round job. It will install the test hook so AI counts as ready, and a new test covers 503 with no hook and no key.
- “Disabling publish only in the client is bypassable.” The publish handler returns HTTP 409 for status `failed` and does not set a public path.
- “A not-configured error after round 1 could be reported as a synthesis failure and hide the real cause.” That run skips synthesis and uses the not-configured sentence. Pieces already stored stay; the conclusion stays empty. The synthesis-failure sentence is only for a configured model that returns no thesis.
- “Putting the warning into the thesis markdown would make the public page honest.” It would also mix operator diagnostics into the document and fight the empty-conclusion rule. The warning stays in the app. Accepted gap: a published page does not repeat `error_text`.

## Risks / Trade-offs

- [Historical `complete` rows that already contain “Round N failed”] → left as stored. This change only affects jobs that run after it. No schema migration.
- [A thesis from a minority of rounds can still say `complete`] → `error_text` and the UI warning include the failed-round count.
- [Published HTML omits the round-failure warning] → accepted. The in-app warning is the contract. Publish of `failed` jobs is rejected.
- [Global `researchLLMHook` shared by tests] → each test restores the previous hook. Tests stay sequential, as they are today.
- [`pkg/morphai` provider work changes the not-configured error string] → ready-check uses `Configured()` and does not depend on that string for the create or start-of-job path.

## Migration Plan

No schema change. `error_text` is already on `research`. Deploy the Morph binary and the Research frontend together. Rollback is reverting the change; in-flight jobs keep their rows.

## Open Questions

None. Status rules above are the contract the tasks implement.
