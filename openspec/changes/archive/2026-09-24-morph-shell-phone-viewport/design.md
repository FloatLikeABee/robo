## Context

See proposal.md for why. The main Morph SPA is the CRA app under `morph/frontend`. `public/index.html` already sends `width=device-width, initial-scale=1, viewport-fit=cover`. `index.css` and `App.css` already use `100dvh` in several shells, and chat/login/MorphNotes already mention `env(safe-area-inset-*)`. Those rules do not yet keep the document inside a phone viewport.

What the code actually does today:

- `html`, `body`, and `#root` have no `max-width: 100%` and no horizontal overflow containment. A descendant wider than the viewport scrolls the document.
- `.header-actions` is `flex-shrink: 0`. At `max-width: 768px` it is capped with `max-width: min(58vw, 280px)` and `overflow-x: auto`, so the Morph AI header scrolls sideways. The title is `flex: 0 0 auto`, so a squeezed title can still paint past the header.
- `.messages-container` scrolls vertically. A flex item’s default `min-width: auto` lets a long error string or preformatted line set the column’s minimum size.
- MorphNotes draws the home-indicator spacer only when `useMediaQuery` says phone. That hook’s first value is false, so the spacer is missing on first paint.
- Case/task and timeline drawers use `width: 100vw` at the `xs` breakpoint. `100vw` includes the scrollbar gutter and can be wider than the layout viewport.
- `/skills` padding is a fixed `24px 20px` and ignores safe areas. Login already pads with `env(safe-area-inset-*)`.
- The agent session rail at `max-width: 768px` is `position: fixed` (off-canvas) and then restated at `width`/`min-width: 80px`. It is out of flow, so it is not the document-scroll bug. Drawer usability stays in #96.

## Goals / Non-Goals

**Goals:**

- Document and primary column fit ~360–430px without horizontal scroll, including empty/loading first paint.
- Top and bottom safe areas apply from CSS on first paint.
- Error and alert banners wrap inside the column.
- One source-level regression check, then a production build and a real layout measurement.

**Non-Goals:**

- Header, chip, or drawer redesign (#96). Wrapping header actions is containment, not a new menu.
- Login visual polish (#97), composer behavior (#98), typography (#99).
- Event Logs, Content Maker, MorphUtils, native apps, PWA.
- Removing internal scroll from code blocks, markdown tables, or data grids when that box itself fits the column.

## Decisions

### 1. Contain the scrollport in `index.css`, and stop the shell from growing

Set `max-width: 100%` and `overflow-x: clip` on `html`, `body`, and `#root`, with `overflow-x: hidden` first as a fallback for engines that ignore `clip`. Give `.app-outer`, `.app`, `.chat-container`, `.chat-header`, `.messages-container`, `.message`, and `.input-container` `min-width: 0` and `max-width: 100%`. Message bubbles use `overflow-wrap: break-word` so a long token wraps only when it would overflow a bounded box. `.error-bubble`, `.MuiAlert-root`, `.skills-panel-alert`, and `.alert` use `overflow-wrap: anywhere` so an error string cannot widen the page.

`overflow-x: clip` on `html` clips the viewport scrollport. It does not create a scroll container, so it does not turn `position: fixed` drawers into absolute ones the way a transform would. Closed off-canvas rails stay off-screen. Open rails that are already within the viewport stay visible.

**Rejected:** body `overflow-x: hidden` alone. It hides overflow without making the header or column fit, and the header would still scroll internally.

**Rejected:** a `visualViewport` resize listener that writes CSS variables. It runs after first paint, which fails the empty/loading criterion, and safe-area env() already exists.

**Rejected:** a new `MobileShell` wrapper component. Login, chat, skills, and MorphNotes already have shells. Shared containment belongs in `index.css`.

### 2. Header actions wrap; they do not scroll

At `max-width: 768px`, remove `overflow-x: auto` from `.header-actions`. The row is `flex: 1 1 auto`, `min-width: 0`, `max-width: 100%`, `flex-wrap: wrap`, `overflow-x: clip`, and it may shrink. With wrapping, its min-content width is one control, not the sum of the buttons, so it cannot force the document wider than the viewport. Buttons wrap inside the row when the leftover width is short. Do not set `flex-basis: 100%`: that would always stack the actions under the title, even when they fit. The title gets `min-width: 0`, `overflow: hidden`, and `text-overflow: ellipsis`. The header stays in normal flow (grid area `head`), so a taller row pushes the transcript instead of covering it.

**Rejected:** keep `overflow-x: auto` and hide the scrollbar. That still scrolls shell chrome, which the acceptance criteria forbid.

**Rejected:** `flex-basis: 100%` so actions always sit on a second row. That is extra height on every phone width, including widths where one row still fits.

**Rejected:** a hamburger that replaces the action row. That is the #96 redesign.

### 3. Safe area in CSS, including the MorphNotes spacer

Keep `viewport-fit=cover`. Do not pad `#root` itself: chat already pads the header and composer, and a second pad would double the inset.

- Chat header and composer keep `env(safe-area-inset-top/bottom)`, and gain left/right insets via `max(existing padding, env(safe-area-inset-left/right))`. Zero on desktop.
- `/skills` uses the same `max(padding, env(safe-area-inset-*))` pattern.
- Login already does this. Only the error line gets `overflow-wrap` and `max-width: 100%`.
- MorphNotes bottom spacer uses `display: { xs: 'block', sm: 'none' }` instead of `useMediaQuery`. The mobile top bar’s minimum height is `calc(56px + env(safe-area-inset-top))` so the inset is added to the 56px row under `border-box`, not eaten by it.

**Rejected:** drop `viewport-fit=cover` and let the browser letterbox. The status bar is `black-translucent`; cover is what lets the dark shell paint edge to edge. The missing piece is padding, not the meta tag.

### 4. Phone drawers use the layout viewport, not `100vw`

At `xs`, case/task and timeline drawer paper width becomes `100%` with `maxWidth: '100%'`. Percentage width on a fixed drawer is the viewport containing block. `100vw` is not.

Do not change data-grid column `minWidth`s. Those grids already sit in a `minWidth: 0` / `overflow: hidden` main column and may scroll inside the grid.

### 5. Lock the agent rail out of the phone grid

At `max-width: 768px`, repeat `position: fixed` and `min-width: 0` on `.app.app--agent .chat-sidebar` so the later `min-width: 80px` rule cannot put an 80px column back in flow. Leave the 80px width alone (#96).

## Risks / Trade-offs

- [Clip hides a control we failed to reflow] → Header wraps and columns use `min-width: 0`. Verification measures `scrollWidth` and checks that header actions are still in the layout, not only that scrolling is disabled.
- [Wrapped header feels tall on a phone] → Buttons wrap only when one row does not fit. #96 can compact them later.
- [Double safe-area padding] → Do not pad `#root`. Only shells that lack an inset get one.
- [`overflow: hidden` fallback on `body` affects sticky or rubber-banding] → Prefer `clip` after `hidden`. Chat sticky behavior is not used on the document; panes scroll internally.
- [Short landscape phones plus the in-flow workspace row] → The workspace row is already in flow under 900px. This change does not make it fixed, so it cannot cover the composer. Composer fit stays #98.
- [Global `.alert` rule is broad] → `max-width: 100%` and `overflow-wrap` do not change color or copy.

## Migration Plan

Frontend-only. Ship with the Morph image/static build. Rollback is reverting the CSS and `AdminLayout` spacer. No data migration.

## Open Questions

None. Portrait phone fit, safe areas, and banner width are specified. Drawer and composer redesigns stay in later stories.

## Grill (proposer and reviewer)

Reviewed against `index.html`, `index.css`, `App.css`, `AdminLayout.js`, `AppDrawer.js`, `LoginPage.js`, `SkillsPage.js`, `SkoolAiChat.js`, and the `100vw` drawers before locking this design.

Alternatives considered: (1) scrollport containment plus wrap and CSS safe areas — chosen; (2) `overflow-x: hidden` only — rejected, clips without fitting the header; (3) JavaScript `visualViewport` metrics — rejected, misses first paint; (4) a new shell component — rejected, duplicates existing shells; (5) removing `viewport-fit=cover` — rejected, the meta tag is already right.

Assumptions that failed the code check:

- “Root containers have a fixed `min-width`.” They do not. The overflow comes from `flex-shrink: 0` header actions, flex `min-width: auto`, and `100vw`.
- “Safe area is missing everywhere.” Chat, login, and the MorphNotes top bar already use it. The hole is the skills page and the MorphNotes bottom spacer’s first paint.
- “The 80px agent rail causes the sideways scroll.” It is `position: fixed` under 768px. Keep it out of flow; do not redesign it here.
- “Header actions should always drop to their own row.” Forcing `flex-basis: 100%` adds a row even when the controls fit. Wrap inside the actions group instead.
