## Why

A MorphNotes Research job whose model calls all fail (for example `MORPH_AI_API_KEY` is unset) still stores concatenated “Round N failed” markdown and finishes with `status=complete`. Operators read that as a finished thesis.

## What Changes

- A Research job with no usable AI, or whose round writes all fail, ends as `failed` with a user-visible reason in `error_text`. It does not store a fake thesis.
- Creating a job when AI is not configured returns HTTP 503 and does not insert a row. A job that is already running (including resume) fails the same way if AI is still not configured.
- Partial round failure stays `complete` only when thesis synthesis succeeds. `error_text` names the failed rounds, and the UI shows that warning next to the thesis.
- If synthesis does not produce a document, the job is `failed`. The Markdown tab is not a concatenation of failed rounds. Successful pieces stay on the Rounds tab.
- The Research list and detail view show `failed` and the reason.

## Capabilities

### New Capabilities

- `morphnotes-research`: Terminal status of a MorphNotes Research job is honest when AI is missing, every round fails, or only some rounds succeed.

### Modified Capabilities

- (none — `openspec/specs/` has no `morphnotes-research` baseline)

## Impact

- MorphNotes Research worker and HTTP create: `morph/handlers/tran_research.go`, tests in `morph/handlers/tran_research_test.go`.
- A configured-key check on the existing Morph AI service (`morph/ai`), without editing `pkg/morphai` (open provider-abstraction work).
- Research UI: `morph/frontend/src/pages/admin/Research.js` at `/morphdata/research`.
- No new secrets, no change to thesis quality when a key is present and rounds succeed.
