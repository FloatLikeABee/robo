# Design

## Context

See proposal.md for why. PR #150 (`9a32e6c`) already vendors html-anything and injects a compact skill via `pkg/dorphai.WithHTMLSkill`. Local E2E on Qwen2.5-Coder-3B showed five leftover bugs: `json.Unmarshal` fails on raw newlines in strings (`parse AI note JSON: invalid character '\n' in string literal`); the overlay text `cyan #38bdf8` is copied into CSS; grill `WritePrompt` ends with “HTML: self-contained `<!DOCTYPE html>`” so the model writes a Dear-Team letter; `QuestionCap` scores word count and misses prepositional parts (`for` / `in`); empty-topic `GuideExcerpt` is outline + golden rules only; `.meta` still uses `system-ui`. Evidence: issue #151 artifacts (`big.json`/`big2.json` parse errors, `grill-draft.html` invalid CSS + letter, `mcp-guide.log` outline).

## Goals / Non-Goals

**Goals:**

- One shared tolerant JSON parse used by big-note and timeline generate.
- Hex-only accent text in the compact skill and form/grill prompts, still ≤512 tokens.
- Grill asks one question per title part (cap 4) and stores a note body, not a letter.
- Big-note prompt restates the idea and forbids invented features.
- Empty `dorph_guide` names html-anything.
- Generated `.meta` (and sibling chrome) is Arial.

**Non-Goals:**

- Changing the public JSON API for notes.
- A new MCP topic.
- Image-build / `deploy/README.md` #114.
- Dropping the compact skill from grill write.

## Self-grill (Superpowers)

Three alternatives were grilled for each fork. Recommendation is A3 + B1 + C2.

### A. JSON envelope

1. **Move markdown/HTML out of JSON** (title-only object, body after a delimiter). Small models still wrap everything in fences; every caller and the prompt change. Rejected as the only fix.
2. **Retry the model on parse error.** Burns tokens and still fails the same way. Rejected.
3. **Repair then unmarshal: strip fences, take the outermost object, escape raw `\n`/`\r`/`\t` inside strings.** Same call sites keep asking for JSON; fixtures lock the E2E failure. **Chosen.** Keep HTML server-side (already true for big notes).

### B. Grill colour names and letter shape

1. **Hex-only prompt text plus body-only rules after the skill, plus sanitize letter/HTML drafts.** Small models copy the last colour phrase they see and ignore “keep the format”. Prompt plus a sanitizer is the smallest pair that matches the screenshot. **Chosen.**
2. **Post-process CSS only** (`cyan #38bdf8` → `#38bdf8`). Leaves the letter intact. Rejected as the only fix.
3. **Stop injecting html-anything on grill write.** Undoes #149. Rejected.

### C. Question count

1. **Always ask 3 questions.** Over-interviews “buy milk”. Rejected.
2. **Count title parts** (comma / semicolon / `and` / `for` / `in` / `with` / `about` / `versus` / `vs`) and take `max(old score, parts)` capped at 4. “Team offsite plan for 12 people in Hangzhou” → 3. **Chosen.**
3. **Raise the word-count thresholds.** An 8-word three-facet title still scores 1. Rejected.

## Decisions

### 1. Shared parse helper

Add `parseModelJSON(raw string, dest any) error` next to `extractJSONObjectAware` in `dorph/handlers`. Walk extracted object bytes; while `inString` and not escaped, replace raw newline/tab with the JSON escape. Unmarshal the repaired bytes. Big-note and timeline generate call this instead of bare `json.Unmarshal`. Tests use fixtures with raw newlines, fences, and surrounding prose.

### 2. Hex-only overlay

Edit `skills/vendor/html-anything/condensed.md` and the `pkg/dorphai` embed (same-bytes test stays). Wording: `accents #38bdf8 #a78bfa #34d399` — no colour word before a hex. Same treatment in `BuildFormHTMLPrompt`. Keep the 512-token budget test.

### 3. Grill write and sanitize

`WritePrompt` still uses `WithHTMLSkill`, then appends: hex only; output the note body, not a letter, not `<!DOCTYPE html>`. `SanitizeDraft` treats a Dear/Best-regards/`[Your Name]` letter or a full HTML page as chatty and falls back to `DraftFromFacts`.

### 4. Title parts

`QuestionCap` = `min(4, max(existing score, titlePartCount))`. `titlePartCount` splits on `,`, `;`, and those facet words as whole tokens.

### 5. Grounding, guide, header

Big-note prompt adds “use only the idea; do not invent features”. `guideOutline` (or the empty-topic excerpt) includes one html-anything line. `htmldoc` `.meta` / `.theme` / form chrome switch `system-ui` to the Arial stack. Frontend case-task CSS that copies `htmldoc` stays in sync if a test already pins it.

## Risks / Trade-offs

- [Repairing JSON could mis-read an unclosed string] → Depth walk already used; tests include valid JSON unchanged.
- [Part split on `in`/`for` over-asks] → Whole-token only; “buy milk” stays 1.
- [Sanitize drops a real note that says “Dear”] → Require greeting plus sign-off or a full HTML document, not a single word.
- [Hex-only text still over budget] → The change shortens the overlay.

## Migration Plan

Code and tests only. Restart Dorph to pick up prompts. Rollback is revert.

## Open Questions

None. Issue #151 lists the five items; the grilled alternatives are resolved above.
