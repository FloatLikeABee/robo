## Why

Both SurveyX (FormsX Survey Bot) and Projects (Morph Engi) force people to retype content that already exists in a document. SurveyX can only seed a survey draft from a `.md`/`.txt` paste or a keyword query, so a printed questionnaire, a scanned form, or a PDF brief has to be transcribed by hand before it can become a survey. Projects has no import path at all: a handover pack containing site logs, contacts, and cash-flow rows must be re-entered row by row through the create forms.

Both apps already have the ingredients — MorphAI clients, multipart upload handling, and (in ComposerX) a working PDF-text plus vision-description pipeline — they are just not wired into these two workflows.

## What Changes

**SurveyX (FormsX) — survey from a document or photo**

- New "Create from file" flow on the Survey Bot page accepting **one PDF or image** (JPEG/PNG/GIF/WebP) per request, alongside the existing `.md`/`.txt` paste.
- PDF files are text-extracted server-side; images are read by a vision model so photographed or scanned questionnaires work.
- The extracted content is sent to MorphAI, which returns a **Survey Bot markdown template** (front-matter, `# Instructions`, `## Q…` blocks) validated against the existing markdown parser before it reaches the user.
- The result lands in the existing draft editor as editable markdown — the user reviews, adjusts, and saves. Nothing is auto-published.
- Extracted source text and a per-file summary are returned so the user can see what the AI read.

**Projects (Morph Engi) — projects from description files**

- New "AI import" flow accepting **up to 3 files** per request, of type **PDF, TXT, CSV, or MD**.
- Real PDF text extraction is added to the Rust backend (today non-text uploads are stored with a "content not extracted" placeholder).
- MorphAI reads the combined content and proposes a **draft plan** of one or more projects, each with optional nested **site logs**, **people** (contractors/contacts), and **flow log entries** (income/expense), because a single handover pack can describe several jobs.
- The draft is returned for review and is fully editable; a separate confirm step writes the selected projects and their children through the existing create paths. **No records are created until the user confirms.**
- Import sessions are persisted so a draft can be reopened, and so a completed import records what was created.

**Shared**

- Per-file size and count limits, explicit unsupported-type errors, and a clear "AI not configured" response when `MORPH_AI_API_KEY` is absent.

## Capabilities

### New Capabilities

- `surveyx-file-import`: Upload a PDF or image to SurveyX, extract its content with AI, and produce a validated, editable Survey Bot markdown template.
- `projects-ai-import`: Upload up to three description files (PDF/TXT/CSV/MD) to Projects, have AI propose one or more projects with nested logs, people, and flow-log entries, then review and confirm creation.
- `ai-document-ingestion`: Shared rules for turning uploaded files into AI-readable text — supported types, per-file and per-request limits, PDF text extraction, image vision description, truncation, and error reporting.

### Modified Capabilities

_None — `openspec/specs/` contains no published specs yet, so all three capabilities above are new._

## Impact

**FormsX backend (Go, `formx/backend/`)**

- New handler for survey-from-file ingestion, registered under the existing protected `/api/v1/survey-bot` group.
- Needs PDF text extraction and a vision-capable chat call. Both exist in the repo but not in this module: `pkg/morphgraph.ExtractPDFFile` (wraps `ledongthuc/pdf`) and ComposerX's `describeImageForPublish` pattern. `pkg/morphai` is text-only, so multimodal support must be added there or a local OpenAI-compatible client used, as ComposerX does.
- `go.mod`: adds a PDF-extraction dependency (and an OpenAI-compatible client if that route is chosen).

**FormsX frontend (React, `formx/frontend/`)**

- `pages/SurveyBot.tsx`: file picker widened to PDF/images, new upload state, source preview.
- `lib/api.ts`: new multipart method next to `surveyBot.aiDraft`.

**Morph Engi backend (Rust, `morph-engi/backend/`)**

- New API module for import sessions: upload/analyze, get draft, confirm.
- `Cargo.toml`: adds a pure-Rust PDF text extraction crate.
- New tables for import sessions and their files, following the `verification_sessions`/`verification_files` pattern in `db/migrations.rs`.
- Reuses existing insert paths for `projects`, `site_logs`, `contractors`, and `flow_log_entries`.

**Morph Engi frontend (Svelte, `morph-engi/frontend/`)**

- `App.svelte`: new tab in the Projects module for upload, draft review, and confirm.
- `lib/api.ts`: multipart helper for multiple files (current `uploadFile` sends a single `file` field).

**Docs**

- `docs/agents/04-formx.md` and `morph-engi/README.md` gain the new endpoints; `docs/agents/10-shared-libraries.md` is updated if `pkg/morphai` gains multimodal support.

**Config**

- No new environment variables required. Both features depend on the existing `MORPH_AI_API_KEY` and degrade to a clear error when it is unset. A vision-capable model name may be read from an existing optional model variable.
