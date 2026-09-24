## 1. Shared visual-first contract

- [x] 1.1 Add `VisualFirstInstructions` in `pkg/morphai` and fold it into `ToolFollowUpPrompt`. Keep Morph Data single-record labeled fields. Do not call the image API from the default tool loop.
- [x] 1.2 Wire Morph AI management instructions + Research/Design skill one-liners to that contract. Pixel-art hint on the image-generator agent only.
- [x] 1.3 Tests: follow-up prompt contains the visual-first contract; image-generator path still returns images without the tool loop.

## 2. Morph AI chat renderer

- [x] 2.1 Add mermaid (dynamic import) to Morph AI `ChatMarkdown`: dark theme, mermaid fences draw as diagrams, invalid source falls back to text.
- [x] 2.2 Render ```pixel``` grids (max 64×64, `#RRGGBB` or `.`/`#`) as nearest-neighbor pixel art. No image API.
- [x] 2.3 Tests or a small self-check for pixel-grid parse limits; browser: Morph AI assistant mermaid + pixel fence on dark chrome.

## 3. Other in-stack assistant chats

- [x] 3.1 Event Logs assistant markdown renders mermaid (dark). Same `ToolFollowUpPrompt` already shared.
- [x] 3.2 Content Maker assistant markdown renders mermaid (dark).

## 4. Markdown and HTML documents

- [x] 4.1 MorphNotes markdown/HTML previews (`darkPreviewSrcDoc` + Research/related views) draw mermaid; invalid source remains visible. Dark-only.
- [x] 4.2 Project and Content Maker markdown/HTML previews draw mermaid the same way.
- [x] 4.3 Data Access analysis markdown preview draws mermaid. Do not put ECharts in chat. Skip mermaid in email HTML.

## 5. Verify

- [x] 5.1 `go test` for `pkg/morphai` prompt tests. Browser: Morph AI visual reply (mermaid), dark-only; spot-check one document preview (Research or Project).
