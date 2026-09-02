## Why

Morph Utils still presents this module as Survey Maker / AI Surveys, which hides that the real job is an operational event log plus published sheets that collect info. Collected sheet answers stay in a results list and never become Events & Info records, and Events & Info ingest still rejects JSON even though operators already have notes as `.md` / `.txt` / `.json`.

## What Changes

- **BREAKING (labels):** Morph Utils left-nav module **Survey Maker** becomes **Event Logs**. The embedded app title, browser title, landing card, and assistant chrome use **Event Logs** (not Survey Maker / SurveyX / SurveysX).
- **BREAKING (labels):** Inner tab **AI Surveys** / Survey Maker sheet builder becomes **Info Sheets**. Publish-and-collect behavior stays: design a sheet, publish a link, respondents submit, answers appear in Info Sheets.
- **Nav order:** Header tabs are **Events & Info** first, then **Info Sheets**. Opening Event Logs lands on Events & Info (not Info Sheets).
- **Info Sheets → Events & Info:** When a published Info Sheet collects a completed response, AI MUST summarize that response into an Events & Info record. The sheet result is still saved. If AI is unavailable, the collect still succeeds and the operator is told the event was not recorded.
- **Events & Info file ingest:** The user MUST be able to upload **`.md`**, **`.txt`**, or **`.json`** so AI can turn the file into Events & Info draft data (review then save, same as today’s ingest). JSON is newly allowed; existing PDF/URL/paste ingest MAY remain.

Internal ids (`sheetx`, `/survey-bot`) MAY stay for URLs. User-facing copy MUST use the new names.

## Capabilities

### New Capabilities

- `event-logs-nav`: Event Logs module naming, Info Sheets tab label, Events & Info before Info Sheets, default landing on Events & Info.
- `info-sheet-event-record`: Completed Info Sheet collections are AI-summarized into Events & Info records without dropping the original answers.
- `events-info-md-txt-json-ingest`: Events & Info accepts `.md`, `.txt`, and `.json` uploads; AI extracts draft event records for review before save.

### Modified Capabilities

- (none — `openspec/specs/` has no archived baselines for this product)

## Impact

- **Morph Utils shell:** `morph-utils/frontend/src/config.ts` (label, shortLabel, description, default embed path `/events-info`).
- **SheetX / FormsX:** `formx/frontend` Layout tabs/title, SurveyBot copy, EventsInfo file picker, `index.html`; default route `/events-info`.
- **FormsX backend:** Survey-bot result finalize + public AI sheet complete → AI Events & Info insert; `events_info_ingest.go` allow `.json`; assistant/welcome strings.
- **Landing:** `landing/index.html` Utils list name.
- **Docs / session lessons:** Product label catalog currently says Survey Maker — update when this ships.
