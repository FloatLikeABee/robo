## 1. Round target and five-round worker

- [x] 1.1 Add `research.round_target` (default 5), backfill rows that already ran more than five rounds to 20, INSERT new jobs with 5. Expose it as `round_count` on GET/list.
- [x] 1.2 Loop the worker to that job’s `round_target` (new jobs: 5). Keep per-round gather, writer, verifier, and error-piece-continue behavior.

## 2. Thesis synthesis

- [x] 2.1 Replace ordered concat + “keep round order” refine with compose then critic LLM passes. Stored Markdown MUST NOT be `## Round N` concatenation. Last-resort concat only if both LLM passes fail.
- [x] 2.2 Tests: stubbed new job yields 5 pieces in order; stored conclusion is a unified document (no Round 1…Round 5 outline); HTML still built; in-flight `round_target=20` still loops to 20.

## 3. Research UI copy

- [x] 3.1 Research.js: progress and helper text use the job’s `round_count` (five for new jobs), not a hardcoded 20. Markdown tab remains the synthesized conclusion; Rounds tab still lists pieces. No theme switch.

## 4. Verify

- [x] 4.1 `go test` research handler tests including the new synthesis assertions.
- [x] 4.2 Browser: new Research job shows 5 rounds, then a rendered conclusion that is not a round-by-round dump; Rounds tab still has the five pieces.
