## Why

In MorphUtils **Data Access**, the header lists Data tables before Data reports. Operators want reports first. Data tables also have no way to run an AI analysis of a table’s contents and keep the result as readable markdown they can download.

## What Changes

- Header section order: **Data reports**, then **Data tables**, then Help. Labels stay Data reports / Data tables. Default landing can stay Data tables (`/` → `/data-tables`).
- On Data tables (table detail, and a control from the list for that table), **AI analysis** opens a **modal**.
- The modal shows the analysis as **markdown** (rendered, not a raw dump).
- The user can **download** the analysis (markdown file). Dark-only UI; no theme switch; no under-title lede.

## Capabilities

### New Capabilities

- `data-access-nav-reports-first`: Data Access header lists Data reports before Data tables.
- `data-tables-ai-analysis-modal`: AI analysis of a data table in a modal, markdown display, download.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baselines)

## Impact

- Data Access frontend (`SharpReport/frontend`): `Sidebar.svelte` nav order; Data tables list + table detail; modal + markdown render + download.
- Data Access API (`SharpReport/backend`): new authenticated analyze-table endpoint (not the existing file-to-table `/data-tables/analyze`). Morph AI key already used elsewhere in this service.
- MorphUtils iframe is unchanged (`VITE_DATAX_URL`); SSO stays Morph AI cookie. No Event Logs / Content Maker / MorphNotes changes.
