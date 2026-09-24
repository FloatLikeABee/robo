## Context

See proposal.md for why. The launcher (`start-all.sh`) backgrounds a subshell, `exec`s the service command, and stores that PID. The original stop calls `kill` on that PID only. A first cut exec'd `python3` `os.setsid` whenever `python3` was on `PATH`. This design rejects that as the primary path; the tasks switch the launcher to job control first.

Root cause, traced in the Go toolchain and reproduced:

- `go run` (`cmd/go/internal/base.RunStdin`) starts the compiled binary with `exec.Command` and does not set a new process group.
- The go command ignores only interrupt and SIGQUIT (`signal_unix.go`). SIGTERM kills `go run` and is not forwarded. The child is reparented and keeps the listen socket and the Badger flock on `./data/badger`.
- Killing only the parent leaves the listener. Signaling the parent's process group removes it. The same parent/child split is how `cargo run` and `npm start` work.
- On Darwin, Morph, Event Logs, and Content Maker APIs are already `go build` binaries (Badger `LC_UUID`). Linux still uses `go run`. npm and cargo fork on both.

Constraints: macOS bash 3.2, BSD `ps`/`lsof`, no `setsid` binary. Linux bash and GNU `ps`. Do not change command names. Do not kill unrelated services that still share a process group from an older launcher.

## Goals / Non-Goals

**Goals:**

- One process-group stop path for every service, on macOS and Linux.
- Restart waits until the group is gone and the service port has no listener, so Badger can be reopened.
- A shell test that fails if only the parent PID is killed.

**Non-Goals:**

- systemd or Docker process managers.
- Replacing `go run` with `go build` on Linux.
- Wiring the shell test into GitHub Actions (another change owns CI).
- Tracking grandchildren by polling after start.

## Decisions

### 1. Make bash job control the primary process group

Before backgrounding, enable monitor mode (`set -m`) so the command's PID is the process-group leader, then `set +m` in the parent. Ignore SIGHUP in the child before `exec` so a Go server, which keeps an inherited ignore, is not killed when the one-shot launcher exits.

Stop sends SIGTERM to that group (`kill -TERM -PGID`), polls until the group is empty (about 5 seconds), then SIGKILL. Signal a group only when the recorded PID equals its PGID and that PGID is not the launcher's PGID.

**Why this over a setsid helper as the primary path:** `set -m` is in bash 3.2. It does not need `setsid`, python, or perl. It was checked with a Go parent that forks a TCP listener: after the launcher exited, both processes were alive in that group; `kill -TERM -PGID` freed the port. The same check on Linux `./start-all.sh restart morph-api` showed `go run` and `/tmp/go-build…/exe/main` sharing one PGID, one listener on `:9090`, and no directory-lock line.

**Rejected: `setsid(1)` or `python3 -c os.setsid` as the only start path.** `setsid(1)` forks when the caller is already a group leader, so the PID written to the pid file exits and the real server is untracked. `command -v python3` can succeed on macOS when `python3` is a Command Line Tools stub that does not run; exec'ing it would fail the service even though perl is present. A setsid helper remains only the fallback below.

**Rejected: walk the process tree and kill descendants.** If the parent exits before the walk, children reparent to init and are missed. That is the original bug the moment `go run` dies.

**Rejected: kill every listener on the port and that listener's whole process group.** Frees `:9090` when the listener holds Badger, but services started by the old launcher share one PGID. Killing that group would stop every sibling and, if the foreground launcher is still the leader, the launcher itself. Port cleanup stays, and it tree-kills only that listener PID. Group-kill is reserved for a PID this launcher recorded.

**Rejected: `go build` on Linux for the Go APIs.** Removes one class of child and leaves npm and cargo broken. Darwin already builds those three APIs; group stop still covers them.

### 2. Fallback when `set -m` fails

Some non-interactive macOS shells refuse monitor mode. If `set -m` fails, replace the process with a new session without forking, in this order: perl `POSIX::setsid` (present on macOS), then python3 or python only if `python -c 'import os'` exits 0 within a couple of seconds (a Command Line Tools stub must not be exec'd), then `setsid(1)` only when this process is not already a group leader. If all of those fail, exec the command in place and stop falls back to the descendant walk plus the port backstop.

### 3. Wait for the group to die, then for the port

Badger's directory lock is released when the process exits, which is not the same moment the listen socket closes. Stop waits until the group has no processes, then polls `lsof` for a TCP listener (about 5 seconds). A remaining listener is tree-killed. Its process group is signaled only when that PID is one this launcher recorded. If `lsof` is not on `PATH`, the launcher warns once that it cannot prove the port is free. If the port is still taken, stop returns an error and restart does not start a second copy.

TIME_WAIT is not a listener. Gin sets `SO_REUSEADDR`, so a closed listener does not block the next bind.

### 4. One stop function for `stop`, `--stop`, and restart

`stop_all` must call the same per-service stop as `stop <name>`. Restart is that stop, then start. Start also refuses to exec while its port still has a listener.

### 5. Legacy processes are not group leaders

A PID file written by the old script names a process whose PGID is the old launcher. Stop must not signal that PGID. It kills that PID's descendants, then any listener on the service port. The next start is a group leader, so the following stop uses the group path.

### Design review

Proposer claim: "python setsid on Linux is enough, and macOS can use the same if python exists."

Reviewer objections, checked against `start-all.sh` and a live `go run`:

- macOS bash 3.2 has no `setsid`, and a python3 shim must not be the first exec. Primary path is now `set -m`. Perl is the first fallback, and python is used only after `import os` succeeds.
- A forking service (`go run`, npm, cargo) keeps children in the group unless the child calls `setsid`. Dev servers do not daemonize. A child that does is out of scope; the port backstop still kills the listener PID.
- A port that stays in LISTEN after SIGKILL fails the stop instead of stacking a second Badger opener.
- Group kill of a shared legacy PGID would take down siblings. The `pid == PGID` guard prevents that. A same-group orphan listener was killed by PID without killing the launcher (shell test).
- IPv6 `*:9090` is what Morph binds. `lsof -nP -tiTCP:9090 -sTCP:LISTEN` sees it.
- Ctrl+C in the foreground launcher only traps INT/TERM. Children in another group are not signaled by the keypress. The trap calls the same stop path. A service that ignores TERM costs a few seconds, then KILL.

The `pid == PGID` rule, the HUP ignore, and the group-then-port wait are what let this survive those cases. No open choice left that would change the spec or the task list.

## Risks / Trade-offs

- [Risk] `set -m` fails and perl/python/`setsid` are missing → Mitigation: descendant walk plus port listener kill. The shell test covers the job-control path; the fallback is the old behavior plus the port backstop.
- [Risk] A service ignores SIGTERM and exits slowly → Mitigation: about 5 seconds, then SIGKILL, then the port poll. Stop errors if the port is still taken.
- [Risk] An unrelated process already bound the dev port → Mitigation: restart tree-kills that listener and warns that it was not started by this launcher. It does not signal that listener's process group, so a foreign leader on `:3000` or `:8000` does not take its siblings with it.
- [Risk] A child calls `setsid` and leaves the group → Mitigation: out of scope for dev servers. The port backstop still clears the listener. Non-listening daemonized grandchildren are not tracked.
- [Risk] Test-only port override env vars are read by the launcher → Mitigation: names are specific (`START_ALL_PORT_OVERRIDE`, `START_ALL_PORT_OVERRIDE_NAME`). They are not written to `.env`.

## Migration Plan

Replace `start-all.sh` in place. The next `stop` or `restart` uses the new path. Rollback is reverting the script; pid files stay `name:pid`. No data migration.

## Open Questions

None. CI registration of `scripts/test-start-all-process-group.sh` waits for the existing CI change and does not change this design.

## Review follow-up

A review of the first process-group stop found three ways the stop path could signal the wrong processes. These rules are part of the design:

1. **Never signal PID 0, PID 1, or an empty PID.** `kill -0 0` is true because it checks the caller's process group. `child_pids` of `0` or of an empty value matches PPID 0 and returns PID 1, and `kill -TERM 0` signals that group. `pid_of`, `kill_recorded_pid`, and `kill_tree` reject those values before any signal. The launcher also refuses to record them.
2. **Fail closed when the launcher's own PGID is empty.** Group-kill runs only when the recorded PID equals its PGID, that PGID is greater than 1, and it is different from a known launcher PGID. If the launcher cannot read its own PGID, it tree-kills the recorded PID and does not signal a group.
3. **Group-kill only PIDs this launcher recorded.** `free_listening_port` used to call the recorded-pid group kill for every listener. An unrelated leader on a dev port would lose its whole group. Unrecorded listeners are tree-killed, with a warning that they were not started here.
4. **The shell test must fail if group kill becomes a parent-only kill.** The fixture forks a listener and a non-listening child. The non-listening child is asserted dead after `kill_recorded_pid` and inside `restart_one`, before port cleanup runs. Port cleanup cannot mask a parent-only kill, because it does not see that child.
5. **`restart morph-api` waits until `GET /health` succeeds** (about 90 seconds). Restart is not finished while the new process is still coming up, so a following status is one healthy Morph API.
6. **Service names match the whole field**, not a regular expression. `stop '.*'` does not select the last pid-file row. Unknown names are rejected before stop, start, restart, or logs.
7. **Missing `lsof` warns on stderr**, once, and does not look like an empty listener list on stdout.
