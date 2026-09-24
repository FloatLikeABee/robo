# Tasks

## 1. Harden the workflow

- [x] 1.1 Pin `actions/checkout`, `actions/setup-go`, and `actions/setup-node` to the commit SHAs in design.md, set checkout `persist-credentials: false`, and set `runs-on` to `ubuntu-24.04`. Verify `.github/workflows/ci.yml` contains those three SHAs and does not contain `@v7` or `ubuntu-latest`.
- [x] 1.2 Keep the five check names (`Go / Morph API`, `Go / Event Logs`, `Go / Content Maker`, `Go / morphai`, `Morph frontend`) and add a workflow comment that those names are the required-check contract. Verify the job `name` fields still produce exactly those strings and that there is no `paths` filter.

## 2. Keep the CI docs aligned

- [x] 2.1 Update the CI section in `docs/agents/12-build-deploy.md` so it still lists the five checks, the local commands with `MORPH_AI_API_KEY` unset, and the decision that every check runs on every pull request (no path filters). Verify by reading that section against `specs/platform-ci/spec.md`.

## 3. Rebase and verify

- [x] 3.1 Rebase this branch onto current `origin/main`. Verify `git merge-base HEAD origin/main` equals `origin/main`. Do not edit `morph/frontend` source.
- [x] 3.2 With `MORPH_AI_API_KEY` unset, run `go vet ./...` and `go test ./...` in `morph`, `formx/backend`, `composerx/backend`, and `pkg/morphai`. Verify each command exits 0.
- [x] 3.3 In `morph/frontend`, run `CI=true npm test -- --watchAll=false`. Verify the unit tests exit 0. Do not change frontend source to make `CI=true npm run build` pass; record that build's result from the pull request checks after push.
