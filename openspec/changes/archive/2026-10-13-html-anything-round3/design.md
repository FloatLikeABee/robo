# Design

## Context

See proposal.md. PR #154 (`aaeaac6`) retries parse and expands on `strings.Count(md, "## ")`, which treats `###` as `##`. Qwen emitted JS backtick JSON (`big2.json`), a 90s deadline (`big6.json`), a 2-`##` note (`big3.json`), and a grill hex table (`grill-draft.html`).

## Goals / Non-Goals

**Goals:** Unwrap backticks; never save an envelope; idea as title fallback; >90s timeout + clean error; real `##` count + scaffold; colour-free grill prompt; hex-table sanitize; 512-token skill budget.

**Non-Goals:** Image-build #114. Changing the compact skill text. A new JSON API.

## Self-grill (Superpowers)

Three alternatives per fork. Chosen: A2 + B3 + C2 + D1.

### A. Backtick JSON

1. **Save raw on parse fail** (current). Produces Untitled + envelope. Rejected.
2. **Convert backtick values to JSON strings, then unmarshal.** Fixture is the real `big2` envelope. **Chosen.**
3. **Regex-extract `"title"` / markdown between fences.** Fragile on nested quotes. Rejected as the only fix.

### B. Timeout

1. **Keep 90s.** Qwen 3B already died. Rejected.
2. **Drop the inner ctx and rely on 300s HTTP.** Request ctx can still be 90s from a proxy. Incomplete.
3. **5-minute generate timeout; map deadline to `the model timed out; retry`.** Matches story-AI’s long wait. **Chosen.**

### C. Three sections

1. **Prompt only.** Already asked; `big3` still has 2 `##`. Rejected as the only fix.
2. **Count line-start `## ` only; regenerate once; scaffold from the idea.** **Chosen.**
3. **Promote `###` to `##`.** Rewrites the model’s outline. Rejected.

### D. Grill hex table

1. **Drop html-anything and hex/card lines from WritePrompt; ask for headings/bullets; strip `| #hex |` rows; keep card CSS in `app.html`.** **Chosen.** Grill is markdown, not HTML.
2. **Keep the skill, append “no hex in the body”.** The skill still lists hex; Qwen copied it. Rejected as the only fix.
3. **CSS `display:none` on hex cells.** Leaves junk in the stored note. Rejected.

## Decisions

1. `repairJSONBacktickStrings` before unmarshal. If fallback text looks like a `{...}` envelope, use the idea for body and title.
2. `bigNoteGenerateTimeout = 5 * time.Minute`. `wrapBigNoteAIError` rewrites deadline errors.
3. `countH2` = lines whose trim has prefix `## ` and not `###`. `thin` / `expand` use that.
4. `WritePrompt` drops `WithHTMLSkill` and every hex/card sentence. `SanitizeDraft` flattens hex-first markdown tables.

## Risks / Trade-offs

- [A backtick in a real JSON string is rewritten] → Only convert `\`` tokens outside JSON strings, after `:`.
- [5 minutes ties up a worker] → One generate; timeout is better than a false 90s fail.
- [Grill without html-anything drifts from Ge] → Card CSS on the page is the chrome; body stays facts.

## Migration Plan

Code and tests. Restart Dorph. Rollback is revert.

## Open Questions

None.
