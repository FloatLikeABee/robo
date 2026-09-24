## Why

A Morph user on a phone (~390px portrait) still cannot use the sign-in screen comfortably. The card is vertically centered in a full-height shell, so the on-screen keyboard covers the password field and Sign in control, and that control is shorter than a 44px touch target. Failed-login text must stay inside the screen, and the page reached after sign-in must still fit the phone. This is issue #97 in epic #94, on top of the merged phone-viewport shell (#95).

## What Changes

- Lay the `/login` form out so username, password, labels, and Sign in fit a ~390px portrait viewport with no document horizontal scroll.
- Keep the focused field visible above the on-screen keyboard, and let the form scroll so Sign in stays reachable when the card is taller than the visible area.
- Keep a failed-login message readable and inside the card.
- Keep Sign in at least 44×44 CSS pixels. This screen has no primary links; do not add a password-reset or OAuth link.
- Leave the post-login Morph AI landing on the existing phone shell. Do not restyle header, drawers, or the composer.

## Capabilities

### New Capabilities

- `morph-phone-login`: Phone layout, keyboard reachability, error wrapping, and touch size for the main Morph SPA sign-in screen, plus a usable landing after a successful sign-in.

### Modified Capabilities

- None. `morph-shell-phone-viewport` already requires the sign-in route to fit the phone width and keep safe areas. This change does not loosen those requirements.

## Impact

- `morph/frontend/src/pages/LoginPage.js` and a login stylesheet next to it.
- A small login-only keyboard helper under `morph/frontend/src/auth/`.
- Frontend unit tests and a production build. No Go API or auth-model change.
- Out of scope: Invite Signup, MorphUtils account modal, password reset, OAuth, SheetX, ComposerX, chat composer (#98), header and drawers (#96), typography (#99), and shared viewport meta changes.
