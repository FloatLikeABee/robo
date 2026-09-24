# Design

## Context

See proposal.md for why. Spec: `specs/morph-phone-chat/spec.md`.

Measured on the current agent shell (Chromium, viewport 500px, which is inside `max-width: 768px`), with the real `App.css` rules and the agent DOM (`display: contents` on `.chat-container`):

- Workspace **open**: grid columns `500px`, areas `"head" "chat" "workspace"`. Header and transcript are full width. The workspace row is 212px, which is `28dvh` from the later phone rule. That is the crowded bottom band.
- Workspace **closed** (`.app.app--agent.app--workspace-collapsed`): grid columns `80px 420px`, areas `"head" "chat"`. Header width 80px, transcript width 80px, assistant bubble width 43px. The reply wraps a few letters per line. The right track is empty. `80 / 390 ≈ 20%`, which matches the iPhone strip.

The shell itself is `width: 100%`. This is not a flex item that failed to grow, and not a missing `width` on `.app-outer`.

Cascade that produces the closed-state columns:

- Unscoped `.app.app--agent.app--workspace-collapsed` (specificity 0,3,0) sets `grid-template-columns: var(--agent-sessions) minmax(0, 1fr)` with `--agent-sessions: 80px`, and two-column areas (`head head` / `sessions chat`).
- At `max-width: 768px`, `.app.app--agent` (0,2,0) sets one column. It loses to the collapsed rule.
- The phone collapsed rule in that same query sets one-column areas (`head` / `chat`) and rows, and does not set columns. Areas update; columns stay `80px` + `1fr`. Named areas occupy only the first track, so the conversation sits in the 80px column.

`readWorkspaceOpen()` returns true when `morphai-workspace-open` is unset, so the first phone visit paints the bottom band. The composer `.agent-include-bar` (Notes / Knowledge) is a separate include-on-send toggle. It uses the same words as the workspace tabs and sits above the composer on every agent load.

Keyboard inset is `calc(100dvh - var(--keyboard-inset, 0px))` on `.app-outer` and `.app` inside the phone query, and `html[data-keyboard-open]` hides `.agent-workspace`. Those rules do not set grid columns.

## Goals / Non-Goals

**Goals:**

- Closed workspace at the phone breakpoint: one flexible column, header and transcript at the layout width.
- Phone first load with no saved choice: workspace closed, transcript not shortened by a workspace row.
- Phone composer: no Notes/Knowledge chips. Panel tabs stay the labeled path. Include flags stay default on.
- Desktop side-by-side shell, saved open/closed choice, and keyboard-inset / safe-area rules stay.

**Non-Goals:**

- A new drawer, overlay, or workspace visual design.
- Changing how wide the band is after the operator opens it (`minmax(0, 28dvh)` stays).
- MorphUtils, satellites, or a desktop grid change.
- A phone control to turn notes or knowledge off for one send.

## Decisions

### 1. Set one column on the phone collapsed rule

Inside `@media (max-width: 768px)`, on `.app.app--agent.app--workspace-collapsed`, set `grid-template-columns: minmax(0, 1fr)` next to the one-column areas. Same specificity as the unscoped collapsed rule, later in the file, so it wins only in that query. Desktop and the 769–900px band keep `80px + 1fr` with areas that span both columns.

The breakpoint stays 768px, where the one-column areas are already introduced. The spec's check is ≤430px. Fixing only at 430px would leave 431–768px on the broken tracks.

Prose already uses `overflow-wrap: break-word` on bubbles. The per-letter wrap is the 80px track, not the wrap mode. Do not switch bubbles to `anywhere` or `break-all`.

**Rejected:** `width: 100%` or `!important` on the header and transcript. The probe showed those boxes are 80px because they are placed in track 1. Stretching them does not delete the empty track.

**Rejected:** dropping columns off the unscoped collapsed rule so the phone `.app.app--agent` column wins. Desktop closed layout needs that 80px sessions track.

**Rejected:** `grid-template` shorthand. The file uses longhands. One longhand next to the areas that already disagree is the whole fix. A shorthand would also reset rows that the later keyboard block sets on purpose.

### 2. Phone default closed, saved choice kept

Add `initialWorkspaceOpen({ stored, phone })`:

- stored `'0'` / `'1'` → false / true, on phone and desktop
- stored unset and `phone` → false
- stored unset and not phone → true (today's `readWorkspaceOpen()` default)

`SkoolAiChat` reads the storage key and, when `window.matchMedia` exists, `matchMedia('(max-width: 768px)')` once for the initial state. If `matchMedia` is missing, treat it as not-phone so Jest does not throw and the desktop default remains. `readWorkspaceOpen()` stays as the desktop-default helper so its meaning does not change. The header toggle and `writeWorkspaceOpen` stay.

**Rejected:** default false for every width. That closes the desktop side panel for everyone who has never toggled it.

**Rejected:** ignore a saved `'1'` on phone. The spec keeps an explicit open choice. The crowded first paint is the unset default, not a stored open.

**Rejected:** a `position: fixed` bottom drawer. Opening already toggles `app--workspace-collapsed`. A drawer would cover the composer and sit on top of the keyboard-inset rules this change does not edit. The open band stays the existing `28dvh` row.

### 3. Hide the include bar on the phone breakpoint only

`.agent-include-bar { display: none }` inside the phone query. The bar is not removed from desktop. `includeNotes` and `includeKnowledge` stay default true, so a phone send still includes both. The workspace tabs are the remaining Notes / Knowledge labels.

**Rejected:** delete the bar in all layouts. Desktop uses it to exclude notes or knowledge from the next send.

**Rejected:** turn the chips into tab openers and also keep the tabs. That is a second path to the same place, which is the duplication the issue names.

**Rejected:** move the toggles into the panel. That keeps a phone exclude control, but it threads new props through `AgentWorkspace` for a control the phone story is removing from the composer. Out of scope for this fix.

### 4. Discarded causes

- Flex / `width: 100%` not filling the viewport. The shell measured 500/500 before any fix. The empty region is the unused `1fr` track.
- `app--sessions-collapsed` (always on the agent shell, no CSS).
- `display: contents` failing. Header, chat column, and workspace were separate grid items in the probe.
- Sidebar `min-width: 80px` putting the rail back in flow. On the phone the sidebar is `position: fixed` at x=-84. The 80px is the grid track from `--agent-sessions`, not the fixed rail.

## Risks / Trade-offs

- [Default closed lands before the column fix, so the first phone paint is the 80px strip] → Both edits are in this change. The layout check measures the collapsed shell, which is the phone default after this change.
- [Phone cannot turn notes or knowledge off for one send] → Accepted. Include stays on. Desktop chips stay. The spec only requires that a send can still include them.
- [A saved open choice still shows the `28dvh` band] → Accepted. The operator asked. Closing uses the existing header control.
- [Resizing a desktop window under 768px does not auto-close a session that started open] → Accepted. Initial state is what the spec requires. The toggle still works.
- [Jest has no `matchMedia`] → The pure function takes `phone` as an argument. The component passes the media result.

## Migration Plan

No stored-data migration. The localStorage key and values stay. Rollback is reverting the CSS longhand, the initial-state helper, and the include-bar rule.

## Open Questions

None. The column mismatch, the default-open helper, and the chip removal are decided above.
