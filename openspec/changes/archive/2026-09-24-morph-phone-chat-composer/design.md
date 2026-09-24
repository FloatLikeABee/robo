## Context

See proposal.md for why. Spec: `specs/morph-phone-chat/spec.md`.

Checked against current `morph/frontend` after the merged shell (#100) and header/drawer (#101) work:

- `.messages-container` sets `overflow-y: auto` and not `overflow-x`. CSS then computes the other axis to `auto`, so a wide child scrolls the transcript sideways.
- Message bubbles already use `overflow-wrap: break-word`. Error bubbles use `overflow-wrap: anywhere`. That wraps prose once the column has a bounded width. It does not stop a `pre` or table from becoming the scroll width of the pane.
- `.chat-markdown pre` scrolls itself (`overflow-x: auto`) but the mermaid/pixel `pre:has(...)` rule sets `overflow: visible`, so a wide diagram can spill out of the bubble.
- Mermaid and images already open the enlarge overlay (`visualLightbox.js` strips pixel width/height). Inline SVG is `max-width: 100%`.
- `.input-container` is a row. The base rule pads with `env(safe-area-inset-left/right)`. The `max-width: 768px` block replaces that with `padding: 10px 12px` and only restores the bottom inset. Side safe area is the leftover from #95.
- Send is 48×48 in that phone block. Attach, skills, and cancel use `.chat-icon-button` at 40×40. Clear is padding-sized, under 44px.
- `.app-outer` is `height: 100dvh` with `overflow: hidden` under 768px. `dvh` tracks browser chrome, not the soft keyboard, on iOS and on Chrome’s default `resizes-visual` mode. The focused field cannot be scrolled into view because the document does not scroll.
- Below 900px the agent grid still reserves `minmax(200px, 38vh)` for the workspace under the chat. The 768px block changes areas but does not reset those rows. A keyboard can shrink the chat row, which holds the composer, to nothing.

## Goals / Non-Goals

**Goals:**

- Phone transcript, welcome, errors, and visual blocks stay in the column. The pane scrolls on the block axis only.
- Composer and its common controls stay tappable and visible above the keyboard, including side safe areas.
- One source check for the CSS contract, one unit check for the keyboard measurement, then a production build and a phone-width layout check.

**Non-Goals:**

- Header, drawers, login visual design, global type scale.
- A second chat component, a chat-protocol change, or helper sub-agent UI.
- Replacing the enlarge overlay.

## Decisions

### 1. Clip the transcript’s inline axis; let wide blocks scroll inside themselves

At `max-width: 768px`, `.messages-container` uses `overflow-x: clip` and `overflow-y: auto`, plus `touch-action: pan-y`. `.chat-markdown`, bubbles, and the welcome block get `max-width: 100%` and `min-width: 0`. `pre`, tables, mermaid, and pixel grids get `max-width: 100%` and `overflow-x: auto`, including the `pre:has(.chat-mermaid)` rule that is `overflow: visible` today. Prose keeps `overflow-wrap: break-word`. Errors keep `anywhere`.

**Rejected:** `overflow-x: hidden` on `html` only. #95 already clips the document. The pane can still scroll sideways inside that clip.

**Rejected:** `word-break: break-all` on every message. It wraps prose by splitting ordinary words. `break-word` wraps only when a token would overflow.

### 2. Shrink the shell by the keyboard overlap, and do not double-count

Add `interactive-widget=resizes-content` to the existing viewport meta. On Chromium that shrinks the layout viewport with the keyboard, so `100dvh` already ends above the keys.

Safari ignores that token. A chat-only binding reads `visualViewport` and sets `--keyboard-inset` on the chat shell:

`covered = innerHeight - visualViewport.height - visualViewport.offsetTop`

Values under 1px become 0. When Chromium has already resized `innerHeight`, covered is 0, so the shell is not shrunk twice. The phone shell height is `calc(100dvh - var(--keyboard-inset, 0px))`. `.app-outer` also has a base `min-height: 100dvh`, which would ignore a shorter height, so the phone rule sets `min-height: 0` on `.app-outer` and `.app` and gives `.app` the same height. The binding runs from the Morph AI chat mount, not from MorphNotes.

While the inset is greater than 0, set `data-keyboard-open` on `documentElement`. At `max-width: 768px` that hides `.agent-workspace` and drops it from the agent grid, so the 200px workspace row cannot crush the composer. The workspace returns when the keyboard closes.

**Rejected:** `100dvh` alone. It does not track the keyboard, and #95 already deferred this.

**Rejected:** `position: fixed` composer pinned to the visual viewport. It leaves the grid, covers the transcript, and fights the in-flow composer.

**Rejected:** `env(keyboard-inset-height)` alone. It needs the Virtual Keyboard API overlay opt-in and is not available in Safari.

### 3. Phone composer: side safe area, wrap chips, 44px controls

Put the overrides in the last `max-width: 768px` block, the one that already sets header controls to 44px. An earlier phone block sets composer icons to 40px and replaces side padding with 12px; a new media block after this one would hide those header selectors from the existing source test, which reads from the last 768px block. Same specificity, later in the file, so 44px and the side insets win. After the padding shorthand is overridden, set composer padding-left and padding-right to `max(12px, env(safe-area-inset-left/right))`. Let `.input-container` wrap, and give attachment and assistant chips `flex-basis: 100%` so they sit above the field instead of beside the send button. Set `.chat-icon-button`, `.chat-icon-button--input`, and `.clear-input-button` to at least 44×44. Send stays 48×48. The enlarge close control becomes 44×44 in that same block.

**Rejected:** A separate mobile composer component. The same form is already the desktop composer.

### 4. Tests before CSS, then a real phone-width pass

A pure function owns the covered-pixel math, with cases for overlay (positive), already-resized layout (zero), and a scrolled visual viewport (zero). A source test locks the phone CSS contract: transcript `overflow-x: clip`, composer side safe-area, 44px composer controls, and the shell height using `--keyboard-inset`. The production build must succeed. A browser check at about 390px confirms wrap, vertical scroll, empty state, and that a simulated keyboard inset keeps the send control in view.

## Risks / Trade-offs

- [Workspace disappears while typing on a phone] → It is below the composer and would otherwise consume the visible height. It returns when the keyboard closes. Desktop is unchanged.
- [Clip hides a control that did not reflow] → Chips wrap to their own row. Code and diagrams scroll inside their box. The enlarge path stays.
- [`interactive-widget` also affects login] → One meta token. Login already uses `100dvh`. Resizing the layout viewport keeps that form above the keyboard. No login stylesheet rewrite.
- [visualViewport missing] → Covered pixels are 0. Chromium still has the meta token.
- [Inset listener runs after first paint] → First paint is already correct at inset 0. The listener only moves the shell after the keyboard opens.

## Migration Plan

Frontend static build only. Rollback is reverting the CSS, the viewport token, and the chat binding. No data migration.

## Open Questions

None. Portrait phone chat, the keyboard, tap size, and in-column visuals are specified. Header, login, and type scale stay in their own stories.

## Grill (proposer and reviewer)

Reviewed against `index.html`, `index.css`, `App.css` (transcript, composer, both 768px blocks, agent grid at 900px and 768px), `SkoolAiChat.js`, `MermaidBlock.js`, and `visualLightbox.js` before locking this design.

Alternatives:

1. Meta `interactive-widget=resizes-content` plus a visualViewport inset that is zero when the layout viewport already shrank, with phone CSS for clip, safe area, tap size, and workspace collapse. Chosen.
2. `100dvh` only. Rejected. The keyboard does not change `dvh` on the engines that cover the composer today.
3. A fixed composer. Rejected. It breaks the in-flow grid and can cover the transcript.
4. `env(keyboard-inset-height)` only. Rejected. Safari does not provide it.
5. A separate mobile chat tree. Rejected. The bug is CSS containment and one measurement, not a second UI.

Assumptions that failed the code check:

- “The composer already has side safe area.” The base rule does. The phone shorthand replaces the left and right insets with 12px.
- “Send is 44px, so the composer is fine.” Send is 48px. Attach, skills, and cancel are 40px in the phone block.
- “`100dvh` keeps the field above the keyboard.” The phone shell is `overflow: hidden`. The document cannot scroll the field up. #95 left this for this story.
- “The transcript cannot scroll sideways.” `overflow-y: auto` turns the other axis into `auto`.
- “The workspace is out of the way under 768px.” The 900px `minmax(200px, 38vh)` row still applies, so the keyboard can shrink the chat row to zero.
