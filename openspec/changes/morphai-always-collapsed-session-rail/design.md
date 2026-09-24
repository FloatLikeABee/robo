## Context

See proposal.md for motivation. Spec: `specs/morphai-session-rail/spec.md`.

Today Morph AI agent shell (`isAgentShell = !singleSession`) uses CSS grid `--agent-sessions: 220px`, or `48px` when `app--sessions-collapsed`. Collapsed **hides** `.sidebar-sessions` and only shows `+`. Header ☰ on ≥769px toggles that collapse via `sessionStorage` key `morphai-sessions-collapsed`. Session rows are titled list items with inline rename and a delete icon.

Still: dark-only Morph AI, same session APIs, `+` still calls `handleNewChat`.

## Goals / Non-Goals

**Goals:**

- Always use the compact rail (no 220px named list).
- Rail ≥78px (48 + 30).
- Keep sessions visible as hashed color buttons; + stays on top.
- Desktop ☰ no longer expands the rail.

**Non-Goals:**

- Session CRUD API, persistence, Files workspace, Skills, AI tools.
- Theme switch / light chrome.
- Changing `singleSession` embeds.
- Per-session user-picked colors (hash from id is enough).

## Decisions

### 1. Lock the agent shell to the collapsed grid column

**Choice:** Always apply `app--sessions-collapsed` on the agent shell. Set `--agent-sessions: 80px` (32px over 48; meets “at least 30px”). Stop toggling `sessionsCollapsed` from ☰ on ≥769px. Drop the expand path and the sessionStorage collapse key (unread leftover is fine to delete).

**Why:** User asked for always-collapsed and a wider strip. 80px is a round floor above 78px.

**Alternative:** Keep ☰ expand for “power users” — rejected, explicit “should not expand”. **Alternative:** exactly 78px — 80px is the same size class and easier in CSS.

### 2. Color buttons, not hidden sessions

**Choice:** Stop `display: none` on `.sidebar-sessions`. Each session is a square (or rounded-square) button. Color from a small dark-safe palette indexed by a stable hash of `session.id`. Current session: outline/ring using `--chat-accent`. `title` + `aria-label` = session title. Click selects; existing delete stays as a compact control (hover or nested icon) so non-default sessions remain removable. Inline rename is out of the rail (no width for it); tooltip is enough unless we keep double-click → `prompt` — skip prompt; YAGNI.

**Why:** User asked for colored buttons and no expansion. Hashing avoids a color picker.

**Alternative:** First letter of the title on the button — extra chrome, not asked. **Alternative:** identical gray pills — fails “different colors”.

### 3. + stays a plus, not “+ New chat”

**Choice:** Agent-shell rail always shows `+` (already the collapsed label). Do not close a mobile overlay in a way that hides the rail on desktop (desktop rail is persistent).

**Why:** Matches the existing collapsed control.

### 4. Narrow viewports

**Choice:** Keep the existing overlay pattern under 769px (☰ opens the same rail as a drawer). The drawer is still the color rail, not the 220px list.

**Why:** Phone still needs a way to reach sessions; expanding to named list would contradict the spec.

## Risks / Trade-offs

- [Many sessions overflow] → Vertical scroll on `.sidebar-sessions` (already there).
- [Hash collisions make two sessions the same color] → Palette of ≥8 hues; hash; if collision with a neighbor, pick next slot. Ceiling: 8–12 colors recycle; tooltip disambiguates.
- [Operators lose rename-in-place] → Titles remain on the session object and in tooltips; rename is a later change if asked.
- [☰ on desktop becomes a no-op] → Hide or no-op the collapse toggle on ≥769px; keep it for the mobile overlay.

## Migration Plan

1. CSS width + always-collapsed class; show color buttons; remove expand toggle.
2. Rollback: restore 220px / 48px toggle and titled rows.

## Open Questions

None. Palette size and 80px vs 78px are implementation details that still meet the spec.
