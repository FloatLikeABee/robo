## Purpose

Morph Data Timelines replace Stories so users can turn file, URL, or pasted source material into a saved, previewable, and publicly publishable chronological markdown timeline.

## ADDED Requirements

### Requirement: Timelines replace Stories in Morph Data navigation
The system SHALL present a Morph Data module labeled **Timelines** (not Stories) and SHALL route users to a Timelines workspace. Requests to legacy Stories paths (`/stories`, `/story-board`) SHALL redirect to the Timelines path.

#### Scenario: Nav shows Timelines
- **WHEN** an authenticated user opens Morph Data navigation
- **THEN** the module is labeled Timelines and opens the Timelines list workspace

#### Scenario: Legacy Stories path redirects
- **WHEN** a user navigates to `/stories` or `/story-board`
- **THEN** they are redirected to the Timelines route

### Requirement: Create Timeline requires at least one source
The system SHALL allow creating a timeline from any combination of: (1) file upload, (2) URL, and (3) pasted content. The create flow SHALL accept all three channels in one form. If none of the three sources is provided, the system MUST reject the create attempt with a clear warning or validation error and MUST NOT call AI generation or persist a timeline.

#### Scenario: Paste-only create succeeds
- **WHEN** the user pastes non-empty content and submits create with no file and no URL
- **THEN** the system accepts the request and proceeds with timeline generation

#### Scenario: File-only create succeeds
- **WHEN** the user uploads a supported file within size limits and submits create with empty paste and no URL
- **THEN** the system accepts the request and proceeds with timeline generation

#### Scenario: URL-only create succeeds
- **WHEN** the user provides a non-empty URL and submits create with no file and empty paste
- **THEN** the system accepts the request and proceeds with timeline generation

#### Scenario: No source shows warning
- **WHEN** the user submits create with no file, no URL, and empty paste
- **THEN** the UI shows a warning and the API returns a validation error without creating a timeline

### Requirement: File upload type and size limits
When a file is used as a source, the system SHALL accept only `.txt`, `.pdf`, and `.md` files. The system SHALL reject files larger than **5 MB** (5,242,880 bytes) with a clear error. Unsupported types MUST be rejected before AI generation.

#### Scenario: Allowed PDF under limit
- **WHEN** the user uploads a `.pdf` file of 5 MB or less as a source
- **THEN** the system extracts readable text and includes it in generation input

#### Scenario: Reject oversized file
- **WHEN** the user uploads a supported-type file larger than 5 MB
- **THEN** the system rejects the upload with an error stating the size limit

#### Scenario: Reject unsupported type
- **WHEN** the user uploads a file that is not `.txt`, `.pdf`, or `.md`
- **THEN** the system rejects the file with an unsupported-type error

### Requirement: AI produces timeline markdown and HTML
Given one or more accepted sources, the system SHALL use AI to produce a timeline document as markdown, SHALL persist that markdown, and SHALL generate HTML from the timeline content for in-app display. The HTML representation MUST be suitable for the same publish/display pattern used by Morph Data publishable documents (Big Notes–style public HTML pages).

#### Scenario: Generation stores both formats
- **WHEN** generation completes successfully from accepted source text
- **THEN** the timeline record stores non-empty markdown content and corresponding HTML content

#### Scenario: Detail shows markdown and HTML
- **WHEN** the user opens a saved timeline
- **THEN** they can view the timeline markdown and an HTML preview

### Requirement: Persist and list timeline outcomes like Big Notes
The system SHALL save every successful timeline generation as a first-class record owned by the user and SHALL list those records in a Timelines workspace comparable to Big Notes (list of saved items, open detail, delete). Users MUST be able to reopen a previously generated timeline and inspect its outcomes.

#### Scenario: New timeline appears in list
- **WHEN** a timeline is created successfully
- **THEN** it appears in the Timelines list with a title and timestamp so the user can open it

#### Scenario: User opens existing timeline
- **WHEN** the user selects a timeline from the list
- **THEN** the detail view loads its saved markdown, HTML, and source summary metadata

#### Scenario: User deletes timeline
- **WHEN** the user confirms delete on a timeline they own
- **THEN** the timeline is removed from the list and is no longer retrievable

### Requirement: Publish timeline HTML publicly
The system SHALL allow publishing a saved timeline’s HTML to a public URL (slug path), similar to Big Notes publish behavior, so viewers can open the timeline without Morph Data authentication.

#### Scenario: Publish allocates public path
- **WHEN** the owner publishes a timeline that has HTML content
- **THEN** the system assigns a published slug/path and returns a public URL

#### Scenario: Public page serves HTML
- **WHEN** an unauthenticated client requests a published timeline slug
- **THEN** the system returns the timeline HTML with content type `text/html`

#### Scenario: Publish without content fails
- **WHEN** the owner attempts to publish a timeline with empty HTML
- **THEN** the system rejects publish with a clear error

### Requirement: Source fetch and extraction failures are explicit
If a URL cannot be fetched or a file cannot be read into usable text, the system SHALL fail the create/generate request with an actionable error and MUST NOT persist a successful timeline for that attempt.

#### Scenario: Unreadable PDF
- **WHEN** a PDF upload yields no extractable text
- **THEN** the system returns an error explaining extraction failed

#### Scenario: Unreachable URL
- **WHEN** the provided URL cannot be fetched successfully
- **THEN** the system returns an error and does not create a timeline
