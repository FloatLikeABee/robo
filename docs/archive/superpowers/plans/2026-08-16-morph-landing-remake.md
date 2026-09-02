# Morph Landing Remake — Implementation Plan

Date: 2026-08-16  
Spec: `docs/superpowers/specs/2026-08-16-morph-landing-remake-design.md`

## Files

| File | Role |
|------|------|
| `landing/index.html` | Structure + copy for AI / Data / Utils / Flow |
| `landing/styles.css` | Keep theme tokens; sparse layout; drop card/orbit/pill styles |
| `landing/script.js` | Mobile nav + scroll reveal + reduced-motion respect |

## Tasks

1. Rewrite `index.html` per approved IA (hero + 4 sections + footer); no CTAs.
2. Rewrite `styles.css`: keep `:root` + bg layers; new section/list styles; remove unused grids.
3. Update `script.js`: nav toggle + IntersectionObserver for `.reveal`.
4. Open/smoke in browser mentally: mobile nav, no leftover old product names.

## Done when

Visitor sees Morph AI, Morph Data, Morph Utils clearly; theme unchanged; no product links.
