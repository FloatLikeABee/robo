## Purpose

Lets a SurveyX user turn an existing document into a survey: upload a PDF or a photo of a questionnaire, and AI produces a validated, editable Survey Bot markdown template instead of the user retyping every question.

## ADDED Requirements

### Requirement: Create a survey draft from an uploaded file

The system SHALL provide an authenticated endpoint that accepts a single uploaded file and returns a Survey Bot markdown template derived from that file's content.

The endpoint SHALL accept PDF and image files (JPEG, PNG, GIF, WebP). It SHALL accept at most one file per request. It SHALL accept optional `title_hint` and `instructions` text fields that steer the generated survey.

The endpoint SHALL be registered alongside the existing Survey Bot template endpoints and SHALL require the same authentication as the existing AI draft endpoint.

#### Scenario: Survey generated from a PDF questionnaire

- **WHEN** an authenticated user uploads a PDF containing a questionnaire
- **THEN** the response contains Survey Bot markdown with front matter, an `# Instructions` section, and one `## Q…` block per question found in the document

#### Scenario: Survey generated from a photographed form

- **WHEN** an authenticated user uploads a JPEG photo of a printed form
- **THEN** the visible questions and answer options are read from the image
- **AND** the response contains Survey Bot markdown reflecting them

#### Scenario: Title hint is honoured

- **WHEN** the request includes a `title_hint` of `Customer onboarding`
- **THEN** the returned markdown's `title` front-matter field is based on that hint rather than being derived from the filename

#### Scenario: More than one file rejected

- **WHEN** a request contains two files
- **THEN** the request fails with a message stating only one file is accepted per request

#### Scenario: Unauthenticated request rejected

- **WHEN** the request carries no valid session
- **THEN** the request fails with HTTP 401 and no AI request is made

### Requirement: Generated markdown is valid before it is returned

The system SHALL validate generated markdown against the Survey Bot template parser before returning success. Markdown that fails validation SHALL NOT be returned as a successful draft.

The system MAY retry generation once when the first attempt fails validation. If validation still fails, the system SHALL return an error that includes the validation message and the rejected markdown so the user can repair it manually.

#### Scenario: Valid markdown returned as a draft

- **WHEN** the AI returns markdown that the Survey Bot parser accepts
- **THEN** the response reports success and includes the markdown

#### Scenario: Invalid markdown reported with detail

- **WHEN** the AI returns markdown that the parser rejects on every attempt
- **THEN** the response is an error containing the parser's validation message and the rejected markdown
- **AND** no template is saved

### Requirement: Generated survey structure

Generated markdown SHALL follow the existing Survey Bot template format: YAML front matter containing `slug`, `title`, and `tags`; an `# Instructions` section; and one `## Q<n> — <label>` block per question with `field`, `collect`, `required`, and `prompt` attributes.

Questions whose source content offers a fixed set of answers SHALL use a selector-style collect mode with an `options` list. Questions with open-ended answers SHALL use a free-text collect mode.

The generated `slug` SHALL be URL-safe and derived from the title.

#### Scenario: Multiple-choice question becomes a selector

- **WHEN** the source document lists a question with fixed choices such as Yes / No / Not sure
- **THEN** the corresponding `## Q…` block uses a selector collect mode and lists those choices as `options`

#### Scenario: Open-ended question becomes free text

- **WHEN** the source document contains a question with a blank line for a written answer
- **THEN** the corresponding `## Q…` block uses a free-text collect mode with no `options`

#### Scenario: Slug is URL-safe

- **WHEN** the generated title contains spaces, punctuation, or non-ASCII characters
- **THEN** the `slug` front-matter value contains only lowercase alphanumerics and hyphens

### Requirement: Draft is reviewable and never auto-saved

The system SHALL return the generated markdown as an editable draft. It SHALL NOT create, update, compile, or publish a Survey Bot template as part of the generation request.

The response SHALL include the text that was read from the uploaded file so the user can confirm the AI worked from the right content.

#### Scenario: Nothing is persisted by generation

- **WHEN** a generation request succeeds
- **THEN** no new Survey Bot template exists in storage
- **AND** the user must invoke the existing save action to persist the draft

#### Scenario: Source content shown to the user

- **WHEN** a generation request succeeds
- **THEN** the response includes the extracted text or image description used as input

### Requirement: Survey Bot page accepts documents and images

The Survey Bot page SHALL let the user pick a PDF or image file and generate a draft from it, in addition to the existing `.md`/`.txt` paste and keyword-query paths.

While generation is in progress the page SHALL show a busy state and SHALL prevent a second concurrent submission. On success the generated markdown SHALL populate the existing draft editor and title field, replacing any unsaved draft only after the user confirms. On failure the page SHALL display the returned error message.

#### Scenario: File picker offers the new types

- **WHEN** the user opens the file picker on the Survey Bot page
- **THEN** PDF and image files are selectable alongside `.md` and `.txt`

#### Scenario: Generated draft fills the editor

- **WHEN** generation from an uploaded PDF succeeds
- **THEN** the markdown editor contains the generated template and the title field is populated
- **AND** the user can edit the markdown before saving

#### Scenario: Existing draft protected

- **WHEN** the editor already contains unsaved markdown and the user generates from a file
- **THEN** the user is asked to confirm before the existing draft is replaced

#### Scenario: Busy state during generation

- **WHEN** a generation request is in flight
- **THEN** the generate control is disabled and shows a busy label

#### Scenario: Error surfaced to the user

- **WHEN** the request fails because the PDF contained no extractable text
- **THEN** the page displays that error message and leaves the current draft unchanged

### Requirement: Existing survey draft paths keep working

The keyword-based AI draft endpoint and the `.md`/`.txt` client-side paste path SHALL continue to behave as they do today.

#### Scenario: Keyword draft unchanged

- **WHEN** a user requests an AI draft from a keyword query with no file
- **THEN** the existing endpoint responds as before

#### Scenario: Markdown paste unchanged

- **WHEN** a user selects a `.md` file on the Survey Bot page
- **THEN** its text is loaded directly into the editor without an AI request
