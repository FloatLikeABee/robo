## Context

See proposal.md for why. Spec: `specs/morph-phone-chrome/spec.md`.

Checked against current `morph/frontend` (main includes header icon paths in `lib/headerAppLinks.js`):

- Morph AI header (`SkoolAiChat.js`) lays sessions toggle, title, then `.header-actions`: optional workspace toggle, Skills (label only, no icon), AI tools, MorphNotes, MorphUtils when `morphUtilsBaseURL` returns a URL, Clear chat, Sign out. Below 768px, `.header-actions` is `overflow-x: auto` with `max-width: min(62vw, 240px)` and chip labels `display: none`. Skills becomes an empty ~40px control. Six 40px controls do not fit beside the title at 390px, so the strip scrolls.
- MorphNotes `AppDrawer` permanent drawer is `display: none` below `md` (900px). The temporary drawer paper is `min(220px, 86vw)` (~56% of 390px). Nav rows set `minHeight: 40`, overriding the theme's 48px. The close `IconButton` is `size="small"`. MUI temporary drawers already trap focus, lock scroll, and dismiss on backdrop and Escape.
- Morph AI sessions sidebar is `position: fixed` and off-canvas below 768px, but `width: min(300px, 88vw)` (300px ≈ 77% of 390). There is a backdrop button and no in-sheet close; the header toggle sits under the sheet. A later `.app.app--agent .chat-sidebar { width: 80px }` rule is more specific, so the agent shell's open sheet stays an 80px rail. Closed agent grid drops the sessions column below 768px, so closed chrome does not reserve that rail.
- Context/knowledge and AI tools already use `.hybrid-drawer` at `width: 100% !important` below 768px, with a backdrop click and a close control. The close control is padding 4px on a 22px glyph, under 44px. AI tools also sets an inline `min(96vw, 1200px)`; the `!important` width wins.
- Case/task detail (`CaseTasks.js`) and the unrouted story drawers (`StoryBoard.js`) set `width: 100vw` at `xs`. `100vw` includes the scrollbar and is the usual sideways-scroll leak. Generic data and the data-grid detail drawers already use `100%` at `xs`.
- Sign-out is the only account control in this header. Settings routes redirect away from profile. There is no profile chip to surface.
- Open PR #100 (shell viewport) edits `index.css`, `App.css` safe-area and header wrap, `AdminLayout.js`, and the same `100vw` lines. This design does not rewrite that work. Phone chrome is additive CSS after the existing 768px block, plus drawer/header components.

## Goals / Non-Goals

**Goals:**

- Meet the spec at the existing 768px phone breakpoint, checked at ~390px.
- One header, one stylesheet. Desktop chips and the ≥769px collapsed session rail stay put.

**Non-Goals:**

- Login, composer, transcript, typography, iframe apps inside sheets.
- A new profile screen. Sign out remains the account entry.
- Editing `index.css` or the MorphNotes safe-area spacer.
- Restyling in-document editor tabs (Markdown / HTML / preview).

## Decisions

### 1. Overflow menu, not a scrolling strip and not a second chip row

**Choice:** Below 768px, CSS hides the inline chip row and the Clear button, and shows a More button (≥44×44) next to Sign out. More opens a panel that lists Skills, AI tools, MorphNotes, MorphUtils (only if configured), and Clear chat, each with a visible label and min-height 44px. The same handlers as the desktop chips. A backdrop button behind the panel closes it. Desktop (`min-width: 769px`) hides More and shows the inline row, so only one set is displayed. The phone block sets `.header-actions { overflow: visible; }` so the panel is not clipped by the existing `overflow-x: auto` strip. With the chips taken out of that strip, the remaining controls fit and the header does not scroll.

**Why:** At 390px the row is about 20px padding + 44px menu + title + actions. Three 44px controls (workspace toggle when present, More, Sign out) fit. Six 44px chips do not, unless they shrink under 44px or wrap onto a second row.

**Rejected:** Horizontal scroll of `.header-actions` (current). Fails the spec and hides the icon-less Skills chip. **Rejected:** Wrap every chip onto another header row. They would be visible, but a second 44px row crowds the first screen, and it fights the shell-viewport change that only wraps when the row already fits. **Rejected:** A separate mobile header component. Same behavior, two trees to drift. **Rejected:** `window.matchMedia` to swap trees. First paint would depend on an effect; CSS show/hide is already correct before JS (the More panel stays closed, which is the right closed state).

### 2. Full-width sheets with an in-sheet close control

**Choice:** Phone MorphNotes paper width becomes `100%` (not `min(220px, 86vw)`, not `100vw`). Phone sessions sidebar width becomes `100%` with `max-width: 100%` and `min-width: 0`, including `.app.app--agent .chat-sidebar`, in a media block that follows the agent 80px rule so it wins at the same specificity. Add a close button inside the sessions sheet, hidden from 769px up. Hybrid close control grows to 44×44 below 768px. Case/task and story papers use `100%` / `maxWidth: '100%'` at `xs` instead of `100vw`.

**Why:** 220/390 is a rail. 300/390 still covers the header toggle, so backdrop-only dismiss fails once the sheet is full width. The spec allows backdrop **or** close; full width needs the close control. `100%` tracks the layout viewport; `100vw` does not.

**Rejected:** Keep 88vw so a sliver of backdrop remains. The sliver is not a reliable 44px target, and the agent rule would still pin the sheet to 80px. **Rejected:** Only enlarge hit targets inside the 220px rail. Leaves the spec's sheet requirement unmet. **Rejected:** Changing `adminRightPanelStyle` `sm: min(100vw, 520px)`. That breakpoint is ≥600px, and `xs` is already `100%`.

### 3. Focus and scroll stay in the open sheet

**Choice:** MUI temporary drawers already trap focus and lock body scroll; do not set `disableScrollLock` on the phone nav drawer. The sessions sheet is `role="dialog"` `aria-modal="true"`. On open, focus moves to its close button. Tab cycles inside the sheet. Escape closes it and returns focus to the sessions toggle. `overscroll-behavior: contain` and `overflow-x: hidden` on the sheet. `.app.app--sidebar-open` keeps `overflow-x: clip` below 768px.

**Why:** The sessions sidebar is not a MUI Modal, so it does not trap focus today. Hybrid overlays already cover the viewport and stop at the overlay node.

**Rejected:** A focus-trap dependency. The cycle is a few lines and a unit test. **Rejected:** `100vw` plus `overflow-x: hidden` on `body` in `index.css`. That file is the shell-viewport change's; `100%` removes the leak at the source.

### 4. Closed chrome stays under 40%

**Choice:** Do not show the MorphNotes permanent drawer below `md`. Do not put the 80px agent rail back into the phone grid. The phone sessions sidebar stays `position: fixed` and translated off-canvas until opened, so its 100% width does not reserve layout space.

**Why:** The closed-rail bug would be introducing a permanent column while "fixing" the sheet. The agent 80px width applies only as the overlay's box, off-canvas, once the later rule sets 100% — still off-canvas when closed.

### 5. Touch targets are min-width/min-height 44, not a new control library

**Choice:** Phone CSS sets 44×44 on `.chat-nav-toggle`, `.chat-header .chat-icon-button` (not the composer buttons — those are story #98), `.header-more-button`, `.header-more-panel` items, `.sidebar-sheet-close`, `.hybrid-drawer-close`, and session rows (`min-height: 44px`). The embedded header's 36px `.chat-icon-button` rule is overridden with the same selector inside the phone query, after the base rule. Menu item labels use `.header-more-panel .header-app-link-label` so the existing `display: none` on chip labels does not blank the menu. MorphNotes nav `minHeight` becomes 44 and the close `IconButton` is 44×44. Desktop sizes stay.

**Rejected:** Raising every `.chat-icon-button` and every MUI `Tab`. Composer controls are out of scope. Editor tabs are in-document, not module chrome. The theme `MuiListItemButton` minHeight 48 already exists and is overridden locally; fix the override.

## Risks / Trade-offs

- [More panel clipped by `.header-actions { overflow-x: auto }`] → Phone rule sets `overflow: visible` on that strip. The panel is `position: absolute` on a wrapper that is not a scrollport.
- [Embedded header pins icons to 36px] → Phone query repeats `.skool-ai-chat--embedded .chat-icon-button` at 44px for header buttons only.
- [More hides chips until a tap] → Labels are visible in the menu, including Skills, which has no icon. Desktop is unchanged.
- [Two copies of chip actions in the DOM] → The inactive copy is `display: none`, so it is not hit-testable or in tab order. Handlers are passed in, not reimplemented.
- [Full-width sheet covers the backdrop] → In-sheet close, Escape, and (for MUI) backdrop when any remains. Sessions backdrop stays in the DOM for clicks that reach it.
- [Agent `width: 80px` beats an earlier overlay rule] → The phone override is the same selector, placed after the agent block.
- [PR #100 also edits `.header-actions` and `100vw`] → This change appends a media block and only replaces `100vw` with `100%` on those papers. It does not edit `index.css` or the safe-area spacer. Rebase if #100 merges.
- [Story drawers are unrouted] → Still switch them off `100vw` so a later route does not reintroduce the scroll leak. No other story behavior changes.
- [`display: none` More on desktop if CSS fails to load] → The inline chip row remains the default outside the media query; More is `display: none` in the base rule and shown only inside the phone query.

## Migration Plan

CSS and component change in the Morph SPA. Rollback is reverting the commit. No data migration.

## Open Questions

None. Account entry is sign-out because that is the only account control in this header.
