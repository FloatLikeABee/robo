## Purpose

MorphUtils Data Access must show a running Data Access UI or a start hint, not a blank browser connection-refused iframe.

## ADDED Requirements

### Requirement: Dead Data Access origin is explained in the shell
When the Data Access embed origin (`VITE_DATAX_URL`, default `http://localhost:5178`) is not accepting connections, MorphUtils MUST show an in-shell message that names Data Access and the launcher command to start it (`./start-all.sh start sharpreport-ui` or the equivalent MorphUtils start path). MorphUtils MUST NOT mount a Data Access iframe to that origin while the probe shows it is down.

#### Scenario: Origin refused
- **WHEN** an operator opens MorphUtils Data Access and nothing is listening on the Data Access embed origin
- **THEN** they see an in-shell start hint instead of the browser “localhost refused to connect” iframe page

#### Scenario: Origin up
- **WHEN** Data Access UI is listening on the embed origin
- **THEN** MorphUtils MUST load the Data Access iframe at that origin (session token query as today)

### Requirement: Starting MorphUtils starts Data Access UI
`./start-all.sh start morph-utils` MUST start MorphUtils UI and Data Access UI (`sharpreport-ui`). It MUST also start Data Access API (`sharpreport-api`) so a freshly started MorphUtils Data Access session can authenticate and load tables, not only an empty Vite shell.

#### Scenario: One launcher command
- **WHEN** an operator runs `./start-all.sh start morph-utils` on a machine where those services were stopped
- **THEN** MorphUtils UI and Data Access UI (and Data Access API) are started
- **AND** opening MorphUtils Data Access can reach the embed origin without a second remembered start command for `sharpreport-ui`
