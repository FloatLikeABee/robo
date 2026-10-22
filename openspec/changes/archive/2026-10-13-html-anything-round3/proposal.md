# Proposal

## Why

PR #154 saved 5 of 6 Qwen 3B big notes, but one stored the raw JS-style JSON envelope as the body with title `Untitled note`, one timed out as `context deadline exceeded`, and only 3 of 6 had three real `##` sections (`###` was counted as `##`). The grill draft is a markdown table of hex codes.

## What Changes

- Unwrap backtick-quoted JSON values. Never store a raw envelope as the note. Empty title becomes the idea.
- Big-note generate uses a longer Ollama timeout and returns a clean retryable error on deadline.
- Count only `##` headings (not `###`). If still thin after one regenerate, scaffold three sections from the idea.
- Grill write prompt has no colour codes and asks for plain structured markdown. `#111827` card chrome is applied at render. Hex tables in a draft are stripped.
- Keep the 512-token compact skill budget.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `ai-json-envelope`: Parser unwraps backtick-quoted values; generate must not store a raw JSON envelope; empty title uses the idea.
- `dorph-generated-document-theme`: Heading count is real `##` only; timeout is longer with a clean retry error.
- `note-ai-grill`: Write prompt is colour-free plain markdown; hex tables are sanitized; cards stay `#111827` at render.
- `html-anything-base-skill`: Grill write no longer injects the compact skill (it is markdown, not HTML). The 512-token budget is unchanged.

## Impact

`dorph/handlers` parse + generate, `dorph/grill` prompt/sanitizer. No API shape change. Image-build #114 stays out of scope.
