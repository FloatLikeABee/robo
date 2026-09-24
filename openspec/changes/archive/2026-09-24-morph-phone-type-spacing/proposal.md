## Why

After the phone shell, header, and drawer fixes, Morph on a ~390px screen still uses 12–15px copy, tight stacks, and error banners that are hard to read or dismiss. Issue #99 is the typography, spacing, and empty/error polish for the main Morph SPA (login, chat, notes).

## What Changes

- At phone width, body copy on login, Morph AI chat, and MorphNotes uses a 16px size and 1.5 line-height so normal text is readable without pinch-zoom.
- Stacked sections and lists keep a consistent gap so controls do not sit on top of each other. The chat title ellipsis can shrink inside the header.
- Empty states (chat welcome, no skills, no notes) wrap inside the column. Any button in those states is at least 44×44 CSS pixels.
- Error banners and toasts on those screens wrap, stay at body size, and can be dismissed. The dismiss control is at least 44×44 CSS pixels.

## Capabilities

### New Capabilities

- `morph-phone-type-spacing`: Phone-width type, stack spacing, empty-state fit, and dismissible error banners for the main Morph SPA.

### Modified Capabilities

- None. Shell fit, header chrome, and (when merged) login keyboard and chat composer stay in their own specs. This change does not relax those requirements.

## Impact

- `morph/frontend/src/index.css` (phone body size before hydration)
- `morph/frontend/src/App.css` (chat, welcome, skills empty, error bubble, title ellipsis)
- `morph/frontend/src/theme.js` (MorphNotes and chat MUI type, alerts, snackbars, small buttons)
- `morph/frontend/src/pages/LoginPage.js` (label, subtitle, and error size; dismiss control; Sign in hit area)
- `morph/frontend/src/SkoolAiChat.js` (dismiss on a chat error bubble only)
- `morph/frontend/src/pages/admin/CaseTasks.js` and `StoryBoard.js` (close on a full-page error alert)
- Frontend unit test and production build. No Go change. Out of scope: new features, dark-mode redesign, a11y audit, other apps, PWA, and the shell/header/login-keyboard/composer structure from #95–#98.
