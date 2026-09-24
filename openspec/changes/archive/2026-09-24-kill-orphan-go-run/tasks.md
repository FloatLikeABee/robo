## 1. Failing test

- [x] 1.1 Add a shell case where `python3`, `python`, `perl`, and `setsid` on `PATH` are stubs that exit non-zero, the service command is an absolute interpreter, and stop must still reap the forked listener
- [x] 1.2 Run that case against the setsid-first launcher and confirm it fails because the stub is exec'd (no listener), not because of a typo

## 2. Launcher

- [x] 2.1 Start each service with `set -m` first so the recorded PID is the process-group leader, and ignore SIGHUP before exec
- [x] 2.2 If `set -m` fails, fall back in order: perl `POSIX::setsid`, a python that exits 0 on `import os` within a couple of seconds, then `setsid(1)` only when not already a group leader
- [x] 2.3 Keep group signal only when the recorded PID equals its PGID and that PGID is not the launcher's; otherwise kill the descendant tree. Wait until the group is empty, then until the service port has no listener. Use that path for `stop`, `--stop`, and restart
- [x] 2.4 Re-run the shell test and confirm the stub-PATH case and the existing restart cases pass

## 3. Docs and evidence

- [x] 3.1 Point launcher help and the local-run note in `docs/agents/12-build-deploy.md` at job control as the primary path, with no `setsid` binary required
- [x] 3.2 Restart Morph API via `./start-all.sh` several times and record one listener, a clean Badger log, and a single running status line
- [x] 3.3 Run `go test ./...` in `morph`
