## Why

Visual-first chat now draws mermaid diagrams, pixel grids, and generated images inside the message bubble. Those visuals stay small and are hard to read. Users need a click-to-enlarge view without leaving the conversation.

## What Changes

- Clicking a graph, chart, diagram, pixel grid, or image in an in-stack assistant chat opens a dark in-app modal with an enlarged copy.
- Escape, backdrop click, and a close control dismiss the modal. Focus is trapped while it is open.
- Cover Morph AI chat first, then the same behavior in Event Logs and Content Maker assistant markdown.
- Invalid mermaid source (fallback text) is not treated as an image. Document HTML/markdown previews, Data Access ECharts reports, and email HTML are out of scope.

## Capabilities

### New Capabilities

- `chat-visual-enlarge`: Chat diagrams, charts, pixel art, and images open in a dark enlarge modal on click.

### Modified Capabilities

- (none — no archived baselines under `openspec/specs/`)

## Impact

- Morph AI: `MermaidBlock`, `PixelGridBlock`, generated `chat-generated-image`, markdown `<img>` in `ChatMarkdown` / `VisualMarkdown`
- Event Logs assistant `VisualMarkdown` (formx)
- Content Maker `AssistantMarkdown` (composerx)
- Dark overlay pattern similar to Morph AI `SkillsModal` (not browser-native dialogs)
- No API, database, or new chart library
