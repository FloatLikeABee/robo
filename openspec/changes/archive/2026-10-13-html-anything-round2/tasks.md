# Tasks

## 1. Robust parse and never-hard-fail generate

- [x] 1.1 Add failing fixtures and tests: missing-comma object (`invalid character '"' after object key:value pair`), a fenced non-JSON reply, and a prose reply with no `{...}`. `parseModelJSON` must accept the missing-comma fixture. A generate helper must retry once then store markdown and return no parse error. Verify `go test` in `morph/handlers` fails before the repair/fallback exists.
- [x] 1.2 Implement missing-comma repair (try each balanced `{...}`), and make big-note generate retry once then fall back to fenced-stripped markdown. Verify the new tests pass and existing handler parse tests still pass.

## 2. Grill fragment, grounding, and card

- [x] 2.1 Add failing tests using the tarball `grill-draft.html`: sanitizer drops doctype/html/head/body/style; `$30,000` against answers that say `30k RMB` falls back to facts; write prompt keeps units and forbids invented dates/places; `AppHTML` cards are `#111827` with an accent border and not a violet fill. Verify they fail.
- [x] 2.2 Implement fragment sanitize, currency/invention grounding, write-prompt rules, and card chrome. Verify `go test` in `morph/grill` pass.

## 3. Big-note depth and budget

- [x] 3.1 Add failing tests that the generate prompt requires at least three `##` sections, that a one-paragraph reply is expanded to ≥3 headings using the idea, and that `HTMLSkillInstructions` is still ≤512 tokens. Verify they fail.
- [x] 3.2 Add the depth rule to the prompt and expand thin markdown from the idea. Verify `go test` in `morph/handlers` and `pkg/morphai` pass.

## Workflow follow-up

- Archive the change after implementation and verification.
- Open a PR that says Closes #153. Do not merge.
