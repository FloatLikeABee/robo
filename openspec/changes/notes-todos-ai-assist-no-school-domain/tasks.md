## 1. Prompts

- [x] 1.1 Rewrite `textAssistGenerateTodoPrompt` and `textAssistGenerateNotePrompt`: lead with the user’s title, body-only output, no school/transportation persona or leftover examples
- [x] 1.2 Rewrite `textAssistTaskChainStepPrompt` to the same domain-neutral frame (no “school transportation / MorphData-style operations”)

## 2. Frontend

- [x] 2.1 On Notes & TODOs improve, send `seed: title` with the body (same `text-assist` improve mode)

## 3. Verify

- [x] 3.1 Same-package tests: prompts include the seed; leftover school-ops strings are absent for generate (empty and titled) and task_chain_step
- [x] 3.2 `go test` the handler package; grep `tran_text_assist.go` so the old persona strings are gone; confirm improve JSON includes title
