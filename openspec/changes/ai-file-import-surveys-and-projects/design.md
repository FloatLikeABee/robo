## Context

See `proposal.md` — Why. This design covers two apps in different languages that need the same shape of pipeline (upload → extract → AI → reviewable draft), plus the shared plumbing each is missing.

Current state that shapes the approach:

- **FormsX (Go)** already has `POST /api/v1/survey-bot/templates/ai-draft`, but it ignores AI entirely: `draftSurveyMarkdown` returns a hardcoded three-question template and `webresearch.Gather` output is discarded (`_ = notes`). The Survey Bot markdown parser (`internal/surveybot.ParseMarkdown`) is the validation gate and already returns useful errors.
- **FormsX** has no PDF or vision capability. `pkg/morphai` is text-only: `Message.Content` is a `string`, so there is no way to attach an image.
- **ComposerX** already solves exactly this problem in Go — `composerx/backend/publish_source.go` extracts PDF text, describes images through an OpenAI-compatible vision call, summarizes each file, and concatenates the results. It is the working reference but lives in `package main` and cannot be imported.
- **`pkg/morphgraph`** has `ExtractPDFFile`/`ExtractPDFBytes` wrapping `ledongthuc/pdf`, but the module also depends on the Neo4j driver.
- **Morph Engi (Rust)** already has the two-step, session-persisted, multipart AI feature pattern in `api/verification.rs` with `verification_sessions` / `verification_files`. Its `extract_text_excerpt` handles text extensions only — a PDF becomes `"[Binary or non-text file — N bytes …]"`.
- **`pkg/morphai-rs`** caps a whole request at `MAX_CHAT_REQUEST_CHARS = 12_000` and enforces it in `trim_messages_for_api`, which drops messages from index 1 upward and then truncates message 0. With only a system and a user message, oversized input silently eats the system prompt.
- Both apps degrade on a missing `MORPH_AI_API_KEY` rather than failing at boot, and both are organization-scoped through existing auth middleware.

## Goals / Non-Goals

**Goals:**

- One extraction contract shared by both apps, so "what counts as readable" is defined once (see `specs/ai-document-ingestion/spec.md`).
- Reuse the existing validation gates rather than inventing new ones: the Survey Bot markdown parser for SurveyX, the existing insert paths' validation for Projects.
- Keep AI output out of the database until a human has seen it.
- Add no new environment variables and no new infrastructure.

**Non-Goals:**

- OCR of scanned PDFs. A PDF without a text layer is an error, not an image-pipeline input. (Images uploaded *as images* do go through vision.)
- Images for the Projects import. The proposal scopes it to PDF/TXT/CSV/MD.
- Streaming or background jobs. Both endpoints are synchronous request/response, like `run_verification` today.
- Refactoring ComposerX to use the new shared extraction package. Worth doing later; out of scope here.
- Changing the existing `ai-draft` endpoint's behavior or the `.md`/`.txt` paste path.

## Decisions

### 1. New Go module `pkg/docextract` rather than reusing `pkg/morphgraph`

`morphgraph.ExtractPDFFile` does what FormsX needs, but importing `github.com/robo/morphgraph` drags in `neo4j-go-driver` and the embedding client for one function.

**Decision:** add `pkg/docextract` — a zero-Neo4j Go module whose only dependency is `ledongthuc/pdf`. It exposes classification (`Classify(name, mime) Kind`), `ExtractPDF`, `ExtractText`, and the truncation helper, mirroring the rules in the ingestion spec.

**Alternatives considered:**

- *Import `pkg/morphgraph`* — rejected: unrelated heavy dependency tree in FormsX.
- *Copy ComposerX's helpers into FormsX* — rejected: a third copy of the same logic (ComposerX, Academi, morphgraph already each have one).
- *Extract from ComposerX in this change* — rejected: widens the blast radius to a third app. `pkg/docextract` is written so ComposerX can adopt it later.

### 2. Extend `pkg/morphai` with a multimodal call instead of adding `go-openai` to FormsX

FormsX needs one vision call. ComposerX gets this from `sashabaranov/go-openai`.

**Decision:** add to `pkg/morphai` an additive vision entry point — a `ContentPart`/multimodal message type and a method that posts an OpenAI-style `content` array to the existing `{BaseURL}/chat/completions` path. The existing `Message` struct and all current methods are untouched, so morph, composerx, and booki are unaffected.

Because the multimodal payload shape only exists on the OpenAI-compatible path, the method SHALL return a clear error when the client is configured for the native DashScope `text-generation/generation` endpoint (`UseNativeAPI`). The vision model name comes from an existing optional model variable, defaulting to a Qwen VL model, and falls back to an error rather than silently using a text-only model.

**Alternatives considered:**

- *Add `sashabaranov/go-openai` to FormsX* — rejected: a large dependency for one call, and it leaves `pkg/morphai` unable to do vision for the next caller.
- *Make `Message.Content` an `any`* — rejected: **breaking** for four apps.
- *Proxy the vision call through bk's Python image reader* — rejected: makes FormsX depend on a service it currently has no relationship with.

### 3. `pdf-extract` crate for Morph Engi, run on a blocking thread with panic isolation

**Decision:** add the pure-Rust `pdf-extract` crate and call it from `tokio::task::spawn_blocking`, wrapped in `std::panic::catch_unwind`. Extraction failure or panic becomes a per-file error, matching the ingestion spec's "one bad file must not abort the request".

**Rationale:** parsing is CPU-bound and must not block the async runtime, and PDF parsers panic on malformed input. `verification.rs`'s existing behavior (placeholder text) is what we are replacing, so a regression here is visible.

**Alternatives considered:**

- *Shell out to `pdftotext`* — rejected: adds an undeclared system dependency to a `cargo run` app.
- *Send raw PDF bytes to the model* — rejected: the configured chat API takes text (and images), not documents.
- *Extract in the frontend with pdf.js* — rejected: puts a trust boundary and a bundle-size cost in the wrong place; the spec says extraction is server-side.

### 4. Keep the AI request inside the shared character budget instead of raising it

`pkg/morphai-rs` trims oversized requests by deleting messages, which would silently discard the system prompt that defines the JSON contract.

**Decision:** budget the prompt explicitly on the caller side. Cap extracted text per file and cap the combined document section so the assembled system + user prompt stays comfortably under `MAX_CHAT_REQUEST_CHARS`, with an explicit truncation marker per the ingestion spec. Do not change the shared constant.

**Rationale:** the truncation is then visible and attributable instead of being an invisible prompt-corrupting side effect. Raising the cap would move the cliff without removing it, and it affects every other `morphai-rs` consumer (UsersPanel, SharpReport).

**Follow-on:** if a later change needs bigger prompts, the right fix is a per-request budget parameter on the shared client, not a bigger global constant.

### 5. Projects import is two calls with a persisted session; SurveyX is one call with no persistence

The two apps differ deliberately.

**Projects** writes to four tables across parents and children, so `analyze` and `confirm` are separate endpoints and the draft is persisted in new `project_import_sessions` / `project_import_files` tables, modelled directly on `verification_sessions` / `verification_files` (same migration style: idempotent `CREATE TABLE IF NOT EXISTS` in the `MIGRATIONS` array, org-scoped, uploads under `uploads/<org_id>/`). Persisting the draft is what makes "reopen a draft" and "already completed" detection possible.

**SurveyX** produces one markdown string that the existing editor already knows how to hold and save. Adding a Mongo collection for it would duplicate the template collection. So generation returns the draft in the response and the user saves through the existing `POST /survey-bot/templates`.

**Alternatives considered:**

- *One-shot create for Projects* — rejected: the spec requires review before any record exists, and a bad AI reading of a handover pack would otherwise scatter records across four tables.
- *Persist SurveyX drafts too* — rejected: no requirement needs it, and unsaved-draft state already lives in the page.

### 6. Confirm writes through in-process SQL, not self-directed HTTP calls

**Decision:** `confirm` performs the same inserts the existing handlers do, in one transaction per project so a project and its children are all-or-nothing, and reports per-record errors for the rest.

**Rationale:** re-entering the app over HTTP would re-do auth, lose transactionality, and multiply failure modes. The trade-off is duplicated validation logic; mitigated by factoring the validation each handler does (name required, code generation, positive amount, direction normalization) into helpers both the handler and the import path call.

### 7. Strict JSON contract, parsed with the existing lenient extractors

**Decision:** both AI calls demand a single JSON object with no markdown fences and are parsed with the existing helpers — `morphai.ExtractJSONObject` (Go) and `extract_json_object` (Rust) — which already tolerate a fenced or chatty response. Unknown fields are ignored; missing optional fields are omitted rather than defaulted to invented values.

For SurveyX the model is asked for markdown, not JSON, because the Survey Bot parser is the contract. Generation validates with `surveybot.ParseMarkdown` and retries once, feeding the parser's error back into the prompt; a second failure returns the error plus the rejected markdown so the user can fix it by hand. This mirrors the existing `ai-draft` handler, which already returns `{"error": ..., "markdown": md}` on validation failure.

### 8. Endpoint placement

- FormsX: `POST /api/v1/survey-bot/templates/from-file` — inside the existing protected `survey-bot` group, next to `ai-draft`, so it inherits the same auth.
- Morph Engi: `POST /api/v1/project-imports/analyze`, `POST /api/v1/project-imports/:id/confirm`, `GET /api/v1/project-imports`, `GET /api/v1/project-imports/:id` — inside the auth-layered `api` router in `api/mod.rs`.

### 9. Frontend upload helpers

FormsX's `lib/api.ts` gets a multipart method beside `surveyBot.aiDraft`, following the existing `uploadPromptMedia` pattern (no `Content-Type` header; let the browser set the boundary).

Morph Engi's `uploadFile` sends exactly one part named `file`, so a multi-file helper is added that appends repeated `files` parts — the same field name `run_verification` already accepts.

## Risks / Trade-offs

- **AI misreads a handover pack and proposes wrong projects** → nothing is written until confirm; the draft is fully editable; per-file extracted excerpts are shown so the user can see what the model read.
- **`pdf-extract` panics or mangles a real-world PDF** → `catch_unwind` inside `spawn_blocking` turns it into a per-file error; other files in the request still produce a draft.
- **Vision path unavailable because the deployment uses the native DashScope endpoint** → the multimodal method returns an explicit configuration error, and PDF upload still works; the error names the setting to change.
- **Prompt budget exceeded by three large files** → per-file and combined caps with visible truncation markers keep the request under the shared limit rather than relying on the client's message-dropping trim.
- **Duplicated validation between the confirm path and existing create handlers drifts** → shared validation helpers, and the spec's scenarios (blank name, generated code, zero amount) are the regression tests.
- **Double-confirm creating duplicate records** → session status guard rejects confirm on a `completed` session.
- **New Go module adds build wiring** (`replace` directive, `go.work` if present) → mechanical and covered in tasks; `pkg/docextract` deliberately has one dependency to keep this cheap.
- **Two apps changed in one change set** → the two halves share only `pkg/docextract` and the ingestion spec; they can be implemented, tested, and merged independently in the task order below.

## Migration Plan

1. Ship `pkg/docextract` and the `pkg/morphai` multimodal addition first — both are additive, so existing consumers keep building.
2. Morph Engi's new tables are created by appending idempotent statements to the `MIGRATIONS` array in `db/migrations.rs`, which runs on every boot and ignores per-statement errors. No backfill, no destructive change.
3. Each half is independently deployable and independently revertable: the FormsX feature is one new route plus UI, the Morph Engi feature is four new routes plus two tables plus UI.
4. Rollback: remove the new routes and hide the UI entries. The new tables can be left in place; they are unreferenced when the routes are gone.
5. Docs (`docs/agents/04-formx.md`, `docs/agents/10-shared-libraries.md`, `morph-engi/README.md`) are updated in the same change so the endpoint tables stay accurate.

## Open Questions

- Which Qwen VL model name to default to for the FormsX vision call. Deferrable: the spec only requires that a vision-capable model is used and that a missing or unusable configuration produces a clear error. The default can be adjusted without touching specs, approach, or tasks.
- Whether the CSV excerpt should be row-sampled (head plus tail) rather than head-truncated for large files. Deferrable: the ingestion spec only requires that rows are preserved and that truncation is marked.
