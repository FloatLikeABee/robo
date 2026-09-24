## Context

See proposal.md. Spec: `chat-visual-enlarge`.

Visual-first chat already draws mermaid (SVG), pixel-grid canvases, Morph AI generated `Images[]`, and GFM markdown images inside the bubble. Those stay at bubble width. Morph AI `SkillsModal` is the existing dark overlay (Escape + backdrop). Confirm dialogs are for confirms, not pictures. Frontends do not share an npm package (same constraint as visual-first). Dark-only; Morph AI is the system chat.

## Goals / Non-Goals

**Goals:**

- One enlarge overlay per chat frontend; click mermaid / pixel / image to fill it.
- Clone the already-drawn visual (do not re-call mermaid or the image API).
- Keyboard: the visual is activatable; Escape closes; focus on close control while open.

**Non-Goals:**

- Document HTML/markdown previews (Research, Project, notes).
- Data Access ECharts reports.
- Email HTML.
- Pinch-zoom, pan, download, or a gallery of every image in the thread.
- Shared `@robo` lightbox package.
- Light/dark theme switch.

## Decisions

### 1. Clone into a product-local overlay, not CSS zoom in the bubble

**Choice:** A single overlay at the chat root. Click copies the current SVG HTML, canvas bitmap, or image `src` into the overlay at a larger max size. Bubble layout does not change.

**Why:** In-place `transform` would clip inside `overflow` bubbles and fight mermaid’s own SVG width.

**Alternative:** New mermaid `render()` at a bigger id — extra work, flicker, id collisions. **Alternative:** Browser fullscreen API — leaves the dark in-app chrome.

### 2. Per-frontend copy, Morph AI first

**Choice:** Morph AI: small overlay component + click handlers on mermaid, pixel canvas, generated `<img>`, and markdown `<img>` in `VisualMarkdown`. Event Logs and Content Maker get the same overlay next to their assistant markdown (no new shared package).

**Why:** Bundlers differ (CRA vs Vite vs Svelte). Visual-first already accepted this split.

**Alternative:** One npm package — rejected (YAGNI).

### 3. Overlay like SkillsModal, not ConfirmDialog

**Choice:** Dark full-viewport overlay, `role="dialog"` `aria-modal="true"`, close button, backdrop click, Escape, body scroll lock. Show the visual large (`max-width`/`max-height` of the viewport). Pointer cursor on enlargable visuals.

**Why:** Confirm dialogs are title + buttons. SkillsModal already matches chat chrome.

**Alternative:** MUI `Dialog` — works in Morph AI only; skip a second modal system.

### 4. Fallback text and chrome are not targets

**Choice:** Only a successfully rendered mermaid SVG, pixel canvas, generated image, or markdown image opens the modal. Avatars, send icons, and mermaid source fallback stay inert.

**Why:** Spec says fallback is not an image.

## Risks / Trade-offs

- [SVG clone is tiny] → Size the overlay visual with CSS (`width: min(100%, …)` / `height: auto`), not a second mermaid pass.
- [Canvas clone is blank] → Copy via `canvas.toDataURL()` at click time, not a live 2d context share.
- [Focus trap incomplete] → Focus the close button on open; restore the previously focused element on close. ponytail: full tab-cycle trap only if a cheap existing helper is already in that frontend.
- [Click on mermaid inner `<a>`] → If mermaid emits links, let the link win; otherwise the diagram click opens the modal.

## Migration Plan

1. Morph AI overlay + mermaid/pixel/generated/markdown images.
2. Event Logs + Content Maker assistant markdown.
3. Rollback: remove overlay components; visuals stay inline.

## Open Questions

None. Document previews stay out of this change unless a later request asks for them.
