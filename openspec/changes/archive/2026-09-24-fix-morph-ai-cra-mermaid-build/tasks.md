# Tasks

## 1. ESLint failures that CI promotes

- [x] 1.1 Stabilize `refreshKnowledge` and `loadSessions` with `useCallback` and list them on their effects; list `platformLabels?.term_facility` on the AdminDataGrid effect. Verify `npx eslint` is clean on those three files.
- [x] 1.2 Remove unused `kind`, `IconButton`, and `id`, and fix `import/first` in `AgentWorkspace.lastSession.test.js`. Verify `npx eslint` is clean on those files.

## 2. Mermaid resolve guard

- [x] 2.1 Keep `fullySpecified: false` and the source-map-loader exclude scoped to `node_modules/mermaid`, and declare `dayjs` at `^1.11.23`. Verify `craco.config.js` does not alias `dayjs` to an absolute path.
- [x] 2.2 Keep `MermaidBlock.test.js` covering flowchart, sequence, render failure, and lightbox. Verify `CI=true npm test -- --watchAll=false --testPathPattern=MermaidBlock` passes.

## 3. Integration

- [x] 3.1 Verify `CI=true npm run build` in `morph/frontend` exits 0, and `go test ./...` exits 0 in `morph` and `pkg/morphai`.
