# Design

## Context

See proposal.md for why. PR #152 (`7e81a6f`) already strips fences, takes one object, and repairs raw newlines. Generate still returns `parse AI note JSON` / `AI did not return JSON for the note` on the remaining Qwen failures. `SanitizeDraft` only drops Dear/sign-off letters, so the #153 grill HTML (no Dear) is kept. `WritePrompt` already says body-only and no `<!DOCTYPE html>`, but the model still emitted a full page, converted 30k RMB to $30,000, and invented Tuesday times. `app.html` already has `--card: #111827`; `.q` is mint and the model card fill was `#a78bfa`. Evidence: issue #153 tarball (`big2`–`big5` parse errors, `grill-draft.html`) and screenshots.

## Goals / Non-Goals

**Goals:**

- Never return a parse error from big-note generate; retry once then markdown fallback.
- Repair the missing-comma object shape seen in `big4`.
- Grill drafts are fragments grounded on answers; cards are `#111827` + accent border.
- Big notes have ≥3 `##` sections.
- Compact skill stays ≤512 tokens.

**Non-Goals:**

- Changing the public JSON API.
- Image-build / `deploy/README.md` #114.
- Dropping html-anything from grill write.
- Retrying more than once.

## Self-grill (Superpowers)

Three alternatives were grilled for each fork. Recommendation is A3 + B3 + C2 + D3.

### A. Parse and never hard-fail

1. **Repair only** (try every balanced `{...}`, insert missing commas). Still fails on a fenced markdown note with no object. Rejected as the only fix.
2. **Retry the model until JSON appears.** Burns the 3B budget and can loop. Rejected.
3. **Strip fences → first balanced `{...}` (then later siblings if that one fails) → missing-comma repair → retry once → fall back to plain markdown.** Matches the issue. **Chosen.** Transport errors still fail.

### B. Grill fragment and grounding

1. **Prompt-only.** Already said no doctype; the model still wrote `<style>` + `$30,000`. Rejected as the only fix.
2. **Strip wrappers only.** Leaves invented Tuesday times and `$30,000`. Rejected as the only fix.
3. **Strip doctype/html/head/body/style, then reject a draft that changes currency/units or invents a weekday+time or place not in the answers; fall back to `DraftFromFacts`.** Prompt restates keep-units / no invented dates. **Chosen.**

### C. Grill card

1. **Change only the model CSS after the fact.** After we strip `<style>`, the violet fill is gone, but `.q` is still mint and a later HTML fragment can refill `#a78bfa`. Incomplete.
2. **Grill chrome `#111827` + accent border, prompt says card fill is `#111827` and accents are borders, never a violet fill.** Smallest pair that matches the screenshot. **Chosen.**
3. **Stop injecting html-anything on grill write.** Undoes #149. Rejected.

### D. Big-note depth

1. **Prompt only** (“at least 3 `##` sections”). The successful #153 note was still one paragraph. Rejected as the only fix.
2. **Reject thin notes as an error.** Conflicts with never-hard-fail. Rejected.
3. **Prompt for ≥3 `##` sections; if the result is thin, retry once; if still thin, expand from the idea locally.** **Chosen.**

## Decisions

### 1. Shared parse

Keep `parseModelJSON`. After fence strip, walk every balanced `{...}` (not only the first prose brace). After newline repair, insert a comma when a finished object value is followed by `"`. Unmarshal the first candidate that works. Tests use reconstructed Qwen fixtures for the three #153 errors plus the tarball `grill-draft.html` for the sanitizer.

### 2. Generate never hard-fails

`generateBigNoteContent` (or a helper it calls) parses; on parse failure calls the model once more; on second parse failure treats the last reply as markdown (fences stripped). Idea text is the last-resort body if the reply is empty. Depth: prompt asks for ≥3 `##` sections; `thinMarkdown` is fewer than 3 `##` headings or one short paragraph; expand splits the idea into sections. No new AI interface.

### 3. Grill sanitize + ground

`SanitizeDraft`: strip fences, drop `<style>…</style>`, unwrap doctype/html/head/body, keep the inner fragment. Then `ungroundedDraft` if the draft has `$` while answers have RMB/CNY/¥, or a weekday+clock, or a place token that is not in title/answers/description. Ungrounded → `DraftFromFacts`. `WritePrompt` after the skill: keep units and currency; do not invent dates or places; card `#111827` + accent border; fragment only.

### 4. Card chrome

`app.html`: question and draft use `--card` (`#111827`) and an accent border; `.q` is ink, not mint-on-violet. Do not paint a card fill `#a78bfa`.

## Risks / Trade-offs

- [A later `{...}` sibling is the wrong object] → Prefer objects that unmarshal into the dest type; tests pin the Qwen shapes.
- [Grounding drops a good draft that mentions `$` in passing] → Only trip when answers name RMB/CNY/¥ and the draft introduces `$`.
- [Local expand is schematic] → Better than one paragraph; retry gives the model one chance first.
- [Write prompt grows past 512] → Budget is on the compact skill, not the whole write prompt; keep the skill test.

## Migration Plan

Code and tests only. Restart Morph to pick up prompts. Rollback is revert.

## Open Questions

None. Issue #153 lists the five items; the grilled alternatives are resolved above.
