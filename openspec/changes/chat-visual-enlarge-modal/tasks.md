## 1. Morph AI overlay

- [x] 1.1 Add a dark enlarge overlay (SkillsModal-style: dialog, Escape, backdrop, close control, focus close on open). Clone SVG HTML, canvas `toDataURL()`, or image `src` — do not re-run mermaid or the image API.
- [x] 1.2 Wire mermaid, pixel canvas, generated chat images, and markdown images in Morph AI chat so click/keyboard opens that overlay. Fallback mermaid source stays inert.

## 2. Other in-stack chats

- [x] 2.1 Event Logs assistant mermaid (and markdown images if present) opens the same kind of dark enlarge overlay.
- [x] 2.2 Content Maker assistant mermaid (and markdown images if present) opens the same kind of dark enlarge overlay.

## 3. Verify

- [x] 3.1 Browser: Morph AI — click mermaid, pixel grid, and a generated or markdown image; modal enlarges; Escape / backdrop / close dismiss. Dark-only. Spot-check Event Logs or Content Maker mermaid click.
