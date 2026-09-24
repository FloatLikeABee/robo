## Context

See proposal.md for why. The main Morph SPA is the CRA app under `morph/frontend`. Phone shell fit and header/drawers are already on `main` (#95, #96). Login keyboard (#97, PR #102) and chat composer (#98, PR #108) are still open. This change only adjusts type, gaps, empty copy, and error dismiss.

What the code does today:

- `index.css` sets the font family and `text-size-adjust: 100%`, and does not set a body size. `App.js` and `AdminLayout.js` wrap their trees in MUI `CssBaseline`, whose default `theme.typography.fontSize` is 14, so chat and MorphNotes paint body text at 14px after hydration.
- `LoginPage.js` is inline styles, not a stylesheet. Labels and the line under the title are 12px. The error line is 14px, wraps, and has no dismiss control. Sign in is `padding: 10px 16px` at 14px, about 34px tall. Inputs are already 16px. There is no `LoginPage.css` on `main`. PR #102 adds that file and keeps the 12px and 14px sizes.
- In `App.css`, `.chat-markdown` is `0.95rem`. The first `@media (max-width: 768px)` sets `.message-bubble` to 15px and welcome paragraphs to 14px. `.error-bubble` wraps and is not a button. The welcome JSX is the mark plus the MORPHAI title, with no prompt list. `.skills-panel-empty` is `0.85rem`. `.skills-picker-empty` is 13px.
- The phone header already sets `text-overflow: ellipsis` on `.chat-header-title-stack .app-title`. The stack is `align-items: flex-start` and the title is `flex: 0 0 auto`, so the title sizes to its text and the ellipsis never gets a narrower box.
- MorphNotes empty copy is MUI `body2` (14px). The notes create control is a `size="small"` button. `MuiButton` root `minHeight` is 40, and small is shorter. Big Notes alerts already call `onClose`. `CaseTasks.js` and `StoryBoard.js` return a full-page `Alert` with no close control when the first load fails. Chat errors render as `<div className="message-bubble error-bubble">`.
- The last `@media (max-width: 768px)` in `App.css` is the header/drawer block. `phoneChrome.test.js` reads from `lastIndexOf` that query, so a later query would drop those assertions. PR #108 appends just above that block's closing brace.

## Goals / Non-Goals

**Goals:**

- 16px body copy and 1.5 line-height at the existing 768px phone breakpoint, checked at about 390px.
- A gap of at least 8px between stacked controls, and a title that ellipsizes inside the header.
- Empty sentences wrap. Create controls that already exist grow to 44px. No new empty-state feature.
- Error text on sign-in, chat, and notes alerts wraps, stays at body size, and has a 44px dismiss control.

**Non-Goals:**

- Login keyboard behavior, `LoginPage.css`, or viewport `interactive-widget` (#97, #98).
- Composer, transcript scroll, or keyboard inset (#98).
- Restacking the notes list and detail into one column.
- Desktop density, a light theme, other apps, PWA.

## Decisions

### 1. One phone breakpoint, three places type is set

Use `@media (max-width: 768px)`, the same query as the shell and header. Verify at about 390px.

- `index.css` sets `body` to `16px` and `line-height: 1.5` inside that query so the first paint is already readable.
- `getAdminTheme` repeats that on `MuiCssBaseline` body, because CssBaseline would otherwise put 14px back. `MuiTypography` `body2` becomes `1rem` / `1.5` inside the same query so notes empty sentences match. Captions stay smaller.
- Chat bubble, welcome, and skills-empty rules are inserted in the existing last `@media (max-width: 768px)` block, above `.app.app--sidebar-open`. `phoneChrome.test.js` slices from `lastIndexOf` that header, so a newer media query would hide those assertions. Leaving the last four lines of the block untouched keeps PR #108's patch context (it inserts keyboard rules just above the closing brace).
- Login sizes stay inline in `LoginPage.js`. Do not add `LoginPage.css`.

**Rejected:** a single new `phonePolish.css`. Chat styles in `App.css` load with the chat bundle and would override an earlier file. Login is not in that bundle. Splitting the overrides by the file that already owns the rule is smaller.

**Rejected:** setting `theme.typography.fontSize` to 16 for every width. That changes desktop tables and headers. The story is the phone.

**Rejected:** a `max-width: 430px` query only. Widths from 431px to 768px would keep the 12px labels.

### 2. Fix the title ellipsis by giving it a width, and keep existing gaps

In that same phone block, set `.chat-header-title-stack` to `align-items: stretch` and `overflow: hidden`, and set `.app-title` to `width: 100%`, `min-width: 0`, `overflow: hidden`, `text-overflow: ellipsis`, `white-space: nowrap`. The header actions stay `flex-shrink: 0`.

Stacked gaps that are already at least 8px (login form 16px, header 8px, message `margin-bottom` 20px, notes stacks) stay. Do not invent a new spacing scale. The collision that is actually in the code is the title painting at its intrinsic width.

**Rejected:** `flex-basis: 100%` on the title stack. That always drops the actions onto a second row, which #96 avoided.

**Rejected:** stacking the notes list over the detail at phone width. That is a layout change, not a gap.

### 3. Empty states: wrap copy, enlarge the create control that already exists

Welcome keeps the mark and title. Add `max-width: 100%` and `overflow-wrap: anywhere` on `.welcome-message` and the title so a narrow column cannot grow the page. Do not add suggestion chips.

`.skills-panel-empty` and `.skills-picker-empty` become `1rem` / `1.5` with `overflow-wrap: anywhere` and `max-width: 100%`.

`MuiButton` root and `sizeSmall` get `minHeight: 44` inside the phone query. The Big Notes create control is the default size, which was 40px; the notes-and-todos create control is `size="small"`. Both need the 44px floor. `body2` at `1rem` makes "No notes yet…" and "Nothing here yet" readable, and `overflow-wrap` on `.MuiTypography-body2` lets the sentence wrap in the list column.

**Rejected:** a shared `EmptyState` component. The three empties already render. A component would be a new surface for one story.

**Rejected:** turning the disabled "Nothing here yet" row into a button. The create control is the New button above it.

### 4. Dismiss is a 44px control on the banner that lacks one

- Sign-in error: `role="alert"`, 16px, line-height 1.5, `overflow-wrap: anywhere`, and a `type="button"` control with `aria-label="Dismiss error"`, 44×44, that calls `setError('')`. `type="button"` so it does not submit the form.
- Chat: the error bubble becomes a row. A `type="button"` with `aria-label="Dismiss error"`, class `error-bubble-dismiss`, removes that message by its index. The same phone block sets the bubble to 16px / 1.5 and the button to 44×44. The edit in `SkoolAiChat.js` is the bubble markup only, not the keyboard effect PR #108 adds.
- Notes: `MuiAlert` message is 16px / 1.5 and wraps; the action `MuiIconButton` is 44×44 inside the phone query. `MuiSnackbar` is inset by 8px and `env(safe-area-inset-bottom)` so a toast is not clipped. `CaseTasks.js` and `StoryBoard.js` gain `onClose` on the full-page load alert that currently has none. Big Notes already passes `onClose`.

**Rejected:** Clear chat as the only way to remove an error. That deletes the transcript.

**Rejected:** auto-hide for errors. A failed sign-in or a failed notes load must stay until the person dismisses it.

**Rejected:** editing `public/index.html` or `.input-container` keyboard rules. Those belong to #98.

## Risks / Trade-offs

- [PR #102 replaces `LoginPage.js` and drops the 16px sizes and the dismiss button] → This branch does not add `LoginPage.css`, so that new file does not conflict. The login hunk is the style numbers plus the dismiss button. Rebase should keep both the keyboard spacer and these sizes.
- [PR #108's keyboard rules and these type rules share the last media block] → Type rules sit above `.app.app--sidebar-open`. The keyboard patch only touches the closing lines of that block, so the two hunks do not overlap. Equal-specificity conflicts are avoided by not restating selectors #108 owns (`.input-container`, `.welcome-message` overflow).
- [CssBaseline beats `index.css`] → The theme override sets the same body size.
- [16px `body2` makes phone lists taller] → Desktop is unchanged. Taller list rows are the readable type this story asks for.
- [Dismissing a load-failure alert reveals an empty page] → That is the dismiss. The page can load again from its existing refresh control.
- [Ellipsis clips a short title] → `text-overflow` only clips when the text overflows the 100% width.

## Migration Plan

Frontend static build only. Rollback is reverting the CSS, theme overrides, login dismiss, chat error button, and the two alert `onClose` handlers. No API or data change.

## Open Questions

None. Phone type size, the title ellipsis, empty-state wrap, and which error surfaces get a dismiss control are specified.

## Grill

Reviewed against `index.css`, `theme.js`, `LoginPage.js`, `App.css` (both 768px blocks and the title stack), `SkoolAiChat.js` welcome and error markup, `NotesTodosContent.jsx`, `BigNotes.js`, `CaseTasks.js`, `StoryBoard.js`, and the open diffs for PR #102 and PR #108.

Alternatives considered: (1) phone rules in the stylesheets that already own each surface, plus inline login type — chosen; (2) one new mobile component tree — rejected, it duplicates the SPA; (3) a global 16px theme including desktop — rejected, desktop density is out of scope; (4) a new `LoginPage.css` — rejected, PR #102 creates that file and this story must not rewrite the login keyboard; (5) container queries — rejected, the SPA already keys phone layout off `max-width: 768px`; (6) a brand-new last `@media` block — rejected after checking `phoneChrome.test.js`, which treats the last `max-width: 768px` block as the header contract.

Assumptions that failed the code check:

- Body text is already 16px because `index.css` leaves the size unset. CssBaseline sets 14px on chat and MorphNotes.
- The #96 ellipsis rules already truncate the title. The title is `flex: 0 0 auto` in a `flex-start` stack, so it never shrinks below its text.
- Error banners are already dismissible. Sign-in, the chat error bubble, and the CaseTasks and StoryBoard full-page alerts have no close control.
- The empty chat has a tappable suggestion list. The welcome render is only the mark and the title.
- The notes two-column pane is a spacing collision. It is a structural split. This story wraps the empty sentence and enlarges the existing New button.
