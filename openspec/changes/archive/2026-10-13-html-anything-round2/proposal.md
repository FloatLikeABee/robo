# Proposal

## Why

PR #152 made the AI JSON parser tolerate raw newlines, but a Qwen2.5-Coder-3B re-run still parsed big notes only 2 of 6 times (fenced replies, non-JSON replies, missing commas). The same run wrote a full HTML grill page that converted 30k RMB to $30,000, invented Tuesday times, and filled a violet card with mint text.

## What Changes

- Big-note generate never hard-fails on a parse error: strip fences, take a balanced `{...}` block, repair missing commas, retry the model once, then fall back to plain markdown.
- Grill drafts are body fragments: sanitize doctype, `html`/`head`/`body`, and `<style>`.
- Grill drafts stay on the interview answers: keep the user's units and currency; do not invent dates or places.
- Grill chrome uses `#111827` cards with accent borders (not mint text on a violet fill).
- Big notes have a minimum depth (several `##` sections), so one paragraph is not enough.
- Keep the 512-token compact html-anything budget.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `ai-json-envelope`: Parse also repairs missing commas between properties; big-note generate retries once then falls back to markdown and MUST NOT return a parse error.
- `note-ai-grill`: Draft is a body fragment grounded on answers; grill cards are `#111827` with accent borders.
- `morph-generated-document-theme`: Big-note generate requires a minimum section depth.

## Impact

`morph/handlers` parse + big-note generate, `morph/grill` sanitizer/prompts/`app.html`, optional GrillModal chrome. No API shape change. No new dependencies. Image-build runbook (#114) stays out of scope.
