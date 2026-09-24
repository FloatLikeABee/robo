# Proposal

## Why

`CI=true npm run build` in `morph/frontend` fails, so Morph AI cannot ship a production bundle. GitHub Actions treats the CRA warning pass as errors. The audit also reported mermaid `dayjs` resolve errors and source-map ENOENT under mermaid chunks; those webpack failures do not reproduce on this checkout, but the ESLint failures do.

## What Changes

- Clear the ESLint warnings that `CI=true` promotes to errors in the Morph AI CRA app, with real hook dependencies and deleted dead variables.
- Keep a narrow CRACO guard so mermaid's strict ESM `dayjs` imports and unpublished source maps cannot fail the bundle if resolution regresses.
- Declare `dayjs` directly at the version mermaid already requires.
- Keep mermaid diagram rendering and the enlarge lightbox as they are.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- (none — `openspec/specs/` has no capabilities. Mermaid rendering behavior is unchanged, so this change sets `skip_specs: true` rather than inventing a requirement.)

## Impact

- `morph/frontend/craco.config.js`, `morph/frontend/package.json`, `morph/frontend/package-lock.json`
- ESLint-only edits in `HybridContextDrawer.js`, `SkoolAiChat.js`, `AdminDataGrid.js`, `aiProgress.js`, `CaseTasks.js`, and `AgentWorkspace.lastSession.test.js`
- New unit coverage in `MermaidBlock.test.js`
- No API, prompt-contract, or run/build command changes
