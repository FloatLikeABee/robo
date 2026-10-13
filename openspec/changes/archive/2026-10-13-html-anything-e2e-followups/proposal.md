# Proposal

## Why

PR #150 loaded html-anything on Morph's HTML paths, but local E2E on Ollama Qwen2.5-Coder-3B still failed: big-note JSON parse died on raw newlines, the grill copied colour names into CSS and wrote a letter instead of a note, one question covered a multi-part title, the big-note invented features, `morph_guide` with no topic omitted html-anything, and generated headers used system-ui.

## What Changes

- Make the AI JSON parser tolerant of raw newlines inside strings, markdown fences, and extra text around the object (shared by big-note and timeline generate).
- Give hex accent values only in injected HTML skill and form/grill prompts (no `cyan #38bdf8`).
- Grill write stays note/TODO body, not a full HTML letter; `QuestionCap` asks enough questions to cover each title part.
- Ground the big-note generate prompt on the idea text and forbid invented features.
- Empty-topic `morph_guide` names html-anything.
- Generated document header (`.meta`) uses Arial, not system-ui.
- Keep the 512-token compact-skill budget.

## Capabilities

### New Capabilities

- `ai-json-envelope`: Parse model JSON envelopes that small local models emit with raw newlines in strings, code fences, and surrounding prose.

### Modified Capabilities

- `html-anything-base-skill`: Compact injection uses hex colour values only; still stays within the 512-token budget.
- `note-ai-grill`: Question count covers each title part; write prompt is hex-only and body-only; drafts that are letters or full HTML pages are rejected.
- `morph-generated-document-theme`: Generated header chrome is Arial; big-note generate is grounded on the idea.
- `morph-for-agents`: Default `morph_guide` excerpt names html-anything.

## Impact

`morph/handlers` JSON extract/parse (big notes, timelines), `morph/grill` prompts and sanitizer, `pkg/morphai` condensed skill + vendor copy, `morph/ai` form HTML prompt, `morph/internal/htmldoc` CSS, `morph/mcp` `GuideExcerpt`. No API shape change. No new dependencies. Image-build runbook (#114) stays out of scope.
