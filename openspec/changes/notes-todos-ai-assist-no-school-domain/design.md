## Context

See proposal.md — Why. `POST /api/tran/text-assist` builds a single user prompt in `morph/handlers/tran_text_assist.go`. `generate_todo` / `generate_note` still say “for a transportation / school operations staff member”; empty generate even suggests “route change, parent communication, vehicle check”. `task_chain_step` still says “school transportation / MorphData-style operations”. Notes & TODOs `AI assist` already sends `seed: title` for generate; improve currently sends body only (no title).

## Goals / Non-Goals

**Goals:**

- Domain-neutral prompts; user’s title is the topic and appears early in the prompt.
- Same endpoint, modes, and UI control.
- Tests that fail if leftover school-ops strings return.

**Non-Goals:**

- Rewriting Morph AI chat, generic-data analysis, or seeded demo data that still mention transportation.
- Changing which button is AI assist, or auto-saving the item.
- Calling a live model in CI.

## Decisions

### 1. Rewrite prompts in `tran_text_assist.go`; do not add a second model call

- **Choice:** Replace the leftover persona and examples in `textAssistGenerateTodoPrompt`, `textAssistGenerateNotePrompt`, and `textAssistTaskChainStepPrompt`. Keep one user message to `ChatCompletion`.
- **Why:** The wrong copy is in those strings. The title is already in `seed`; the model follows the staff-member frame instead.
- **Alternative:** A extra system message “never assume school bus” while leaving the old user prompt. Rejected — two conflicting frames.

### 2. Lead with the user’s title, then rules

- **Choice:** Prompt order: title/seed first (quoted), then “write a todo/note body about that title only”, then bans (do not invent school/bus/student/schedule unless the title said so), then “output only the body”.
- **Why:** The current template states the school persona first and buries `Context:` last, which is how “make money on AI stock” became a bus-route checklist.
- **Alternative:** Keep “Context:” at the end after deleting the persona. Weaker; still easy for the model to ignore.

### 3. Generate fills body only; do not emit a competing title line

- **Choice:** Instruct: actionable checklist or short paragraphs for the given title; do not output a different title; Morph keeps the user’s title field.
- **Why:** Today the todo prompt asks for “one clear title line, then bullets”, so the model can replace the topic in that first line.
- **Alternative:** Overwrite the title from the model. Rejected — user already typed the title.

### 4. Pass title on improve

- **Choice:** `NotesTodosContent.jsx` include `seed: title` on `mode: "improve"` as well as generate. `textAssistImprovePrompt` already appends seed when non-empty.
- **Why:** Spec requires title context so polish cannot drift.
- **Alternative:** Only backend prompt change. Insufficient when improve never receives the title.

### 5. Guard with string tests, not a live LLM

- **Choice:** Same-package tests: generated prompts MUST contain the seed; MUST NOT contain leftover phrases (`transportation / school operations`, `school transportation`, `staff member`, `route change, parent communication, vehicle check`, `MorphData-style operations` as the domain frame).
- **Why:** Deterministic; catches regressions if someone pastes old copy back.
- **Alternative:** Snapshot a live DashScope reply for “make money on AI stock”. Flaky and needs keys.

## Risks / Trade-offs

- [Model still hallucinates school copy] → Prompt bans leftover domain unless the seed contains it; cannot 100% guarantee a live model. Manual check once after apply with that title.
- [task_chain_step used outside Notes & TODOs] → Same leftover frame; rewrite it domain-neutral too so other helpers do not keep the Skool persona.
- [Empty generate] → UI already requires a title (`disabled` until title). Empty-seed prompt still must not list school examples.

## Migration Plan

- Deploy Morph API with new prompt strings. No schema or client version bump.
- Rollback: revert `tran_text_assist.go` (and the improve `seed` line).

## Open Questions

- None.
