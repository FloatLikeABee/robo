## Purpose

Lets Content Maker operators open Published contents and see every published page (and saved HTML drafts) instead of a list error, including when the store is SQLite with TEXT timestamps.

## ADDED Requirements

### Requirement: Published history lists without scan failure
Authenticated `GET /publishes/history` SHALL return HTTP 200 with `items` (array) and `total` (number). Each item MUST include `id`, `name`, `slug`. The handler MUST NOT return HTTP 500 with `failed to list published pages` solely because `created_at` / `updated_at` are stored as SQLite TEXT. An empty table MUST return `items: []` and `total: 0`.

#### Scenario: Existing published row appears in history
- **WHEN** at least one `published_pages` row exists and an authenticated client requests publish history
- **THEN** the response is 200
- **AND** `items` includes that row’s `id`, `name`, and `slug`

#### Scenario: Empty history is not an error
- **WHEN** there are no published pages
- **THEN** the response is 200 with an empty `items` array and `total` 0

### Requirement: Published contents UI shows the list
The Content Maker **Published contents** screen SHALL display published pages from that history API. It MUST NOT toast or banner `failed to list published pages` when the API returns 200. If the API still fails, the screen MAY show an error, but a successful list MUST show names (or “No published pages yet” when empty).

#### Scenario: Operator opens Published contents after a successful publish
- **WHEN** a page was published and the operator opens Published contents
- **THEN** that page’s name appears in the published list

### Requirement: Drafts list also survives TEXT timestamps
Authenticated `GET /publish-drafts` SHALL return HTTP 200 with `items` and `total` when `publish_drafts` timestamps are SQLite TEXT, not HTTP 500 `failed to list publish drafts` for that reason.

#### Scenario: Drafts list with TEXT timestamps
- **WHEN** `publish_drafts` has at least one row with TEXT `created_at`/`updated_at`
- **THEN** `GET /publish-drafts` returns 200 and includes that draft’s `id` and `name`
