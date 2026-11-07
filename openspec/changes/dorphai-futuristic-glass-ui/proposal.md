# Proposal

## Why

Dorph AI chat still reads as a flat, heavy dark console (blue-on-near-black). The product owner wants a softer, mid-luminance “futuristic” atmosphere—purple and orange accents, glassy surfaces, and dispersed light—without flipping to a bright daytime theme or adding a new theme picker.

## What Changes

- Introduce a **twilight glass** visual preset for Dorph AI (SkoolAiChat + shared chat CSS tokens): mid-dark base (~not `#020408` black, not `#e8e4dc` beige light), purple primary accent, orange secondary/highlight, frosted panels, and a subtle animated or static dispersion background.
- Replace the hard-coded `data-chat-theme='dark'` default with the new preset (keep the attribute mechanism for embeds and admin token sync).
- Apply glass styling to primary chat surfaces: header, sidebar/session sheet, message bubbles, input bar, drawers/modals that reuse chat tokens—**without** changing layout or behavior.
- Add accessibility guardrails: minimum contrast on body text and primary buttons; reduced-motion and `prefers-reduced-transparency` fallbacks (opaque surfaces).
- Align DorphData admin MUI theme accents with the same purple/orange pair where it already mirrors chat tokens (`theme.js`), scoped to palette/background—not a full admin reskin.
- **Not in scope**: light/dark toggle, DorphUtils/Project/Content Maker reskins, new npm UI libraries, WebGL/Three.js backgrounds.

## Capabilities

### New Capabilities

- `dorph-ai-chat-visual-theme`: Observable visual requirements for Dorph AI chat (tokens, glass surfaces, accent colors, motion/accessibility fallbacks, embed inheritance).

### Modified Capabilities

- (none)

## Impact

- `dorph/frontend/src/App.css` — chat CSS variables and component classes under `.skool-ai-chat`.
- `dorph/frontend/src/SkoolAiChat.js` — theme attribute value / localStorage key behavior.
- `dorph/frontend/src/theme.js` — admin palette alignment with chat accents.
- Possible small updates to chat-related components (drawers, `AssistantReply`, hybrid context) only where they hard-code legacy blue/flat colors outside tokens.
- No API or backend changes. No new dependencies.
