## Purpose

Expose the BK (Ground Control) workspace from Morph AI’s header app bar with the same session handoff pattern as Morph Utils.

## ADDED Requirements

### Requirement: Morph AI header lists BK after Morph Utils

The Morph AI chat header app link row SHALL list apps in this order: Morph Data, Morph Utils, BK.

#### Scenario: Header order on chat page

- **WHEN** a signed-in user views the Morph AI chat page header
- **THEN** Morph Data appears first, Morph Utils second, and BK third

### Requirement: BK link opens the BK UI with session token

When a UsersPanel session token is available, the BK header link SHALL include `userspanel_token` as a query parameter on the BK UI base URL.

#### Scenario: Token appended when session exists

- **WHEN** the user clicks the BK header link and a Morph/UsersPanel session token is present
- **THEN** the browser navigates to the configured BK UI URL with `userspanel_token` set to that token

#### Scenario: No token when unsigned

- **WHEN** no session token is available
- **THEN** the BK header link SHALL still open the configured BK UI URL without adding `userspanel_token`

### Requirement: BK URL is configurable

The BK UI base URL SHALL be configurable via `REACT_APP_BK_URL` with default `http://localhost:3000`.

#### Scenario: Default local URL

- **WHEN** `REACT_APP_BK_URL` is unset in the Morph frontend build
- **THEN** the BK header link targets `http://localhost:3000`

### Requirement: BK header shows icon and label

The BK header control SHALL display the label **BK** and an icon consistent with other Morph AI header app links.

#### Scenario: Visible label

- **WHEN** the header app links render
- **THEN** the third app link text is **BK**
