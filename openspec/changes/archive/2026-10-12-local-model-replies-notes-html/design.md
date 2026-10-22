## Context

See proposal.md for why. The current reply path is one shared sentence, `dorphai.VisualFirstInstructions` in `pkg/dorphai/context.go`: "Visual-first: short caption, then a mermaid diagram…". It is concatenated into Dorph chat (`managementToolInstructions`), tool follow-ups, note and TODO text assist, case-task drafts, research compose, the Event Logs assistant, and the Content Maker assistant. There is no response post-processor that adds diagrams. `ExtractJSONObject` only treats a reply as a tool call when it parses as a management call; the invented version JSON in Ge's screenshot is model output, shown as-is.

A second push is always-on. `appendAlwaysOnAgentSkills` injects Research, Design, and Graphs on every Dorph chat. Graphs' body is that same visual-first sentence. Research says "Diagram first". Seeded skill bodies in Badger win over the code constant, so editing only the constant would not change an existing local database.

The management prompt also tells the model to answer records as `**Label:**` fields and to emit a lone JSON tool call. A small model applies that pattern to "what model are you using".

The workspace regression is `7f6547e` (after Agent notes #139). It removed `workspaceOpen`, the header show/hide control, and `.app--workspace-collapsed`, and it made `initialWorkspaceOpen()` always return true. Tab order flipped from Notes & TODOs then Context & Knowledge to the reverse. `openspec/specs/dorph-phone-chat/spec.md` still requires the phone pane to start closed. The Agent notes modal is a separate overlay and stays.

Generated HTML shares `dorph/internal/htmldoc` (Georgia, one sky-blue accent, one purple). Big notes still have a light-theme branch. Case-task preview JS and Project `build_project_html` copy the same serif rules. The form HTML prompt asks for dark grey plus a single orange.

## Goals / Non-Goals

**Goals:**

- One short reply-shape instruction at the shared constant, plus a one-line model identity fact taken from the resolved provider and model.
- Stop Graphs from being injected on every chat, including when an old visual-first body is stored.
- Restore the pane open/close behaviour and tab order from the parent of `7f6547e`.
- One dark sans multi-accent document theme, used by every Dorph HTML renderer and by prompts that ask a model to write HTML.

**Non-Goals:**

- Removing the management, Event Logs, or Content Maker tool catalogs.
- Changing Agent notes APIs, the modal, or `/agent-notes` redirect.
- Restoring the Files workspace.
- Render, image, or env changes.
- A light/dark switch.

## Decisions

### 1. Replace the shared sentence; do not add a second prompt layer

Set `ReplyShapeInstructions` to one short rule and assign `VisualFirstInstructions` to that same text, so every current concatenation updates without a second prompt. The rule: answer what was asked, stay short, use a diagram, chart, table, or code only when the question is structure, quantities, or code and that form is clearer. No caption, mermaid, or pixel grid for an ordinary question. Do not invent facts. The Graphs skill gets its own sentence and does not reuse this alias.

Alternative: detect Ollama and send a different prompt. Rejected. The same sentence harms every model, and a second prompt would drift.

Alternative: strip diagrams in a post-processor. Rejected. That would delete a diagram the operator asked for, and it would not stop invented version text.

### 2. Put the real provider and model in the prompt

Add `ModelIdentityInstructions(provider, model string)` in `pkg/dorphai`. Dorph chat (`chatWithManagementTools` and `GenerateChatResponseWithAgent`) prefixes it. The values come from `Client` resolution even when the API key is missing, so a configured Ollama model is named and an empty resolution says unknown. The instruction forbids invented versions and dates. Note assist and the satellite assistants get the reply-shape text only.

Alternative: answer identity questions in Go before calling the model. Rejected. It misses paraphrases ("which LLM is this") and does not fix the rest of the reply shape.

### 3. Graphs is opt-in; Research and Design drop "diagram first"

`appendAlwaysOnAgentSkills` keeps Research and Design and drops Graphs. Always-on text comes from the code defaults. If a stored body still contains "diagram first", "visual-first", or "short caption", the code default replaces it so Ge's existing Badger rows do not keep the old order. A selected Graphs skill still adds its own "one diagram when it helps" body.

Alternative: delete Research and Design from the always-on list. Rejected. They are the existing default skills; only the diagram order is the bug.

The management preamble keeps the tool catalog and the record-summary rules. It adds one line: those `**Label:**` fields and the JSON tool object are only for a real data call, not for an ordinary question.

Note text assist drops the extra "you may include a mermaid" lines and uses the shared reply-shape rule.

### 4. Restore the pane from `7f6547e^`

`initialWorkspaceOpen({ stored, phone })` again honors a saved `0`/`1` and otherwise returns `!phone`. `SkoolAiChat` stores the choice and toggles `.app--workspace-collapsed`. CSS hides the pane and drops the extra grid row and the phone `28dvh` band while collapsed. Tabs are English "Notes & TODOs" then "Context & Knowledge". Composer include icons stay; they are not the pane. Do not restore a Files tab.

### 5. One document theme, dark even if the stored name says light

`htmldoc` body font becomes `Arial, "Helvetica Neue", Helvetica, sans-serif`. Root variables carry at least three accents (cyan, violet, mint) and headings, links, callouts, and chart bars use different ones. `color-scheme: dark`. Big-note HTML always uses the dark stylesheet. Case-task preview and Project HTML copy the same font and accents. The form HTML prompt and the Content Maker page prompt ask for that theme instead of a single orange or an unspecified theme. Chat markdown keeps the dark chat surface and sets the same sans stack.

Alternative: a shared CSS file imported by Rust. Rejected. Project HTML is a format string today; duplicating the few tokens is smaller than a new cross-language asset pipeline.

## Risks / Trade-offs

- [Small models still ignore a short instruction] → The identity fact is in the prompt, tests lock the builder, and Graphs is no longer always-on. Ge still needs a local check on Ollama.
- [Stored custom Research or Design text that mentions "diagram first" is replaced on the always-on path] → Selecting the skill still sends the stored body. The replacement only catches the known bad phrases.
- [Operators who wanted the pane always open lose that unless they save it open] → Desktop default remains open. Only a saved close, or a phone with no choice, starts closed.
- [Existing published HTML is not rewritten until regenerate] → New renders and regenerations pick up the theme. No migration job.

## Migration Plan

Code and tests only. Restart Dorph (API and the frontend dev server or a frontend rebuild) to see it. No data migration and no deploy change. Rollback is reverting the commit.

## Open Questions

None. The three outcomes are specified, and the root causes above are in the current tree.
