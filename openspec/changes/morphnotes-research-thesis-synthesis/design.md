## Context

See proposal.md for motivation. Spec: `specs/morphnotes-research/spec.md`.

Today `researchRoundCount = 20` in `morph/handlers/tran_research.go`. After the loop, `assembleResearchMarkdown` concatenates `## Round N` blocks, then one LLM call: “Refine this stepwise… Keep round order (Round 1…Round 20).” Tests assert the stored Markdown contains `Round 1` and `Round 20`. UI copy in `Research.js` uses `round_count` (20) and “20 online rounds”. `round_count` is not stored on the row; GET always reports the process constant.

Still: one in-process job, mutex, `webresearch.Gather`, verifier = second LLM call, SQLite pieces, Morph-native HTML publish, dark-only MorphNotes.

## Goals / Non-Goals

**Goals:**

- New jobs: five verified rounds.
- After the last round: thesis synthesis + critic polish into stored `markdown_content` / `html_content`.
- Persist each job’s round target so a twenty-round job already running does not stop at five.
- Tests and UI match five + synthesis (not round-order headings).

**Non-Goals:**

- Changing ingest, file types, cancel, publish, MarkdownEditor, or nav.
- Parallel jobs, extra web backends, or OS subagents.
- Rewriting historical conclusions already stored on complete jobs.
- Theme switch.

## Decisions

### 1. Five rounds, not a slider

**Choice:** `researchRoundCount = 5` for newly created jobs. Not operator-configurable.

**Why:** User said twenty is too much. Five still gives distinct angles without a long wait. A slider is another UI and unspecified.

**Alternative:** 8 or 10 — more wait, same synthesis problem. **Alternative:** 3 — thin source set for a thesis-quality absorb. **Alternative:** Operator picks N — not asked.

### 2. Store `round_target` on the job row

**Choice:** SQLite column `round_target INTEGER NOT NULL DEFAULT 5` via `sqliteAddColumnIfMissing`. `INSERT` uses 5. GET/list expose it as `round_count`. Worker loops `1..round_target` for that id. Backfill: `UPDATE research SET round_target = 20 WHERE current_round > 5 OR id IN (SELECT research_id FROM research_piece GROUP BY research_id HAVING COUNT(*) > 5)`.

**Why:** Resume and in-flight jobs must not adopt the new constant. GET `20/5` on an old complete job would be wrong.

**Alternative:** Always use the process constant — rejected (breaks in-flight 20-round jobs and old complete rows).

### 3. Two LLM passes: compose, then critic

**Choice:** Status `refining`. Pass 1 (compose): all piece markdown + verification + original prompt; instruct a doctoral-thesis / professional-essay rewrite — structure by argument, absorb essence, drop failed/unverified claims or mark them uncertain, no Round N outline. Pass 2 (critic): the compose draft + the same source pack; instruct a ruthless editor to fix structure, redundancy, holes, and voice; output markdown only. Store pass 2. If pass 2 fails, store pass 1 if non-empty. If both fail, last-resort concatenate (emergency only; tests cover the success path, not this fallback as the product bar).

**Why:** One “keep round order” refine is what we are deleting. Two calls is the smallest “must be perfect” bar without a third model family.

**Alternative:** One mega-prompt — cheaper, easier to leave as a polished concat. **Alternative:** Human outline step — extra UI, not asked.

Ceiling: synthesis context is O(sum of five pieces). Upgrade: map-reduce if pieces grow huge.

### 4. Rounds tab stays; Markdown tab is synthesis only

**Choice:** Do not change list/detail chrome except copy (`5`, not `20`) and `round_count` from the row. MarkdownEditor still edits stored conclusion.

**Why:** Spec keeps pieces visible; conclusion is a different artifact.

## Risks / Trade-offs

- [Synthesis invents structure and drops a good round] → Critic pass + verification text in context; operator can still read Rounds and Raw-edit the conclusion.
- [In-flight jobs vs new constant] → `round_target` column + backfill.
- [Tests still assert Round 1 / Round 20] → Rewrite stubbed-run assertions: five pieces; stored MD must not be round-heading concatenation; stub compose/critic prompts by distinctive prefixes.
- [LLM quality is not “perfect” in a formal sense] → Two-pass prompt is the product bar; not a human thesis committee.

## Migration Plan

1. Add `round_target`, backfill, ship worker + tests, then UI copy.
2. Rollback: restore 20 and old refine prompt; column can remain unused.
3. Do not rewrite already-complete `markdown_content`.
