## Context

See proposal.md for why. The only main-SPA auth screen is `morph/frontend/src/pages/LoginPage.js`, routed at `/login` outside `ProtectedLayout`. Success calls `releaseStuckOverlays()` and `navigate`s to `safeReturnPath` (default `/`), which renders Morph AI (`App` → `SkoolAiChat`). There is no second auth screen, no primary link, and no password-reset control in this file.

What the code does today, after the merged shell-viewport change (#95 / PR #100):

- The page is a `100dvh` grid with `placeItems: 'center'`, safe-area padding, a `maxWidth: 400` card, and inputs at `width: 100%`, `boxSizing: 'border-box'`, `fontSize: 16`.
- The error line already sets `maxWidth: '100%'` and `overflowWrap: 'anywhere'`. It is not an alert.
- Sign in uses `padding: '10px 16px'` and `fontSize: 14`. That box is about 36px tall, under 44px. The button is full width, so width is already over 44px.
- `html`, `body`, and `#root` already clip horizontal overflow. Login is not a separate stylesheet.
- `public/index.html` viewport is `width=device-width, initial-scale=1, viewport-fit=cover`. It does not set `interactive-widget`.
- Global `* { box-sizing: border-box }` means extra `padding-bottom` on a `min-height: 100dvh` shell is inside the box and does not create scroll room until content plus padding exceeds that minimum.
- iOS does not shrink the layout viewport when the keyboard opens. A card centered in `100dvh` puts Sign in in the lower half, under the keyboard, and the page is not taller than the layout viewport, so the operator cannot scroll it clear.
- Open PR #101 edits `App.css`, `SkoolAiChat.js`, and drawer files for header/chips. This story must not rewrite those.

## Goals / Non-Goals

**Goals:**

- Phone portrait sign-in fits, wraps errors, and gives Sign in a 44px hit area.
- The card starts at the top of the shell in CSS, and a focused field can scroll Sign in into the visible viewport when a keyboard overlaps the layout viewport.
- Successful sign-in still lands on the existing shell with no stuck backdrop.
- One unit-tested keyboard calculation, a source regression check, a production build, and a ~390px layout measurement.

**Non-Goals:**

- Header, chips, drawers (#96), composer (#98), typography (#99).
- Changing `index.html` viewport meta, `index.css` containment, or `App.css`.
- Password reset, show-password, OAuth, invite signup, MorphUtils account modal, API auth changes.
- A new login component or a second mobile tree.

## Decisions

### 1. A login stylesheet, top-aligned on a phone, scroll container owned by the shell

Move the login presentation into `LoginPage.css` next to the page. Keep the dark card. At `max-width: 480px`, set `align-content: start` so the card sits at the top on first paint. Above that width, keep vertical centering.

The shell is `min-height: 100dvh` (not a fixed height), `width: 100%`, `max-width: 100%`, `overflow-x: clip`. A fixed `height: 100dvh` plus inner `overflow-y: auto` is the weaker path: iOS often does not move an inner scroller when the keyboard opens. Growing the document lets the visual viewport scroll. If measurement shows `overflow-x: clip` on `html` blocks that vertical scroll, the shell itself becomes the scroll container (`max-height: 100dvh; overflow-y: auto`) and the spacer lives inside it. Inputs and the card use `min-width: 0`, `width: 100%`, `box-sizing: border-box`. Inputs stay at `font-size: 16px` and gain `min-height: 44px`. Sign in gets `min-height: 44px` and `min-width: 44px`. The error keeps `overflow-wrap: anywhere`, `max-width: 100%`, and `role="alert"`. Safe-area padding stays on the shell via `env()`, including `scroll-padding-top` so a scrolled field does not hide under the notch. Do not pad `#root` again.

Stable `id` and `name` (`username`, `current-password`) plus the existing `autocomplete` values let the platform keyboard and password manager fill the fields. `enterkeyhint` is `next` on the username and `go` on the password. No show-password control and no extra links.

**Rejected:** leave the rules inline. Inline style cannot express a media query, and a `matchMedia` hook would miss first paint the way MorphNotes `useMediaQuery` did.

**Rejected:** edit `App.css` or `index.css`. The shell containment is already there, and `App.css` is the open #96 change.

### 2. Keyboard overlap is a login-only spacer, not a viewport-meta change

Export a pure helper from `morph/frontend/src/auth/authKeyboard.js`:

- `authKeyboardOverlap({ innerHeight, visualViewport, focused })` returns `0` when the field is not focused, when `visualViewport` is missing, or when `innerHeight - height - offsetTop` is under 120px (browser chrome, not a keyboard). Otherwise it returns that overlap in CSS pixels.
- `controlHidden(rect, visualViewport, gap)` is true when the control's bottom is below the visible bottom or its top is above the visible top, with an 8px gap.

`LoginPage` listens for focus on the form, writes the overlap to a spacer element after the card (`height: var(--auth-keyboard-inset)`), and on the next frame scrolls the focused field or Sign in with `scrollIntoView({ block: 'nearest' })` only when `controlHidden` says so. Blur that leaves the form, and unmount, set the inset back to `0`. The spacer is a child after the card, so it adds document height. Padding on the border-box shell would not, because global `box-sizing: border-box` keeps that padding inside `min-height: 100dvh`.

**Rejected:** `<meta name="viewport" content="..., interactive-widget=resizes-content">`. That resizes every route when a field is focused, including the chat composer (#98), and it is not a login-file change.

**Rejected:** `place-items: safe center` only. `safe` switches to start when the item overflows its alignment container. The keyboard does not shrink the iOS layout viewport, so the centered card still sits in the middle of a full-height shell and the page does not grow.

**Rejected:** a `visualViewport` listener that rewrites shell safe-area insets on first paint. #95 already rejected that for safe areas. This listener runs only after focus, which is when a keyboard exists.

### 3. Do not restyle the post-login shell

Keep `releaseStuckOverlays()` immediately before `navigate`. The landing at `/` is the Morph AI shell already constrained by #95. Verification measures that document at ~390px. No `App.css` or header edits. A stuck backdrop would be fixed in `releaseStuckOverlays` only if measurement shows one; the current function already clears a hidden MUI modal and a locked `body` overflow.

**Rejected:** a mobile landing layout or header compaction. That is #96.

**Rejected:** a separate post-login screen. The product lands in Morph AI.

## Risks / Trade-offs

- [Top-aligned card looks sparse on a tall phone] → That is the point: the lower half is where the keyboard goes. Desktop widths stay centered.
- [URL-bar resize looks like a keyboard] → Ignore overlap under 120px, and only while a sign-in field is focused. Portrait keyboards are well above 120px. Landscape phones are outside this story.
- [Chrome already shrinks `innerHeight` with the keyboard] → Overlap is then ~0 and the spacer stays 0. The shorter viewport scrolls the document if the card does not fit. No double inset.
- [Document cannot grow because `html` clips overflow] → Fall back to the shell as the scroll container and confirm Sign in moves with `scrollIntoView`. Do not change the shared `index.css` clip.
- [Spacer left on after blur] → Clear on blur outside the form and on unmount.
- [`scrollIntoView` before the spacer is in layout] → Set the CSS variable, then scroll in `requestAnimationFrame`.
- [Clipping a control that failed to shrink] → The card is `width: 100%` with `min-width: 0`. Verification checks `scrollWidth`, not only that overflow is clipped.
- [16px inputs feel large] → Required to stop iOS zoom, which itself causes horizontal pan.

## Migration Plan

Frontend static build only. Rollback is reverting `LoginPage.js`, `LoginPage.css`, and `authKeyboard.js`. No data migration and no API change.

## Open Questions

None. Portrait phone fit, keyboard reachability, error wrapping, the 44px Sign in target, and the untouched landing shell are specified.

## Grill (proposer and reviewer)

Reviewed against `LoginPage.js`, `appRouter.js`, `ProtectedLayout.js`, `releaseStuckOverlays.js`, `index.css`, `public/index.html`, `auth/morphSession.js`, and the archived `morph-shell-phone-viewport` design before locking this.

Alternatives considered: (1) phone top-align plus a login-only keyboard spacer — chosen; (2) `interactive-widget=resizes-content` on the shared viewport meta — rejected, it resizes chat and is not local to login; (3) `place-items: safe center` with no spacer — rejected, iOS does not shrink the layout viewport for the keyboard, so the centered card stays under it and cannot scroll; (4) a separate mobile login component — rejected, one screen and a stylesheet are enough; (5) restyling the post-login header — rejected, that is open #96 and the landing shell already fits under #95.

Assumptions that failed the code check:

- “The login card is wider than 390px.” It is `width: 100%` with `maxWidth: 400` and border-box inputs. Horizontal overflow is not the main bug. The bug is vertical placement, the short button, and the error not being an alert. Still measure `scrollWidth` so a long token cannot widen the page.
- “Safe area is missing on login.” The page already pads with `env(safe-area-inset-*)`. Keep it. Do not pad `#root`.
- “Extra padding-bottom on the shell creates keyboard scroll room.” Global `box-sizing: border-box` puts that padding inside `100dvh`. A spacer child is what grows the scroll height.
- “The post-login page needs a new phone layout.” `/` is Morph AI. `releaseStuckOverlays()` already runs on success and on every pathname change. Do not edit `App.css` while #101 is open.
- “There are primary links to enlarge.” There are none. Do not add a forgot-password link in this story.
