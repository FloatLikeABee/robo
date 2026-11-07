# Proposal

## Why

The first twilight preset (`dorphai-futuristic-glass-ui`) still reads as a dark overlay: page base `#14101c`, glass fill at 6% white, 16px blur. The product owner wants the same futuristic purple/orange look, but **noticeably lighter** and **more glassy**—without a daytime beige theme or a new picker.

## What Changes

- Retune the existing `twilight` token set (no new theme id): raise page and panel luminance into dusk, not night, and not near-white.
- Strengthen glass: higher frosted fill, stronger backdrop blur, brighter specular edges so blooms show through chrome.
- Lighten the dispersion mesh (lavender + peach blooms on a mid-dusk field) so the page itself feels brighter.
- Keep WCAG AA on body text and primary buttons; keep reduced-motion / reduced-transparency fallbacks, raising opaque fallback surfaces to match the new dusk floor.
- Nudge admin MUI dark background/paper to the same dusk floor so DorphNotes does not sit as a black island next to chat.
- **Not in scope**: new theme value, light/dark toggle, WebGL, other products, layout/behavior changes.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `dorph-ai-chat-visual-theme`: twilight default must be dusk-luminance with stronger frosted glass (this capability is introduced in in-flight change `dorphai-futuristic-glass-ui`).

## Impact

- `dorph/frontend/src/App.css` — twilight tokens, mesh, glass blur/fill, fallbacks.
- `dorph/frontend/src/theme.js` — admin dark background/paper to match dusk floor.
- No API, no new files unless a tiny contrast check is needed. No new dependencies.
