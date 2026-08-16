## Purpose

Defines how uploaded files are turned into AI-readable text across Morph apps: which file types are accepted, what limits apply, how PDFs and images are read, and how failures are reported so a partial upload never silently produces an empty result.

## ADDED Requirements

### Requirement: Supported file types

The system SHALL classify each uploaded file as `document`, `image`, or `unsupported` based on its filename extension and declared MIME type, and SHALL reject unsupported files with a message naming the accepted types.

Documents are PDF, TXT, MD, and CSV. Images are JPEG, PNG, GIF, and WebP. Any other type is unsupported.

Individual features MAY accept a narrower set than this list; a feature that accepts only documents SHALL treat an image as unsupported.

#### Scenario: PDF is accepted as a document

- **WHEN** a file named `brief.pdf` with MIME type `application/pdf` is uploaded
- **THEN** it is classified as `document` and accepted for text extraction

#### Scenario: Image is accepted where images are allowed

- **WHEN** a file named `form.jpg` with MIME type `image/jpeg` is uploaded to a feature that accepts images
- **THEN** it is classified as `image` and accepted for visual reading

#### Scenario: Unsupported type is rejected with guidance

- **WHEN** a file named `plan.dwg` is uploaded
- **THEN** the request fails with a message that lists the accepted file types
- **AND** no AI request is made

#### Scenario: Image rejected by a document-only feature

- **WHEN** an image is uploaded to a feature whose accepted types are PDF, TXT, CSV, and MD
- **THEN** the request fails with a message naming those four types

#### Scenario: Extension and MIME type disagree

- **WHEN** a file has a supported extension but a generic MIME type such as `application/octet-stream`
- **THEN** classification falls back to the extension and the file is accepted

### Requirement: Upload limits

The system SHALL enforce a maximum file count per request and a maximum byte size per file, and SHALL reject an over-limit request before making any AI request.

Each feature declares its own file count limit. The per-file size limit SHALL be at most 8 MiB for documents and at most 8 MiB for images.

#### Scenario: Too many files

- **WHEN** a request contains more files than the feature's limit
- **THEN** the request fails with a message stating the maximum number of files allowed
- **AND** no file is stored and no AI request is made

#### Scenario: File too large

- **WHEN** an uploaded file exceeds the per-file byte limit
- **THEN** the request fails with a message naming the offending file and the limit in MiB

#### Scenario: No files supplied

- **WHEN** a request contains no file parts
- **THEN** the request fails with a message stating at least one file is required

### Requirement: Document text extraction

The system SHALL extract plain text from document files server-side. PDF files SHALL be parsed for their embedded text layer. TXT, MD, and CSV files SHALL be read as UTF-8 text, with invalid byte sequences replaced rather than causing failure.

Extraction SHALL NOT store a placeholder such as "content not extracted" in place of real text for a supported document type.

#### Scenario: Text extracted from a PDF

- **WHEN** a PDF containing a selectable text layer is uploaded
- **THEN** its text content is extracted and made available to the AI request

#### Scenario: PDF has no extractable text

- **WHEN** a PDF contains only scanned images and yields no text
- **THEN** that file is reported with an error explaining no text could be extracted
- **AND** other files in the same request are still processed

#### Scenario: CSV rows are preserved

- **WHEN** a CSV file is uploaded
- **THEN** its header row and data rows are included in the extracted text so the AI can interpret columns

#### Scenario: Malformed UTF-8 in a text file

- **WHEN** a TXT file contains invalid UTF-8 byte sequences
- **THEN** the invalid sequences are replaced and the remaining text is extracted successfully

### Requirement: Image content reading

Where a feature accepts images, the system SHALL read the image with a vision-capable AI model and produce a textual description that includes any text visible in the image, so photographed or scanned documents can be used as source content.

#### Scenario: Text in a photographed document is read

- **WHEN** a photo of a printed questionnaire is uploaded
- **THEN** the resulting description includes the questions and answer options visible in the image

#### Scenario: Vision reading fails

- **WHEN** the vision model call fails or returns empty content
- **THEN** that file is reported with an error and the failure message is surfaced to the user

### Requirement: Extracted content truncation

The system SHALL cap the amount of extracted text sent to the AI per file and per request, and SHALL mark truncated content as truncated so the AI and the user both know content was cut.

#### Scenario: Long document is truncated with a marker

- **WHEN** an extracted document exceeds the per-file character cap
- **THEN** the text is truncated at the cap and an explicit truncation marker is appended
- **AND** the request proceeds rather than failing

### Requirement: AI availability reporting

When no AI credentials are configured, the system SHALL reject ingestion requests with a distinct, actionable error rather than returning an empty or fabricated result.

#### Scenario: AI is not configured

- **WHEN** an ingestion request is made and `MORPH_AI_API_KEY` is unset
- **THEN** the request fails with HTTP 503 and a message stating that AI is not configured

### Requirement: Per-file result reporting

The system SHALL return a per-file result for every uploaded file, identifying the file by its original name and reporting either the content read from it or the error that prevented reading it. A failure on one file SHALL NOT abort processing of the remaining files.

#### Scenario: Mixed success and failure

- **WHEN** a request contains one readable file and one unreadable file
- **THEN** the response includes a successful result for the readable file and an error result for the unreadable one
- **AND** the AI request proceeds using only the readable content

#### Scenario: Every file fails

- **WHEN** no uploaded file yields usable content
- **THEN** the request fails with an error explaining that no content could be read
- **AND** no AI generation request is made
