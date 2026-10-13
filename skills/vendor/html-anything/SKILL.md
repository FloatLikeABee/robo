---
name: html-anything
description: Base skill Morph AI loads when it produces HTML or markdown that becomes HTML. Vendored from nexu-io/html-anything (Apache-2.0). Use for notes, grill drafts, generated documents, and form HTML.
---

# html-anything (Morph base HTML skill)

Vendored from [nexu-io/html-anything](https://github.com/nexu-io/html-anything), Apache-2.0. LICENSE and NOTICE live in this directory. Morph injects `condensed.md` into HTML-producing prompts so local Ollama stays inside a 512-token budget. This file is the readable skill for agents.

## When

Load this skill whenever the output is HTML, or markdown that Morph will render as HTML: note AI, grill final content, timelines, research, stories, form pages.

## Format

Keep the format the caller asked for.

- HTML page: self-contained document, `<!DOCTYPE html>` … `</html>`, no markdown fences, first character `<`.
- Note or grill body: body text only. No greeting, no title line, no raw HTML document.
- Generated markdown: markdown the server can render. Do not invent an `html` key when the caller said markdown only.

## Craft (from html-anything)

- Cover every user point. Do not invent facts or lorem.
- Clear heading hierarchy. 8px rhythm. Prose about 65 characters wide.
- Contrast at least 4.5. Prefer inline CSS or SVG over random image URLs.
- Structured data becomes a small table or chart, not a wall of prose.

## Ge overlay (wins)

- Font: Arial, Helvetica Neue, Helvetica, sans-serif.
- Futuristic dark mode only. Background `#070b14`. Text `#e8eef7`.
- Varied colours: cyan `#38bdf8`, violet `#a78bfa`, mint `#34d399`.
- Clear content. No light page.

Agents: see `docs/agents/16-morph-for-agents.md`. The compact text Morph injects is `condensed.md`.
