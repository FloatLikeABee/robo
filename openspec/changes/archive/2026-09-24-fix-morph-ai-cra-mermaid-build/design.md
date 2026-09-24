# Design

## Context

See proposal.md for why the bundle must build. Morph AI is Create React App plus CRACO (`morph/frontend/craco.config.js`). `MermaidBlock.js` lazy-imports `mermaid` and calls `mermaid.render`. `mermaid@11.17.2` exports `dist/mermaid.core.mjs`. Chunks such as `ganttDiagram-EL5Y4UJY.mjs` contain `import dayjs from "dayjs"` and `import dayjsIsoWeek from "dayjs/plugin/isoWeek.js"`. Those `.mjs` files have no `sourceMappingURL`. `dayjs` is already a mermaid dependency (`^1.11.21`) and is hoisted at `1.11.23`.

Evidence before choosing a fix (webpack, ESLint plugin removed, `ignoreWarnings` cleared):

- Dev config (`npm start`) and production config: 0 errors, 0 warnings.
- Same result with CRACO reduced to the previous `ignoreWarnings` only.
- `CI=true npm run build` on the unfixed tree failed on ESLint warnings, not on `Can't resolve 'dayjs'`.

## Goals / Non-Goals

**Goals:**

- `CI=true npm run build` exits 0.
- Mermaid flowcharts and sequence diagrams still render, and the lightbox still enlarges them.
- Keep CRA unejected.

**Non-Goals:**

- Changing the visual-first prompt contract.
- Reproducing the audit's nine webpack errors that this checkout does not emit.
- Silencing ESLint or webpack by setting `CI=false`.

## Decisions

### 1. Treat the ESLint warnings as the root cause of the CI failure

`CI=true` makes CRA fail the build on warnings. The warnings that fail are missing hook deps and unused bindings, not module resolution.

- `refreshKnowledge` in `HybridContextDrawer.js` becomes a `useCallback` with `[]` and is listed on the effect that calls it. It only uses setters, so the callback stays stable and the effect still runs when the drawer opens or the panel changes.
- `loadSessions` in `SkoolAiChat.js` becomes a `useCallback` depending on `isAgentShell` (itself `!singleSession`) and is listed on the mount effect. `pickRestoredSession` stays a module function so it is not recreated per render.
- `AdminDataGrid.js` lists `platformLabels?.term_facility` on the effect that builds place-column labels, so a label change refreshes those columns.
- Delete unused `kind` in `aiProgress.js`, and unused `IconButton` plus the unused `id` local in `CaseTasks.js`. The save payload never included that `id`.
- Move the `jest.mock` in `AgentWorkspace.lastSession.test.js` below its imports (`import/first`).

Rejected: `eslint-disable` comments and `CI=false`. Both hide the next warning. The new workflow runs `CI=true npm run build`, so a local `CI=false` would still fail in GitHub Actions.

Failure mode: a new effect dependency refetches more often. `refreshKnowledge` is stable. `loadSessions` changes only if `singleSession` changes. The place-label effect reruns when the facility term changes, which is the stale-closure bug the lint rule was pointing at.

### 2. Keep a narrow CRACO guard for mermaid ESM, and do not alias `dayjs`

The audit's `Can't resolve 'dayjs'` is the failure mode of webpack's `fullySpecified` resolver on mermaid `.mjs`. It does not reproduce here, including in the dev config. The guard stays so a future mermaid chunk that drops the `.js` extension, or a lockfile that fails to hoist `dayjs`, fails closed inside mermaid only:

- One module rule: `test: /\.mjs$/`, `include: /node_modules/mermaid/`, `resolve.fullySpecified: false`.
- Exclude that same directory from `source-map-loader`.
- Keep ignoring `Failed to parse source map` for other third-party maps (dompurify via jspdf).
- Declare `dayjs` as `^1.11.23` so the package is a direct dependency, matching what is already installed.

Rejected alternatives, checked against this tree:

1. **Alias `dayjs` to `node_modules/dayjs`.** Tried. CRA `ModuleScopePlugin` then errors: the alias is an import from outside `src/`. The production build fails even though the package exists.
2. **Downgrade or remove mermaid.** `MermaidBlock` renders by importing this package. Removing it deletes flowchart and sequence diagrams. A CJS downgrade is unnecessary: `11.17.2` already resolves, and the plugin imports are extension-qualified.
3. **Eject CRA, or set `GENERATE_SOURCEMAP=false`.** Eject forks `react-scripts`. Turning source maps off hides first-party maps as well as the third-party ENOENT the loader is not even emitting (no `sourceMappingURL` in the published chunks).

Failure mode of the guard: a bad `include` regex matches nothing and the guard is a no-op. That is acceptable, because the unguarded compile is already clean. A too-wide `fullySpecified: false` could resolve the wrong file outside mermaid; the rule is limited to `node_modules/mermaid`.

### 3. Cover render and lightbox in a unit test

`MermaidBlock.test.js` mocks `mermaid` and checks flowchart SVG, sequence SVG, render rejection falling back to the source, and a click opening the lightbox without the bubble `width`/`height`. That matches the enlarge rule: clone the SVG so it scales to the overlay.

## Risks / Trade-offs

- [Guard looks like it fixed a bug the current webpack does not show] → Design and the PR state that ESLint is the CI failure, and that both dev and prod compiles are clean with and without the guard.
- [Hook dependency edits change fetch timing] → Callbacks are stable; only the intended inputs are new deps.
- [`dayjs` direct dependency drifts from mermaid's range] → Pin the caret to `^1.11.23`, inside mermaid's `^1.11.21`.

## Migration Plan

No data migration and no run-command change. Rollback is reverting the frontend commit. `npm ci` picks up the lockfile `dayjs` entry.

## Open Questions

None. Dev versus production was measured: neither webpack mode emits the `dayjs` or source-map errors on this checkout.
