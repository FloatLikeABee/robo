## Why

Morph AI’s left chat-sessions column still expands into a named list, and the collapsed state is a 48px strip that hides sessions entirely. Operators need a permanently compact rail that stays usable: wide enough to hit, plus a visible session switcher that does not steal chat width.

## What Changes

- Morph AI main chat (not `singleSession` / embedded-only) keeps the left sessions column **always collapsed** as a vertical rail. It MUST NOT expand to the titled session list.
- Widen the rail by **at least 30px** vs the current 48px collapsed width (minimum **78px**).
- Keep a **+** control at the top to create a new session.
- Show existing sessions as **colored buttons** (distinct colors per session), not as a titled list and not by hiding the list.
- Header ☰ on desktop no longer expands/collapses this rail. Narrow viewports MAY still overlay the same rail.
- Dark-only. No API or session-store changes.

## Capabilities

### New Capabilities

- `morphai-session-rail`: Always-collapsed Morph AI chat session rail: wider than today’s 48px strip, + for new chat, sessions as distinct colored buttons.

### Modified Capabilities

- None. `openspec/specs/` has no existing Morph AI session-sidebar capability.

## Impact

- `morph/frontend/src/SkoolAiChat.js` (collapse toggle, session row markup, + label)
- `morph/frontend/src/App.css` (`--agent-sessions`, `.app--sessions-collapsed`, session row styles)
- Browser-only. No Go handlers, SQLite, or other products.
