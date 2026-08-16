## 1. Shared Go extraction module (`pkg/docextract`)

- [x] 1.1 Create `pkg/docextract/go.mod` as module `github.com/robo/docextract` (Go 1.21) with `github.com/ledongthuc/pdf` as its only direct dependency
- [x] 1.2 Implement `Classify(name, mime string) Kind` in `pkg/docextract/classify.go` returning `KindDocument`, `KindImage`, or `KindUnsupported` per the type table in `specs/ai-document-ingestion/spec.md`, falling back to the extension when the MIME type is generic
- [x] 1.3 Implement `ExtractPDF(path string) (string, error)` and `ExtractPDFBytes([]byte) (string, error)` in `pkg/docextract/pdf.go`, returning a distinct "no text extracted" error for PDFs with no text layer
- [x] 1.4 Implement `ExtractText(raw []byte) string` in `pkg/docextract/text.go` for TXT/MD/CSV, replacing invalid UTF-8 rather than failing, and preserving CSV newlines so header and data rows survive
- [x] 1.5 Implement `Truncate(s string, maxChars int) (string, bool)` appending an explicit truncation marker, and export the per-file and per-request caps as constants
- [x] 1.6 Write unit tests in `pkg/docextract` covering: each classification case, generic-MIME fallback, unsupported type, invalid UTF-8, CSV row preservation, and truncation marking
- [x] 1.7 Run `cd pkg/docextract && go test ./...` and confirm it passes

## 2. Multimodal support in `pkg/morphai`

- [x] 2.1 Add a multimodal message type (text + image data-URL parts) to `pkg/morphai/message.go` without altering the existing `Message` struct
- [x] 2.2 Add `ChatCompletionVision(ctx, parts, model)` to `pkg/morphai/client.go` posting an OpenAI-style `content` array to `{BaseURL}/chat/completions`, reusing the existing retry, rate-limit, and error-parsing paths
- [x] 2.3 Return an explicit configuration error from the vision method when `UseNativeAPI` is set, naming the setting that must change
- [x] 2.4 Add vision model resolution to `pkg/morphai/config.go` reading an existing optional model variable with a Qwen VL default, and no new required environment variable
- [x] 2.5 Add unit tests for message serialization shape, the native-endpoint error, and vision model resolution
- [x] 2.6 Run `cd pkg/morphai && go test ./...`, then `cd morph && go build ./...` and `cd composerx/backend && go build ./...` to confirm existing consumers still compile

## 3. FormsX backend — survey from file

- [x] 3.1 Add `replace github.com/robo/docextract => ../../pkg/docextract` and the require entry to `formx/backend/go.mod`, then run `go mod tidy`
- [x] 3.2 Create `formx/backend/internal/handler/survey_bot_from_file.go` with a multipart handler accepting one `file` part plus optional `title_hint` and `instructions` text fields
- [x] 3.3 Enforce upload limits in the handler: at most one file, per-file byte cap, and an unsupported-type error naming PDF and the accepted image types
- [x] 3.4 Extract content by kind — `docextract.ExtractPDF` for PDFs, `morphai.ChatCompletionVision` for images — and return HTTP 503 with an "AI not configured" message when the client is unconfigured
- [x] 3.5 Build the generation prompt instructing the model to emit only Survey Bot markdown (front matter with `slug`/`title`/`tags`, `# Instructions`, `## Q<n> — <label>` blocks with `field`/`collect`/`required`/`prompt`), selector `collect` with `options` for fixed-choice questions and free text otherwise
- [x] 3.6 Validate the result with `surveybot.ParseMarkdown`, retry once with the parser error fed back into the prompt, and on a second failure return the validation message together with the rejected markdown
- [x] 3.7 Normalize the generated `slug` to lowercase alphanumerics and hyphens using the existing `normalizeSlug`, deriving the title from `title_hint` when supplied
- [x] 3.8 Return the extracted text or image description in the response alongside the markdown, and create no template in the request path
- [x] 3.9 Register `POST /survey-bot/templates/from-file` in the protected survey-bot group in `formx/backend/internal/handler/handler.go`
- [x] 3.10 Add handler tests covering: two-file rejection, unsupported type, no-text PDF, validation-failure response shape, and slug normalization
- [x] 3.11 Run `cd formx/backend && go build ./... && go test ./...`

## 4. FormsX frontend — Survey Bot page

- [x] 4.1 Add a multipart `surveyBot.fromFile` method to `formx/frontend/src/lib/api.ts` next to `aiDraft`, following the `uploadPromptMedia` pattern with no explicit `Content-Type`
- [x] 4.2 Widen the file input `accept` in `formx/frontend/src/pages/SurveyBot.tsx` to include PDF and image types alongside `.md`/`.txt`
- [x] 4.3 Route selected files by type in `onUploadFile`: `.md`/`.txt` keep the existing client-side read, PDF and images call `surveyBot.fromFile`
- [x] 4.4 Add a busy state that disables the control and shows a busy label while generation is in flight, blocking concurrent submissions
- [x] 4.5 Prompt for confirmation before replacing a non-empty unsaved draft, then populate the markdown editor and title field on success
- [x] 4.6 Display the returned error message on failure while leaving the current draft unchanged
- [x] 4.7 Show the extracted source text or image description in a collapsible panel so the user can see what the AI read
- [x] 4.8 Run `cd formx/frontend && npm run lint && npm run build`

## 5. Morph Engi backend — PDF extraction and prompt budget

- [x] 5.1 Add the `pdf-extract` crate to `morph-engi/backend/Cargo.toml`
- [x] 5.2 Create `morph-engi/backend/src/services/doc_text.rs` exposing async PDF extraction via `tokio::task::spawn_blocking` wrapped in `std::panic::catch_unwind`, returning a per-file error on failure or panic
- [x] 5.3 Add classification and text extraction for TXT/MD/CSV in the same module, replacing invalid UTF-8 and preserving CSV rows, and rejecting images and other types as unsupported for this feature
- [x] 5.4 Add per-file and combined truncation helpers with explicit truncation markers, sized so the assembled system plus user prompt stays under `morphai::MAX_CHAT_REQUEST_CHARS`
- [x] 5.5 Replace the placeholder branch in `api/verification.rs`'s `extract_text_excerpt` with a call into the new module so uploaded PDFs there also yield real text
- [x] 5.6 Add unit tests for classification, unsupported-type rejection, CSV row preservation, truncation marking, and panic isolation
- [x] 5.7 Run `cd morph-engi/backend && cargo build && cargo test`

## 6. Morph Engi backend — import sessions

- [x] 6.1 Append idempotent `CREATE TABLE IF NOT EXISTS project_import_sessions` and `project_import_files` statements to the `MIGRATIONS` array in `morph-engi/backend/src/db/migrations.rs`, modelled on the verification tables, with columns for org scope, status, instruction, draft JSON, created project ids, and timestamps
- [x] 6.2 Create `morph-engi/backend/src/api/project_import.rs` and register it in `api/mod.rs`
- [x] 6.3 Implement `POST /project-imports/analyze` accepting multipart with up to three `files` parts plus an optional instruction, rejecting a fourth file and unsupported types before any AI call, and storing uploads under `uploads/<org_id>/`
- [x] 6.4 Insert the session with status `analyzing` and one `project_import_files` row per upload holding its original name, mime, size, excerpt, and per-file error
- [x] 6.5 Fail the request with HTTP 503 and an "AI not configured" message when `state.ai` is absent, and fail with a "no content could be read" error when no file yielded usable content
- [x] 6.6 Build the analysis prompt demanding a single JSON object: a list of proposed projects, each with project fields plus nested `site_logs`, `people`, and `flow_log_entries`, one project per distinct job, omitting unknown optional fields rather than inventing values
- [x] 6.7 Parse the reply with `morphai::extract_json_object`, store the draft with status `drafted`, and on AI failure store status `failed` with the message and return an error
- [x] 6.8 Implement `GET /project-imports` and `GET /project-imports/:id`, both filtered by `organization_id` so another org's session reads as not found, returning the draft plan and file metadata
- [x] 6.9 Return per-file results in the analyze response identifying each file by original name with either its excerpt or its error

## 7. Morph Engi backend — confirm and record creation

- [x] 7.1 Factor the validation currently inline in `create_project`, `create_site_log`, `create_contractor`, and `flow_log::create_entry` into shared helpers (name required, code generation from name, positive amount, income/expense direction normalization) and call them from the existing handlers
- [x] 7.2 Implement `POST /project-imports/:id/confirm` accepting an edited draft naming which proposed projects and which nested records to create
- [x] 7.3 Reject confirm on a session already in status `completed` with an "already completed" error so no duplicates are created
- [x] 7.4 Create each accepted project and its included children in one transaction per project via the shared helpers, using the submitted edited values
- [x] 7.5 Link nested records to the newly created project id: `site_logs.project_id`, `project_contractors`, and flow-log entries associated per the draft
- [x] 7.6 Report per-record errors for proposals that fail validation while still creating the valid siblings, and return a summary of what was created and what failed
- [x] 7.7 Set session status to `completed` and record the created project ids
- [x] 7.8 Add tests covering: fourth-file rejection, unsupported type, subset confirm, edited values used, excluded child skipped, blank name rejected, generated code, zero-amount entry reported, double confirm rejected, and cross-org access denied
- [x] 7.9 Run `cd morph-engi/backend && cargo build && cargo test`

## 8. Morph Engi frontend — AI import view

- [x] 8.1 Add a multi-file upload helper to `morph-engi/frontend/src/lib/api.ts` appending repeated `files` parts, leaving the existing single-part `uploadFile` unchanged
- [x] 8.2 Add an `import` tab to the Projects `ModuleShell` tabs in `morph-engi/frontend/src/App.svelte`
- [x] 8.3 Build the upload panel: file selection capped at three with an explanatory message on the fourth, an optional instruction field, and a submit control
- [x] 8.4 Add a busy state that disables submit and shows a busy label during analysis
- [x] 8.5 List each uploaded file with its read status and display per-file errors while still showing the draft from the remaining files
- [x] 8.6 Render the draft as editable proposed projects with their nested site logs, people, and flow-log entries, each project and nested record editable or excludable
- [x] 8.7 Wire the confirm action to the confirm endpoint and refresh the projects list on success without a manual reload
- [x] 8.8 Display the backend's "AI not configured" message instead of an empty draft when AI is unavailable
- [x] 8.9 Run `cd morph-engi/frontend && npm run build`

## 9. Docs and verification

- [x] 9.1 Add the new survey-bot endpoint and the vision/PDF dependency notes to `docs/agents/04-formx.md`
- [x] 9.2 Document `pkg/docextract` and the `pkg/morphai` vision method in `docs/agents/10-shared-libraries.md`
- [x] 9.3 Add the four `project-imports` endpoints and the three-file/PDF-TXT-CSV-MD limits to `morph-engi/README.md`
- [x] 9.4 Manually verify the SurveyX flow end to end: start formx api and ui via `./start-all.sh`, upload a text-layer PDF and a photographed form, confirm a valid editable draft appears and nothing is saved until the existing save action runs
- [x] 9.5 Manually verify the Projects flow end to end: start morph-engi api and ui, upload a PDF brief plus a CSV of costs plus a MD contact list, confirm multiple proposed projects with nested records, edit one, exclude one, confirm, and check the new projects and their children appear
- [x] 9.6 Verify the unconfigured-AI path on both apps by unsetting `MORPH_AI_API_KEY` and confirming each returns the 503 "AI not configured" message
- [x] 9.7 Run `openspec validate ai-file-import-surveys-and-projects --strict` and confirm it passes
