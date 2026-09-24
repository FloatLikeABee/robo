# Tasks

## 1. Phone collapsed column

- [x] 1.1 Add a failing frontend test that the `max-width: 768px` rule for `.app.app--agent.app--workspace-collapsed` sets `grid-template-columns` to a single `minmax(0, 1fr)` track, and that the unscoped collapsed rule still sets `var(--agent-sessions)`. Run it and confirm it fails because the phone rule does not set columns.
- [x] 1.2 Add that longhand on the phone collapsed rule in `morph/frontend/src/App.css` and re-run the test until it passes. Do not change keyboard-inset height or the keyboard rule that hides `.agent-workspace`.

## 2. Phone workspace default

- [x] 2.1 Add `initialWorkspaceOpen({ stored, phone })` tests: unset + phone is closed; unset + desktop is open; stored `1` stays open on a phone; stored `0` stays closed. Run them and confirm they fail because the helper does not exist.
- [x] 2.2 Implement the helper in `AgentWorkspace.js` and use it from `SkoolAiChat` for the agent shell, guarding a missing `matchMedia`. Re-run the helper tests until they pass. Leave `readWorkspaceOpen()` unset-means-open.

## 3. Phone include chips

- [x] 3.1 Add a failing test that the phone query sets `.agent-include-bar` to `display: none`, and that the base rule is still `display: flex`. Run it and confirm it fails.
- [x] 3.2 Add the phone `display: none` rule and re-run the test until it passes. Do not remove the bar from desktop markup.

## 4. Integration

- [x] 4.1 Run `CI=true npm test -- --watchAll=false` and `CI=true npm run build` in `morph/frontend`. Both exit 0.
- [x] 4.2 Measure a phone-width fixture (about 390px and 430px) of the collapsed agent shell: header, transcript, and composer widths match the viewport, the assistant bubble is wider than a few letters, and the workspace band is not in the grid. Record the widths.
