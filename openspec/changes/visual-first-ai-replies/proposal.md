## Why

Platform AI still answers like a memo: long markdown prose in Morph AI, Event Logs, Content Maker, MorphNotes documents, Project markdown, and Data Access analysis. The product motto is visual-first — charts, graphs, mermaid for structure/code, and occasional pixel art — including stored markdown and HTML, not only the live chat bubble.

## What Changes

- Make visual-first the default reply contract for every in-stack AI tool loop: short caption, then a diagram/chart when the content is structure, process, comparison, or quantities.
- Render mermaid (flow/sequence/class for programming-like content; pie/xychart for numbers) in chat markdown and in markdown/HTML document previews. Dark-only mermaid theme.
- Sometimes illustrate a character, item, or scene with pixel art: reuse Morph AI’s existing image-generator path when that agent is selected; in ordinary chats, allow a compact fenced pixel grid (no extra image-API call on every turn).
- Keep Morph Data single-record answers as compact labeled fields (a fake chart of one address is worse than a list). Do not revive platform-chat drawers. No theme switch.

## Capabilities

### New Capabilities

- `visual-first-ai-replies`: Shared visual-first prompt contract for in-stack AI chats, plus mermaid/chart (and occasional pixel art) rendering in those chats.
- `visual-first-documents`: Markdown and HTML documents produced or previewed in MorphNotes, Project, Content Maker, and Data Access analysis show the same diagrams/charts instead of leaving mermaid as a dead code fence.

### Modified Capabilities

- (none — no archived baselines under `openspec/specs/`)

## Impact

- `pkg/morphai` (shared follow-up / visual-first instruction text used by Morph AI, Event Logs, Content Maker)
- Morph AI `ChatMarkdown.js` + management/tool prompts; existing `image-generator` agent for pixel-style pictures
- MorphNotes HTML/markdown previews (`darkPreviewSrcDoc`, Research and related document views)
- Event Logs assistant markdown (`formx`), Content Maker markdown/HTML (`composerx`), Project markdown (`morph-engi`), Data Access analysis markdown (`SharpReport`)
- New mermaid dependency only on frontends that render those markdown/HTML surfaces
- Tests for prompt contract and fence rendering; browser check on Morph AI chat
