## Context

See proposal.md for motivation.

Morph Utils left nav still labels this iframe **Survey Maker** (`config.ts`) and embeds `${sheetxUrl}/survey-bot`. SheetX `Layout` title is Survey Maker; header tabs are **AI Surveys** then **Events & Info**; index and `*` redirect to `/survey-bot`. Info Sheets (Survey Bot) already publish public `/s/:slug` links and store `SurveyBotResult` answers. Completing a sheet (`finalizeSurvey` / public AI sheet) does not write Events & Info.

Events & Info already has `POST /api/v1/events-info/ai-ingest` for txt/md/pdf + URL + paste, returning drafts for review. `.json` is rejected. Keep the outer Morph Utils left nav; inner sections stay header tabs.

## Goals / Non-Goals

**Goals:**

- User-facing Event Logs / Info Sheets / Events-first landing.
- Auto-record one Events & Info row per completed Info Sheet when AI works.
- Allow `.json` (plus existing md/txt) on Events & Info ingest; review-then-save.

**Non-Goals:**

- Renaming URL ids (`sheetx`, `/survey-bot`) or env vars (`VITE_SHEETX_URL`).
- Removing PDF/URL/paste ingest.
- Changing Content Maker, Data Access, or Project.
- Replacing the Morph Utils left nav with SheetX’s inner tabs.

## Decisions

### 1. Labels only; keep module id `sheetx`
- **Choice**: `UTILS_MODULES` `label` → Event Logs, `shortLabel` → Logs (or Events). Description mentions event log + info sheets. Embed URL → `/events-info`. Formx document title, Layout header, assistant title/welcome, landing Utils name, SurveyBot h1/tab → Info Sheets.
- **Rationale**: Session lesson said Survey Maker; this change explicitly replaces it. Internal ids stay so iframe/env wiring does not churn.
- **Alternatives**: Rename module id to `event-logs` — extra redirects, not asked.

### 2. Tab order and default route
- **Choice**: Layout nav: Events & Info, then Info Sheets. `App.tsx` index and unknown routes → `/events-info`. Morph Utils embed path `/events-info`. Keep `/survey-bot` as the Info Sheets route.
- **Rationale**: User asked Events & Info first and to land there.
- **Alternatives**: Keep default on Info Sheets — contradicts “goes before”.

### 3. Sheet collect auto-writes Events & Info
- **Choice**: After a `SurveyBotResult` is inserted (assistant finalize and public AI sheet complete), if Morph AI is configured, summarize answers → `title` + markdown `detail` and insert one Events & Info document. Store `source: info_sheet` and `source_result_id` on the event (or equivalent detail footer) so the same result is not inserted twice. If AI is missing or fails, still return success on collect; log/surface “event not recorded” on the result (flag or admin hint). Optional “Record in Events & Info” on an existing result for retry — include if cheap.
- **Rationale**: Respondent already submitted; they should not confirm an internal log. Operator can delete a bad event. Distinct from file ingest, which is operator-driven and stays review-then-save.
- **Alternatives**: Queue a draft for the operator — extra step the user did not ask for. Block collect on AI failure — drops field data.

### 4. JSON on the existing ingest endpoint
- **Choice**: Extend `isAllowedEventIngestFile` with `.json` / `application/json`. Read bytes as UTF-8; if JSON, compact-pretty print as source text; then the same AI draft pipeline. File picker `accept` includes `.json`. Keep `.pdf` and URL/paste.
- **Rationale**: User named md/txt/json. Reuse ingest rather than a second upload API. Auto-saving a file would skip review and duplicate the sheet-collect path.
- **Alternatives**: New “quick add file” that auto-saves one record — skips the existing draft UI and is harder to undo.

### 5. Copy sweep, not route sweep
- **Choice**: Replace user-visible Survey Maker / AI Surveys strings in Morph Utils, formx Layout/SurveyBot/index.html, assistant suggestions, landing. Do not rename Go types `SurveyBot*` in this change.
- **Rationale**: Behavior specs care about labels; a backend rename is a separate refactor.
- **Alternatives**: Full SurveyBot → InfoSheet rename now — large diff, easy to miss.

## Risks / Trade-offs

- [AI summary invents facts] → Prompt: only use answers; operator can delete the event.
- [Slow collect if AI is inline] → Run summarize after result insert; do not fail the HTTP collect on AI timeout; bound the AI call.
- [Duplicate events] → Dedupe on `source_result_id`.
- [JSON arrays of many items] → Existing ingest cap (25 drafts) still applies.

## Migration Plan

- Deploy Morph Utils frontend + formx frontend/backend together so embed path and APIs match.
- No DB migration required if extra source fields live on the Events & Info document JSON.
- Rollback: revert labels and skip the summarize hook; ingest JSON allow-list is backward compatible.

## Open Questions

None that change specs or tasks. Retry button on old results can ship in the same Info Sheets results row if it stays a small extra control.
