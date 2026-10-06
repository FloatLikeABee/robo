## Context

See proposal.md for why. Current behavior that shapes the approach:

- Morph AI agent shell can collapse the workspace (`app--workspace-collapsed`). Phone first load starts closed (`openspec/specs/morph-phone-chat`). The hide control is the header button labeled Hide workspace / Show workspace.
- Workspace tabs are ordered Notes & TODOs, then Context & Knowledge. The in-memory default is already Context & Knowledge, but a stored tab wins, and the first painted tab is Notes.
- Assistant replies render in `.assistant-bubble`. Diagrams and images already open a dark near-full-viewport lightbox. There is no enlarge path for long text.
- AI tools is an iframe drawer. `.ai-tools-frame-wrap` is `background: #fff`, so the wait before the frame paints is a white panel.
- Project (`morph-engi`) page tokens are already dark navy, but active chrome still uses violet (`#5B3FD6`, `bg-violet`) and several panels and the create modal use neutral gray (`#1a1a22`, `#2B2B33`). Rose is used for delete links and the danger confirm.
- MorphNotes Tasks is a wrapping card grid (`height: auto`, one-line description). Big notes is a list-plus-detail page. The address `/stories` already belongs to Timelines.

## Goals / Non-Goals

**Goals:**

- One always-visible workspace, with Context & Knowledge first.
- A text enlarge modal that reuses the dark stage idea without clipping the transcript.
- A dark AI tools wait that resets every time the modal opens.
- Project chrome aligned to the existing dark-blue tokens.
- Stick notes as a stable, equal-size, searchable sticker grid. Stories is a label change only.

**Non-Goals:**

- Restoring a Files workspace.
- Renaming `/big-notes`, `/case-tasks`, or `/stories`, or changing APIs.
- Clamping assistant replies to ten lines in the thread.
- Restyling the AI tools app inside the iframe after it has loaded.
- A light/dark switch.
- Turning Stories (Big notes) into stickers.

## Decisions

### 1. Always show the workspace, including the phone band

Remove the header toggle and stop applying the collapsed layout. Ignore a stored closed flag instead of migrating it.

**Why this over keeping the phone closed:** The only open control is the button the user is removing. Leaving the phone band closed with no button would make Notes and Knowledge unreachable on a phone. The crowded-phone bug was a narrow chat column beside an empty region. Full-width chat with the workspace in the lower band keeps that fix and matches "do not hide the panel."

**Alternative:** Wide screens always open, phone stays closed behind a different control. Rejected. It keeps a hide/show path the request removes.

### 2. Default tab is Context & Knowledge; a stored Notes choice still wins

Reorder the tab list. Select Context & Knowledge when `readWorkspaceTab` returns nothing. If the session has stored `notes`, open Notes.

**Why not force Context & Knowledge on every visit:** The request is a default, not a lock. Forgetting an explicit Notes choice would fight the existing per-session tab memory.

**Alternative:** Clear stored tabs so everyone lands on Context & Knowledge once. Rejected. It throws away a real choice along with the old order.

### 3. Enlarge by rendered height, and do not clamp the bubble

Show the control when the assistant bubble's scroll height is greater than ten times its computed line height. Newline count is the wrong signal: a wrapped paragraph can be many visual lines with no line breaks. The thread keeps the full reply. The modal is a dark, near-full-viewport stage with scroll, Escape, backdrop, and a close control. It does not go through the diagram lightbox, which only accepts an image or SVG.

**Why not collapse to ten lines:** The request is "able to be enlarged," not "hide the rest." Clamping would make the thread worse for people who already scrolled the reply.

### 4. Dark AI tools wait is the parent frame, reset on each open

Set the frame wrapper to the chat dark surface and show an indicator until the iframe load event. Set the loading flag true whenever the modal opens, including a reopen.

**Why not a ready message from the AI tools app:** The white panel is the parent's `#fff` wrapper, visible before the other origin paints. A cross-app ready signal would not fix that, and it cannot be required for a frame that fails to boot. After load, the embedded app's own theme is out of scope. If that app is still light, the indicator has already done its job.

### 5. Retint Project chrome; leave danger confirm red

Map page, header, cards, selected tab, primary button, and create modal to the existing navy and blue tokens (`#0c1220`, `#141c2a`, `#1a2332`, `#3b82f6`, `#38bdf8`). Drop violet as the active color and gray `#1a1a22` / `#2B2B33` fills. Delete links in the page use the muted text color. Only the confirm button on a destructive dialog stays rose, so delete does not look like a primary action.

**Why the red the user sees is treated as chrome, not a new theme:** There is no red page theme. The off-palette pieces are deep violet on active controls and rose on delete. Neutral gray modals read as a second theme next to the navy page.

### 6. Stickers are a CSS grid over the current task list

Keep the loaded task rows. Render them in `repeat(auto-fill, minmax(...))` with `aspect-ratio: 1` and an 8px gap. Show title plus a clamped multi-line body inside the square. Color is the palette slot for that record's index in the full list sorted by id (twelve fixed colors). Search hides rows and does not recolor them. The first twelve ids are unique; later rows repeat. That ceiling is acceptable: unique colors for an unbounded list is impossible. `id % 12` was rejected because two of the first twelve notes can share a slot.

Search filters the in-memory list on title and description. No new API.

**Why not random colors:** A new color on every render makes the board impossible to scan. Stable assignment is the same request with a board you can learn.

**Why Stories is not this grid:** The sticker sentence follows Stick notes. Big notes stays the current document page under the Stories label. `/stories` stays Timelines so the label does not steal that address.

## Risks / Trade-offs

- [Phone transcript is shorter because the band is always there] → Chat stays full width and scrolls. Do not bring back a hide control.
- [Stored Notes tab still opens Notes] → Tab order still leads with Context & Knowledge. Only a missing stored tab selects it.
- [Iframe load fires before the inner app looks ready] → The white parent frame is gone. Inner-app paint after load is out of scope.
- [More than twelve stickers repeat colors] → Palette of twelve; index assignment keeps the first twelve unique. Do not randomize.
- [Stories label collides with the old Timelines address name] → Visible Timelines label stays Timelines. Do not retarget `/stories`.

## Migration Plan

- No data migration. Ignore `morphai-workspace-open=0`.
- Rollback is reverting the UI change. Routes and APIs are untouched.

## Open Questions

None. Palette size is twelve. Phone workspace stays in the existing lower band.
