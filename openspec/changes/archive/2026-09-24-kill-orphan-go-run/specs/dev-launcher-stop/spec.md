## Purpose

Defines how the local dev launcher stops and restarts a service so a forked child cannot keep the port or the Badger lock.

## ADDED Requirements

### Requirement: Restart clears the previous listener before bind
The launcher SHALL stop the previous process and its children before the replacement process binds the service port. After `restart` of Morph API, exactly one listener SHALL remain on that service's port.

#### Scenario: Morph API restart
- **WHEN** Morph API is running via the launcher and the operator runs restart for that service
- **THEN** the previous listener on port 9090 is gone before the new process binds
- **AND** restart waits until `GET /health` succeeds
- **AND** status reports a single healthy Morph API

#### Scenario: Badger lock is not held across restart
- **WHEN** Morph API is using its Badger data directory and the operator restarts it
- **THEN** the new process log does not contain a directory-lock error

### Requirement: Stop leaves no orphaned children
Stop and `--stop` SHALL terminate the service process and children it forked (including `go run` binaries, cargo binaries, and npm/node dev servers). After stop, that service SHALL NOT be listening on its port.

#### Scenario: Stop one service
- **WHEN** a launcher-started service is running and the operator stops that service
- **THEN** its recorded process and its forked listener are both gone

#### Scenario: Stop everything
- **WHEN** the operator runs stop with no service, or `--stop`
- **THEN** no launcher-started service child remains listening on that service's port

### Requirement: Stop works on macOS and Linux
The launcher SHALL reap forked children on Linux and on macOS bash 3.2 without requiring a `setsid` binary. Existing command names (`start`, `stop`, `restart`, `status`, `logs`, `--install`) SHALL keep working.

#### Scenario: No setsid binary
- **WHEN** the host has bash 3.2 and no `setsid` binary
- **THEN** stop still kills a child that the service forked after start

### Requirement: Operators can read the stop behavior
The launcher help and the local-run section of `docs/agents/12-build-deploy.md` SHALL state that stop and restart kill the service's children and wait until the port is free.

#### Scenario: Help text
- **WHEN** the operator runs the launcher with `--help`
- **THEN** the text describes process-group stop and the port wait

### Requirement: Stop refuses unsafe process ids
The launcher SHALL NOT signal PID 0, PID 1, or an empty PID from `pid_of`, `kill_recorded_pid`, or `kill_tree`. It SHALL NOT signal a process group when it cannot determine its own process group id.

#### Scenario: Pid file contains zero
- **WHEN** the pid file records a service with PID 0, PID 1, or an empty PID
- **THEN** stop does not signal that value
- **AND** processes in the launcher's own group keep running

#### Scenario: Launcher group is unknown
- **WHEN** the launcher cannot read its own process group id and a recorded service leads another group
- **THEN** stop does not signal that process group

### Requirement: Port cleanup does not group-kill unrecorded listeners
The launcher SHALL group-kill only PIDs it recorded. A listener on a service port that is not one of those PIDs SHALL be stopped as that process tree only, and the launcher SHALL warn that it did not start that listener.

#### Scenario: Foreign listener leads a group
- **WHEN** a process the launcher did not record is listening on a service port and leads a process group that contains another process
- **THEN** stop or port cleanup terminates that listener
- **AND** the other process in that group is still running
- **AND** the operator sees a warning that the listener was not started by this launcher

### Requirement: Service names match exactly
The launcher SHALL select pid-file rows by exact service name. A name that is not a known service SHALL be rejected before stop, start, restart, or logs changes any process.

#### Scenario: Regex-like name
- **WHEN** the operator stops `.*`
- **THEN** no recorded service is selected
- **AND** the command fails as an unknown service

### Requirement: Missing lsof is visible
When `lsof` is not on `PATH`, the launcher SHALL warn that port checks cannot see listeners and SHALL NOT treat that warning as a list of PIDs.

#### Scenario: lsof is absent
- **WHEN** a port check runs and `lsof` is not installed
- **THEN** the launcher warns once that it cannot prove the port is free
