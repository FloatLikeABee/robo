## 1. Rail layout

- [x] 1.1 Agent shell always uses the collapsed rail; `--agent-sessions` ≥78px (80px). Stop hiding `.sidebar-sessions`. Remove desktop expand-to-220px and the `morphai-sessions-collapsed` toggle.
- [x] 1.2 Header ☰ on ≥769px does not expand the rail; under 769px it still overlays the same color rail. `singleSession` still hides the rail.

## 2. Session buttons

- [x] 2.1 Rail shows **+** (new session) then one colored button per session. Hash `session.id` to a ≥8-color dark-safe palette; current session has a distinct ring. `title`/`aria-label` = session title. Click switches session; non-default sessions stay deletable without widening the rail.

## 3. Verify

- [x] 3.1 Morph AI frontend compiles. Browser: rail ≥78px, no named list, + creates a session, two sessions are different colors, current is marked, click switches.
