## 1. Event Logs without Info Sheets

- [x] 1.1 Remove the Info Sheets `NavLink` from `formx/frontend/src/components/Layout.tsx`. Header sections are Events & Info only
- [x] 1.2 In `formx/frontend/src/App.tsx`, redirect `/survey-bot` and the old `/forms*` aliases to `/events-info`. Stop rendering `SurveyBot`. Delete `SurveyBot.tsx` if nothing else imports it. Keep public `/s/:slug` and `/f/:slug`
- [x] 1.3 MorphUtils Event Logs copy in `morph-utils/frontend/src/config.ts`: drop Info Sheets. Same for Event Logs `index.html` meta description and `formx/README.md` operator line

## 2. Shared graphs contract

- [x] 2.1 Add `VISUAL_FIRST_INSTRUCTIONS` in `pkg/morphai-rs/src/context.rs` matching Go `morphai.VisualFirstInstructions`. Include it in `tool_follow_up_prompt` and `tool_follow_up_prompt_with_instruction` the same way Go does. Export it from `lib.rs`
- [x] 2.2 Append `morphai.VisualFirstInstructions` to first-turn Event Logs (`formsXAssistantInstructions`) and Content Maker (`tranMailAssistantInstructions`)
- [x] 2.3 Append `VISUAL_FIRST_INSTRUCTIONS` to Data Access `DATA_AI_INSTRUCTIONS` and Project `ENGI_INSTRUCTIONS` (keep Data AI report headings; still no chart for a greeting or a single labeled record)
- [x] 2.4 Seed Morph AI builtin skill `builtin-graphs` (name Graphs, enabled by default) whose instructions are the VisualFirst contract. Update `skills_builtin_test.go` seed counts
- [x] 2.5 MorphNotes: `textAssistBodyOnlyRules` may emit mermaid fences (still no replacement title). Append VisualFirst to `caseTaskAIDraftPrompt` and research compose/write prompts in `tran_research.go`. Tests: text-assist no longer forbids mermaid; research compose prompt contains Visual-first

## 3. MorphNotes task create: markdown + HTML

- [x] 3.1 `caseTaskAIDraft` JSON: required `title`, `markdown` (fallback `description`), `start_at`, `end_at`. Do not require `location` or `detail`. Prompt says the document is markdown (mermaid allowed). Parser + tests in `tran_case_tasks_ai_test.go`
- [x] 3.2 Create form in `CaseTasks.js`: keep Generate (prompt + upload), title, start/end. Hide description, map, JSON detail, attachments when `!editing`. Generate fills title/times/`description` from `markdown`; empty location/detail; do not require detail JSON. Confirm copy talks about replacing the document, not map/JSON
- [x] 3.3 Create save: title, times, description=markdown, empty location/detail. Markdown tab edits that document. `markdownToHTMLFragment` in `caseTaskViewDocs.js` keeps fence language (`language-mermaid`) so HTML preview draws mermaid via `darkPreviewSrcDoc`. Edit of an existing task may still show map/JSON/attachments

## 4. Verify

- [x] 4.1 `go test` for `pkg/morphai`, `morph/handlers` (skills seed, text-assist, case-task AI draft, research compose prompt). Rust morphai-rs compile if tests exist
- [x] 4.2 Browser: Event Logs header has no Info Sheets; `/survey-bot` lands on Events & Info; MorphUtils Event Logs card has no Info Sheets copy
- [x] 4.3 Browser: MorphNotes Tasks create shows generate + title + times only; generate produces Markdown/HTML (mermaid if the prompt is a process); save; edit of an old task still opens
- [x] 4.4 Browser (spot-check): one MorphUtils module AI (Event Logs or Data Access) and MorphNotes notes assist can return a mermaid diagram for a structure/quantity question. Dark-only
