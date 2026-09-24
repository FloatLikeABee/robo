# Proposal

## Why

The Morph monorepo had no GitHub Actions, so a Mermaid dependency change broke the Morph AI frontend production build and nothing stopped the merge. Maintainers need the same Go tests and Morph frontend build to run on every pull request and on `main`, without live AI keys.

## What Changes

- Add one workflow, `.github/workflows/ci.yml`, that runs on pull requests to `main` and on pushes to `main`.
- Run `go vet ./...` and `go test ./...` for Morph API (`morph`), Event Logs (`formx/backend`), Content Maker (`composerx/backend`), and `pkg/morphai`, each with the Go version from that module's `go.mod`.
- Run Morph frontend `npm ci`, unit tests, and `CI=true npm run build` on Node 22.
- Leave `MORPH_AI_API_KEY` unset. Do not skip tests that pass without an external service.
- Cancel a superseded run on the same pull request.
- Document the checks and local reproduction in `docs/agents/12-build-deploy.md` and add a CI status badge to the root README.
- Do not change Morph frontend source. ESLint failures under `CI=true` stay until the separate frontend-build fix lands.

## Capabilities

### New Capabilities

- `platform-ci`: GitHub Actions runs the four Go module checks and the Morph frontend check on pull requests and pushes to `main`, with stable check names, no AI credentials, and a documented local reproduction path.

### Modified Capabilities

- (none — `openspec list --specs` reports no existing capabilities)

## Impact

- New `.github/workflows/ci.yml`. Root `README.md` badge. CI section in `docs/agents/12-build-deploy.md`.
- Actions used: `actions/checkout`, `actions/setup-go`, `actions/setup-node`, pinned by commit SHA.
- Out of scope: Rust, Python, and the other frontends; fixing Mermaid or ESLint in `morph/frontend`; merging the pull request.
