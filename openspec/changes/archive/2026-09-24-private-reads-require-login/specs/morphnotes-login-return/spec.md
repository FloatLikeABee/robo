## Purpose

Send a signed-out MorphNotes visitor, and any Morph UI call that gets 401 from a private data API, to login and back to the page they were on, without an open redirect.

## ADDED Requirements

### Requirement: Signed-out MorphNotes opens login
A browser session with no Morph token that opens a MorphNotes path under `/morphdata` MUST be sent to `/login` with a return path for that page.

#### Scenario: Opening MorphNotes without a token
- **WHEN** a browser with no Morph token requests a `/morphdata` path
- **THEN** the UI navigates to `/login`
- **AND** the login URL carries the MorphNotes path as the return target

### Requirement: API 401 in the Morph UI returns to login
When the Morph UI receives HTTP 401 from a Morph API call other than login, including while the path is under `/morphdata`, it MUST clear the local session and navigate to `/login` with the current path as the return target.

#### Scenario: Morph Data 401 is not ignored
- **WHEN** a Morph Data page receives HTTP 401 from a private data API
- **THEN** the UI navigates to `/login` with the current path as the return target

### Requirement: Return path is a same-origin relative path
The return target MUST be accepted only when it is a same-origin relative path beginning with a single `/`. Values that are absolute URLs, protocol-relative (`//`), contain a scheme, or contain a backslash MUST be discarded. After a successful login the UI MUST navigate to the accepted path. A discarded value MUST land on `/`.

#### Scenario: Relative path is kept
- **WHEN** the return target is `/morphdata/research`
- **THEN** login navigates there after success

#### Scenario: Open redirect is rejected
- **WHEN** the return target is `https://evil.example/phish` or `//evil.example`
- **THEN** login does not navigate to that host
- **AND** the destination is `/`
