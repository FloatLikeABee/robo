# Design

## Context

See proposal.md for why. Requirements are in `specs/platform-ci/spec.md`.

The repo has no `go.work`. Relevant modules and their `go` lines:

| Module | Directory | `go` line | Lockfile for cache |
|--------|-----------|-----------|--------------------|
| Morph API | `morph` | 1.25.0 | `morph/go.sum` |
| Event Logs | `formx/backend` | 1.25.0 | `formx/backend/go.sum` |
| Content Maker | `composerx/backend` | 1.25.6 | `composerx/backend/go.sum` |
| morphai | `pkg/morphai` | 1.21.0 | none; `go.mod` only |

Those three apps `replace` sibling packages (`pkg/morphai`, `pkg/repoenv`, and others). A change under `pkg/morphai` can break `morph`, `formx/backend`, and `composerx/backend` even when those directories are untouched.

Root `package.json` sets `engines.node` to `>=20`. `morph/frontend` is CRA (`react-scripts` 5) and has unit tests (`*.test.js`). There is no `.nvmrc`. `MORPH_AI_API_KEY` is a local `.env` secret and is not required by the current Go tests or frontend unit tests.

Issue #15 asks for these checks and excludes a full cargo/bk matrix. PR #66 fixes the Morph frontend production build (eslint warnings that `CI=true` promotes, plus the Mermaid/dayjs compile). It is not merged. This change must not edit `morph/frontend` source.

## Goals / Non-Goals

**Goals:**

- One workflow whose five check names can be required branch checks without renaming later.
- Go vet and test per module, frontend install/test/build, no AI key.
- Module and npm caches that follow the lockfile.
- Actions pinned so a moved tag cannot change what required checks run.

**Non-Goals:**

- CI for Rust, Python, MorphUtils, Event Logs UI, Content Maker UI, or Project.
- Making `CI=true npm run build` pass by editing frontend source.
- Retries, path filters, or a `go.work` file.
- Merging the pull request.

## Decisions

### 1. One workflow file, not one workflow per app

- **Choice**: `.github/workflows/ci.yml` holds all five checks. The README badge points at that file.
- **Rationale**: v1 is five checks with one cancel-on-same-PR rule. One file keeps concurrency, permissions, and triggers in one place.
- **Rejected**: A workflow per app (`ci-morph.yml`, `ci-formx.yml`, …). Each file would repeat triggers and the cancel group, the badge would cover only one of them, and the check set would drift. The repo has more apps than this slice; splitting now does not match the issue's out-of-scope note.

### 2. Go matrix with fixed display names, not copied jobs and not one loop

- **Choice**: One `go` job, `strategy.fail-fast: false`, `name: Go / ${{ matrix.name }}`, with `include` rows whose `name` values are `Morph API`, `Event Logs`, `Content Maker`, and `morphai`. The frontend stays a separate job named `Morph frontend`.
- **Rationale**: The matrix is the readable form of four nearly identical steps. `fail-fast: false` lets every module report when one fails. Separate job `name` values are what GitHub shows as check names.
- **Rejected**: Four copy-pasted jobs. They would drift (one module forgets `unset`, another forgets `go vet`). A single job that loops modules in bash would publish one check, so branch protection could not require "Go / Event Logs" on its own, and a failure would be harder to see.

### 3. No path filters

- **Choice**: Every run reports all five checks. No `paths` / `paths-ignore` filters.
- **Rationale**: These names are meant to become required checks. GitHub does not treat a skipped required check as success under classic branch protection, so a docs-only PR would be stuck. Filters are also wrong for this repo's `replace` graph: editing `pkg/morphai` must still test `morph`, `formx/backend`, and `composerx/backend`. A filter on each module directory would miss that. The jobs are short (about one to two minutes once caches are warm).
- **Rejected**: Skip unaffected jobs. Faster on SharpReport-only or bk-only PRs, and it would hide a broken consumer of a shared package or block merge when a required check is skipped.

### 4. Pin actions by commit SHA, with the release tag in a comment

- **Choice**: Pin the three actions to the commits of the releases already selected, and comment the tag:
  - `actions/checkout` `3d3c42e5aac5ba805825da76410c181273ba90b1` (v7.0.1)
  - `actions/setup-go` `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` (v7.0.0)
  - `actions/setup-node` `820762786026740c76f36085b0efc47a31fe5020` (v7.0.0)
- **Rationale**: Required checks should not change because someone moves a major tag. v7 is the current major for all three. The comment keeps upgrades readable.
- **Rejected**: Floating `@v7` tags (what the first draft used). Convenient, and a compromised or retagged release would change CI without a diff in this repo. Dependabot can still bump a pinned SHA later; that bump is a reviewable diff.

### 5. Check names are a contract

- **Choice**: The five strings in the spec are the job names. A comment at the top of the workflow says not to rename them without updating branch protection. Matrix `name` values stay the product labels, not the directory paths (`formx/backend` must not become the check name).
- **Rationale**: Renaming a job detaches it from a required check and the PR looks like the check never ran. Directory paths are worse labels and would change if a folder is renamed for packaging reasons.
- **Rejected**: Naming checks after directories or after the workflow filename. The workflow name stays `CI`; the check names stay the job names.

### 6. Cache on lockfiles; do not use `pull_request_target`

- **Choice**: `actions/setup-go` caching stays on (modules and build outputs), keyed by `cache-dependency-path` (`go.sum`, or `go.mod` for `pkg/morphai` because it has no third-party requirements and no `go.sum`). `actions/setup-node` uses `cache: npm` with `morph/frontend/package-lock.json`. Triggers stay `pull_request` and `push`. `permissions: contents: read`. Checkout uses `persist-credentials: false`.
- **Rationale**: GitHub cache entries written on a feature branch are not restored on the default branch. A pull request therefore cannot poison `main`'s cache. `pull_request_target` would run untrusted code with the base context and is a known cache-poisoning and token-theft path; we do not need it. Lockfile hashes mean a dependency change misses the old cache. `persist-credentials: false` avoids leaving the checkout token in `.git/config` for later steps that only run tests.
- **Rejected**: Disabling caches. That would make every run download modules again and drops a requirement of the original CI request. Also rejected: a hand-rolled `actions/cache` key that ignores `go.sum` / the npm lockfile. That key would restore a stale tree after a dependency bump. Also rejected: `go test -count=1` to bypass Go's test cache. The tests are deterministic and local; `-count=1` throws away the build cache this design keeps.

### 7. No retries; do not paper over the frontend build

- **Choice**: Each command runs once. If a future Go test truly needs an external service, skip it in the test with a reason rather than adding a workflow retry. Do not edit `morph/frontend` to clear eslint warnings. `CI=true npm run build` stays in the frontend job, so the job stays red until PR #66 (or an equivalent fix) is on `main` and this branch is rebased.
- **Rationale**: Current Go tests and the frontend unit tests passed with `MORPH_AI_API_KEY` unset (local run and CI run 35951120620). Retries would hide a real flake. The frontend failure on that run was eslint warnings promoted by `CI=true`, not a missing secret. Duplicating #66's source edits here would conflict with that review.
- **Rejected**: `continue-on-error` on the production build. The check would look green while the build is broken, which is the failure this CI exists to catch. Also rejected: splitting unit tests into a sixth check so tests can be green while the build is red. That renames the required set away from the five names already published.

### 8. Runner image and Node major

- **Choice**: `ubuntu-24.04` rather than `ubuntu-latest`. Node stays major `22` (satisfies `engines.node` `>=20`; frontend unit tests were run on Node 22).
- **Rationale**: The first CI run warned that `ubuntu-latest` moves to Ubuntu 26 on 2026-10-19. Pinning 24.04 keeps the required checks on the image they already passed on until we choose to move. Node 22 is the major used to verify the unit tests; pinning a patch would churn for no spec change.
- **Rejected**: Staying on `ubuntu-latest` for "always current." The image flip is dated and unreviewed. Also rejected: Node 20 just because the README says "20+ recommended." The engine range allows 22, and 22 is what the tests were verified on.

## Risks / Trade-offs

- [Required check renamed by accident] → Names are in the spec, the workflow comment, and the docs table. A rename is a spec change.
- [SHA pin goes stale or is revoked] → Comment records the tag. Upgrading is a one-line diff per action, reviewed like any other change.
- [`pkg/morphai` has no `go.sum`] → Cache key is `go.mod`. The module has no third-party requirements, so a missing sum file is expected, not a failed cache.
- [Same-branch cache restore after a bad commit] → A later commit on that branch can restore a cache saved by an earlier commit with the same lockfile hash. That does not cross onto `main`. Lockfile changes miss. We do not run untrusted code via `pull_request_target`.
- [Frontend check red until #66 merges] → Expected. Go checks stay green. Do not weaken `CI=true`.
- [Shared package edit still runs all four Go modules] → Extra minutes, accepted so `replace` consumers are actually tested.
- [Flake in sqlite or httptest] → Do not add retries. Fix or explicitly skip with a reason in the test, and list the skip in the PR.

## Migration Plan

1. Land the workflow on the existing CI pull request (do not merge from this change).
2. Rebase onto current `main`. If #66 is already in `main`, the frontend check should go green with the rebased tree. If it is not, leave the frontend red and report the five check states.
3. After merge, branch protection can require the five names. Rollback is deleting `.github/workflows/ci.yml` (and the badge); no data migration.

## Open Questions

None. #66's merge timing does not change the workflow; it only changes whether the frontend check is green after rebase.
