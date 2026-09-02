## 1. Event Logs nav and labels

- [x] 1.1 Rename Morph Utils module Survey Maker → Event Logs (`config.ts` label/shortLabel/description) and embed `/events-info`
- [x] 1.2 In SheetX Layout: title Event Logs; tabs Events & Info then Info Sheets; assistant copy Event Logs; default routes (`App.tsx` index/`*`) to `/events-info`
- [x] 1.3 Relabel Info Sheets UI (SurveyBot headings, browser `index.html` title) and landing Utils name; leave `/survey-bot` path and `sheetx` id

## 2. Info Sheet → Events & Info

- [x] 2.1 After a completed Info Sheet result is saved, if AI is configured, insert one Events & Info summary (dedupe by result id); collect still succeeds if AI is missing or fails
- [x] 2.2 Tests: summarize+insert on complete; no duplicate; collect still saves when AI is unset
- [x] 2.3 Optional retry control on an Info Sheets result that was not recorded

## 3. md / txt / json ingest

- [x] 3.1 Allow `.json` on Events & Info AI ingest (pretty-print JSON as source text); keep md/txt/pdf; update file picker `accept`
- [x] 3.2 Tests: json accepted into drafts with no persist; unsupported type 400; 503 when AI unset

## 4. Verify

- [x] 4.1 Update session-lessons product label (Event Logs / Info Sheets) and smoke Event Logs landing, Info Sheets publish/collect → event, and md/txt/json ingest review-then-save
