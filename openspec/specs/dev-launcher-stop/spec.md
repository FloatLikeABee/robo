# dev-launcher-stop Specification

## Purpose

Defines how the local dev launcher stops and restarts a service so a forked child cannot keep the port or the Badger lock.

## Requirements

### Requirement: Restart clears the previous listener before bind
The launcher SHALL stop the previous process and its children before the replacement process binds the service port. After `restart` of Morph API, exactly one listener SHALL remain on that service's port.

#### Scenario: Morph API restart
- **WHEN** Morph API is running via the launcher and the operator runs restart for that service
- **THEN** the previous listener on port 9090 is gone before the new process binds
- **AND** status reports a single running Morph API

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
