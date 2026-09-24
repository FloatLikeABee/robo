## Why

Morph on a phone still scrolls sideways and can sit under the status bar or home indicator. Viewport meta and some `100dvh` / safe-area rules already exist, but the shell chrome and primary column can still be wider than a ~390px viewport, and the MorphNotes home-indicator spacer appears only after a media-query hook runs. This is the foundation slice of epic #94 (issue #95).

## What Changes

- Keep the main Morph SPA inside the layout viewport at phone widths (~360–430px CSS), including first paint, with no document-level horizontal scroll of shell chrome or the primary column.
- Respect `env(safe-area-inset-*)` on top and bottom for chat, login, skills, and MorphNotes without waiting on JavaScript.
- Let the Morph AI header actions wrap inside the header instead of scrolling sideways. Do not redesign chips, drawers, or the composer.
- Keep error and offline-style banners (`role="alert"`, chat error bubbles, MUI alerts) inside the viewport width.
- Replace phone-width `100vw` drawers that can widen the document with a percentage of the containing viewport.

## Capabilities

### New Capabilities

- `morph-shell-phone-viewport`: Phone-width fit, safe areas, and no shell/primary-column horizontal scroll for the main Morph SPA.

### Modified Capabilities

- None. `openspec/specs/` has no existing shell-viewport requirement.

## Impact

- `morph/frontend/public/index.html` (viewport meta already present; keep `width=device-width` and `viewport-fit=cover`)
- `morph/frontend/src/index.css` (document containment)
- `morph/frontend/src/App.css` (chat shell, header, skills page, error bubbles)
- `morph/frontend/src/AdminLayout.js` (safe-area spacer without a first-paint miss)
- Phone-width `100vw` drawers in MorphNotes (`CaseTasks.js`, `StoryBoard.js`)
- Frontend unit test plus production build. No Go API change. Out of scope: header/drawer redesign (#96), login polish (#97), composer (#98), typography (#99), other apps.
