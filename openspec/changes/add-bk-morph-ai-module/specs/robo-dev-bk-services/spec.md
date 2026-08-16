## Purpose

Run BK API and UI as first-class services in the robo monorepo dev launcher alongside Morph, Morph Utils, and other apps.

## ADDED Requirements

### Requirement: start-all defines bk-api and bk-ui services

`start-all.sh` SHALL register services `bk-api` and `bk-ui` in the default `ALL_SERVICES` list.

#### Scenario: Full stack start includes BK

- **WHEN** a developer runs `./start-all.sh` or `./start-all.sh start all`
- **THEN** `bk-api` and `bk-ui` are started with the other default services

### Requirement: bk-api runs the FastAPI backend

`bk-api` SHALL start from `bk/` by running `python main.py`, logging to `.robo-dev/logs/bk-api.log`.

#### Scenario: API health endpoint

- **WHEN** `bk-api` has started successfully
- **THEN** the BK API is reachable at `http://localhost:8000` (or the port configured in `bk/.env`)

### Requirement: bk-ui runs the React frontend

`bk-ui` SHALL start from `bk/frontend` via `npm start`, logging to `.robo-dev/logs/bk-ui.log`.

#### Scenario: UI reachable

- **WHEN** `bk-ui` has started successfully
- **THEN** the BK UI is reachable at `http://localhost:3000` (unless overridden by `PORT`)

### Requirement: bk alias starts both services

`./start-all.sh start bk` (and `restart bk`, `stop bk`, `status` for alias `bk`) SHALL target both `bk-api` and `bk-ui`.

#### Scenario: Alias resolution

- **WHEN** a developer runs `./start-all.sh start bk`
- **THEN** both `bk-api` and `bk-ui` are started

### Requirement: status and list document BK URLs

`./start-all.sh status` and `list` SHALL show BK service names and default URLs (`http://localhost:8000` for API, `http://localhost:3000` for UI).

#### Scenario: status output

- **WHEN** a developer runs `./start-all.sh status`
- **THEN** lines for `bk-api` and `bk-ui` include their local URLs
