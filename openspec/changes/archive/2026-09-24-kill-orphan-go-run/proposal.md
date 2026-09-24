## Why

`./start-all.sh restart morph-api` kills the recorded PID and leaves the `go run` child listening on `:9090` with the Badger directory lock, so the new process cannot start. The same parent/child split affects every `go run`, `cargo run`, and npm dev server the launcher starts.

## What Changes

- Stop, `--stop`, and restart signal the service's whole process group (TERM, wait, then KILL), not only the recorded PID.
- The next start waits until that service's port has no listener.
- Help text and the local-run section of `docs/agents/12-build-deploy.md` describe this.
- A shell check under `scripts/` can be run later in CI.
- Existing commands (`--install`, `start`, `status`, `logs`, `stop`, `restart`) keep their names and arguments.

## Capabilities

### New Capabilities

- `dev-launcher-stop`: How `start-all.sh` stops and restarts a service so forked children do not keep the port or the Badger lock.

### Modified Capabilities

- None.

## Impact

- `start-all.sh` (all services it launches: Go, Rust, Python, npm).
- `docs/agents/12-build-deploy.md` local-run section only.
- `scripts/test-start-all-process-group.sh`.
- No API, schema, or dependency changes. No secrets. macOS and Linux.
