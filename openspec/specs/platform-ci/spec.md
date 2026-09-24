# platform-ci Specification

## Purpose

Runs the Morph platform's Go tests and the Morph frontend production build on every pull request and on main, so a broken build cannot merge unnoticed, without calling a live AI provider.

## Requirements

### Requirement: CI triggers

The repository MUST run the platform CI workflow on every pull request targeting `main` and on every push to `main`.

#### Scenario: Pull request to main

- **WHEN** a pull request targeting `main` is opened or updated
- **THEN** the platform CI workflow starts

#### Scenario: Push to main

- **WHEN** a commit is pushed to `main`
- **THEN** the platform CI workflow starts

### Requirement: Stable check names

The workflow MUST report exactly these check names, and MUST report each of them on every run: `Go / Morph API`, `Go / Event Logs`, `Go / Content Maker`, `Go / morphai`, and `Morph frontend`. A run MUST NOT skip one of these checks because the pull request touched only some paths.

#### Scenario: Unrelated path still reports every check

- **WHEN** a pull request targeting `main` changes only files outside the Go modules and `morph/frontend`
- **THEN** all five checks are still reported

### Requirement: Go module checks

`Go / Morph API` MUST run `go vet ./...` and `go test ./...` in `morph`. `Go / Event Logs` MUST run the same commands in `formx/backend`. `Go / Content Maker` MUST run them in `composerx/backend`. `Go / morphai` MUST run them in `pkg/morphai`. Each check MUST use the Go version declared in that directory's `go.mod`. `MORPH_AI_API_KEY` MUST be unset for these commands. A test that passes without an external service MUST NOT be skipped.

#### Scenario: Tests run without an AI key

- **WHEN** a Go check runs
- **THEN** `MORPH_AI_API_KEY` is unset and `go vet ./...` and `go test ./...` run in that check's module directory

### Requirement: Morph frontend check

`Morph frontend` MUST run `npm ci` in `morph/frontend`, then `CI=true npm test -- --watchAll=false`, then `CI=true npm run build`. It MUST use Node.js 22. This capability MUST NOT require a change to Morph frontend source in order to exist.

#### Scenario: Frontend commands run in order

- **WHEN** the Morph frontend check runs
- **THEN** it installs from the lockfile, runs the unit tests once, and then runs the production build with `CI=true`

### Requirement: Superseded runs cancel

A newer workflow run for the same pull request MUST cancel the older in-progress run for that pull request. Runs for different pull requests MUST NOT cancel each other.

#### Scenario: New commit on the same pull request

- **WHEN** a pull request already has a platform CI run in progress and a new commit is pushed to that pull request
- **THEN** the older run is cancelled and the new run proceeds

### Requirement: No secrets in the workflow

The workflow MUST NOT contain API keys, tokens, or other secrets, and MUST NOT read `MORPH_AI_API_KEY` from repository secrets.

#### Scenario: Workflow file is public-safe

- **WHEN** a reviewer reads `.github/workflows/ci.yml`
- **THEN** it contains no secret values and does not map `MORPH_AI_API_KEY` from secrets

### Requirement: Visible status and local reproduction

The root README MUST show a status badge for this workflow. `docs/agents/12-build-deploy.md` MUST describe what each check runs and the commands to reproduce each check locally with `MORPH_AI_API_KEY` unset.

#### Scenario: Operator reproduces a check locally

- **WHEN** an operator follows the CI section in `docs/agents/12-build-deploy.md`
- **THEN** the documented commands match the five checks and tell them to leave `MORPH_AI_API_KEY` unset
