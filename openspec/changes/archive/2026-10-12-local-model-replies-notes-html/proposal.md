## Why

Local Ollama replies on Morph AI are diagram-heavy and invent model facts, because every chat and note assist is told to lead with a caption and a mermaid diagram. The Notes & TODOs workspace no longer opens the way it did before the October workspace change, and generated HTML documents still use a serif face and a single accent.

## What Changes

- Chat, note AI assist, and the other prompts that share that instruction answer the question in short plain language. A diagram, chart, table, or code block appears only when it makes that answer clearer. The reply reports the configured provider and model, or says they are unknown, and does not invent a version or date.
- The Graphs skill stops riding along on every reply. Research and Design stay available without a "diagram first" order. Tool catalogs stay, so data actions still work.
- The Morph AI workspace pane returns to its previous opening behaviour: desktop open unless the operator saved it closed, phone closed until asked, a show/hide control, and Notes & TODOs before Context & Knowledge. The Agent notes modal stays.
- Generated Markdown and HTML documents use an Arial-like sans stack, a dark-only futuristic theme, and several accent colors. Light document themes are not emitted.

## Capabilities

### New Capabilities

- `morph-chat-reply-shape`: Default assistant replies match the question, stay concise, and do not invent model identity.
- `morph-workspace-pane`: The Notes & TODOs / Context & Knowledge pane opens and closes as it did before the always-open change.
- `morph-generated-document-theme`: AI-generated HTML and Markdown documents render in the dark multi-accent sans theme.

### Modified Capabilities

- None. `morph-phone-chat` already requires the phone workspace to stay closed until asked; this change makes the code match that spec again.

## Impact

- `pkg/morphai` reply instructions and tool follow-ups.
- Morph chat prompt assembly (`morph/handlers/management_chat.go`, `morph/handlers/skills.go`, `morph/ai/ai.go`) and note text assist.
- Satellite assistants that append the same instruction (Event Logs, Content Maker).
- Morph AI workspace pane in `morph/frontend` (`AgentWorkspace.js`, `SkoolAiChat.js`, `App.css`).
- Shared HTML document CSS (`morph/internal/htmldoc`), big-note and case-task renderers, Project document HTML, and HTML generation prompts.
- No Render, deploy, or Agent notes modal contract changes.
