## Purpose

Keeps the local Data Access UI talking to the SharpReport API on the same port the API actually binds, so MorphUtils Data tables do not fail with empty 500s from a dead proxy target.

## ADDED Requirements

### Requirement: Dev UI proxies to SHARPREPORT_PORT

When Data Access runs locally on port 5178, browser and SSR requests to `/api`, `/public`, and `/metabase` MUST be forwarded to `127.0.0.1` on `SHARPREPORT_PORT`. If `SHARPREPORT_PORT` is unset or empty, the target port MUST be 3050. `/api/messages` MUST still go to Morph on 9090.

#### Scenario: API listens on a non-default SHARPREPORT_PORT

- **WHEN** `SHARPREPORT_PORT` is set to a port other than 3050 and the SharpReport API is listening there
- **AND** a signed-in user opens Data Access Data tables (via MorphUtils or `localhost:5178`)
- **THEN** `GET /api/v1/auth/me` and `GET /api/v1/data-tables` reach that API
- **AND** those requests MUST NOT return an empty HTTP 500 from a proxy to 3050 when nothing is listening on 3050

#### Scenario: Default port when env is unset

- **WHEN** `SHARPREPORT_PORT` is unset
- **THEN** the Data Access UI MUST target `127.0.0.1:3050` for those API paths
