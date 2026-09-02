## 1. Nav

- [x] 1.1 In Data Access header nav, list Data reports before Data tables (Help last)

## 2. API

- [x] 2.1 Add authenticated `POST /api/v1/data-tables/:id/ai-analysis` that returns markdown for that table (capped sample; do not reuse file-import analyze)

## 3. Modal UI

- [x] 3.1 Shared AI analysis modal: dark overlay, rendered markdown, download `.md`, error if AI fails
- [x] 3.2 Wire AI analysis from table detail and from each Data tables list card (list Analyze must not only navigate)

## 4. Verify

- [x] 4.1 Header order is Data reports, Data tables, Help
- [x] 4.2 Open analysis on a table: modal markdown + download `.md`; failed AI shows an error (no fake file)
