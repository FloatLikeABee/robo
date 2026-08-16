# Morph Landing Remake — Design

Date: 2026-08-16  
Status: Approved in chat (approach + section plan); awaiting spec review before implementation.

## Goal

Remake `landing/` so it accurately describes the **current Morph system**, keeps the existing **dark neon theme**, and stays **simple and beautiful**. Scroll-only: no product entry CTAs.

## Constraints

- Keep theme tokens and atmosphere: `--bg-deep`, cyan/magenta neon, Orbitron + DM Sans (+ mono accents), grid + soft scanlines.
- Static site: `index.html`, `styles.css`, `script.js` (no new framework).
- No primary CTA to Morph AI, Utils, GitHub, or contact forms.
- Follow landing composition rules: brand-first hero, one job per section, no card grids in the hero, no pill clusters, no dense app catalogs.

## Content model (current Morph)

| Layer | What it is |
|-------|------------|
| **Morph AI** | Conversational front door — chat sessions, notes/TODOs, hybrid context, skills |
| **Morph Data** | Operational records — assets, big notes, timelines, generic data, import, related admin work |
| **Morph Utils** | Maker shell — Survey Maker, Content Maker, Data Access, Projects |

Out of primary story (do not feature as peer product cards): FormsX-as-FormsX naming, Mail/ledger/Academi/Access as separate marketing tiles, old “assistants picker” language.

## Information architecture

1. **Sticky header** — Morph logo/wordmark; nav anchors: AI · Data · Utils · Flow  
2. **Hero** — Brand-forward: MORPH + one headline + one supporting sentence; full-bleed atmosphere from existing bg treatments; no pills, no CTA buttons, no orbit chip cloud  
3. **Morph AI** — One headline + short copy  
4. **Morph Data** — One headline + short copy  
5. **Morph Utils** — One headline + short copy + simple vertical list (Survey Maker, Content Maker, Data Access, Projects) — not a card grid  
6. **How it fits** — One short paragraph linking AI → Data → Utils  
7. **Footer** — Brand + year (existing tone OK: quiet, not salesy)

## Visual / interaction

- Reuse CSS variables and background layers; simplify layout spacing for breathing room.
- Typography: Orbitron for display/brand; DM Sans for body; IBM Plex Mono sparingly for section labels if useful.
- Motion (2–3): hero title stagger, gentle section reveal on scroll, subtle grid/glow pulse. Prefer CSS; light JS for nav + optional IntersectionObserver.
- Mobile: hamburger nav retained; stacked sections; hero remains one composition.

## Copy direction (draft)

- **Headline:** Ask. Record. Make.  
- **Support:** Morph brings AI chat, operational data, and maker tools together in one system.  
- Section leads stay factual and short; no “open source hub / localhost” pitch unless it fits in one quiet footer line.

## Files to change

- `landing/index.html` — rewrite structure and copy  
- `landing/styles.css` — keep tokens; remove unused card/orbit/pill styles; refine section layout  
- `landing/script.js` — nav + optional scroll reveal  

## Non-goals

- Linking into live Morph apps  
- Documenting every submodule (Booki, Academi, BK, SharpReport) on the landing  
- Replacing the visual theme with a new palette  
- Building a CMS or React app  

## Success criteria

- A visitor understands Morph AI / Data / Utils in under a minute.  
- Page feels quieter and clearer than the current multi-card catalog.  
- Theme continuity is obvious (neon dark Morph, not a generic SaaS template).
