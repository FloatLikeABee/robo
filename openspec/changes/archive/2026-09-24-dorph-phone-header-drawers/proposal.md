## Why

On a phone (~390px) the Dorph AI header packs menu, app chips, and account actions into a sideways-scrolling strip of sub-44px targets, and the DorphNotes and session drawers stay narrow rails. A person cannot reach every control or dismiss a sheet without fighting the page.

## What Changes

- At phone width, primary header actions stay on screen: sessions menu, sign-out, and a More control. App chips (Skills, AI tools, DorphNotes, DorphUtils when configured) and Clear chat open from that menu with visible labels and ≥44×44 CSS px targets. The header does not scroll sideways.
- Phone drawers and side panels (DorphNotes nav, Dorph AI sessions, context/knowledge and AI tools sheets, case/task and story panels) open as full-width overlay sheets with a ≥44×44 close control and backdrop dismiss where the sheet does not cover the backdrop. Open sheets do not let the background scroll sideways. Closed chrome does not reserve more than 40% of the viewport.
- Desktop header and the collapsed Dorph AI session rail stay as they are.

## Capabilities

### New Capabilities

- `dorph-phone-chrome`: Phone-width header overflow, touch targets, and overlay drawers for the main Dorph SPA (Dorph AI and DorphNotes).

### Modified Capabilities

- None. No existing `openspec/specs/` requirement describes this chrome.

## Impact

- `dorph/frontend` only: Dorph AI header (`SkoolAiChat.js`), header CSS (`App.css`), DorphNotes `AppDrawer.js`, and the case/task and story drawer widths that still use `100vw`.
- No API, auth, or Go changes. No login, composer, transcript, typography, or iframe-satellite redesign. Shared viewport/safe-area rules from the open shell-viewport work stay intact; this change adds phone chrome behavior beside them.
