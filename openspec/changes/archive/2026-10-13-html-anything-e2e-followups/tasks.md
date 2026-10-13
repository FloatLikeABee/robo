# Tasks

## 1. Tolerant AI JSON parser

- [x] 1.1 Add failing fixtures and tests that `parseModelJSON` accepts raw newlines inside strings, markdown fences, and extra prose, and that valid JSON is unchanged. Verify `go test` in `morph/handlers` fails before the helper exists.
- [x] 1.2 Implement `parseModelJSON` (fence strip, object extract, escape raw newlines/tabs in strings) and switch big-note and timeline generate to it. Verify the new tests pass and existing handler tests still pass.

## 2. Hex-only skill and 512-token budget

- [x] 2.1 Add a failing test that the compact skill and form HTML prompt contain `#38bdf8` but not `cyan #38bdf8`, and that the 512-token budget test still holds. Verify they fail.
- [x] 2.2 Change vendor `condensed.md`, the morphai embed, and `BuildFormHTMLPrompt` to hex-only accents. Verify `go test` in `pkg/morphai` and `morph/ai` pass.

## 3. Grill questions, body-only, no letter

- [x] 3.1 Add failing tests: `QuestionCap("Team offsite plan for 12 people in Hangzhou")` is 3–4; write prompt has hex-only, body-only, and no letter; `SanitizeDraft` drops a Dear-Team HTML letter. Verify they fail.
- [x] 3.2 Implement title-part counting, write-prompt rules after the skill, and letter/HTML sanitizer. Verify `go test` in `morph/grill` pass and `buy milk` still caps at 1.

## 4. Grounded big-note, guide, header

- [x] 4.1 Add failing tests that the big-note generate prompt forbids invented features and includes the idea, that `GuideExcerpt("")` contains `html-anything`, and that generated document CSS uses Arial for `.meta` and does not use `system-ui` there. Verify they fail.
- [x] 4.2 Ground the big-note prompt, mention html-anything in the empty-topic guide excerpt, and switch `htmldoc` header chrome to Arial. Verify `go test` in `morph/handlers`, `morph/mcp`, and `morph/internal/htmldoc` pass.

## Workflow follow-up

- Archive the change after implementation and verification.
- Open a PR that says Closes #151. Do not merge.
