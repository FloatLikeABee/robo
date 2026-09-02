## Purpose

Lets Morph Data Generic Data import full PDFs, CSV, and Excel, persist the content, and create records from pasted text via AI JSON extract.

## ADDED Requirements

### Requirement: PDF import captures the full text layer
When a user imports a text-based PDF into Generic Data, the system SHALL extract the document's text layer for the whole file (all pages), not a partial scrape of embedded strings. Image-only / scanned PDFs with no text layer MAY be rejected with a clear error.

#### Scenario: Import a multi-page text PDF
- **WHEN** a user imports a multi-page PDF that has a text layer
- **THEN** Generic Data stores readable content covering the whole document
- **AND** later viewing or AI analysis can use that stored content

#### Scenario: Import a scanned PDF with no text
- **WHEN** a user imports a PDF that has no extractable text layer
- **THEN** the import fails with an error that the PDF has no extractable text
- **AND** no empty Generic Data record is saved as a successful import

### Requirement: CSV and Excel import into Generic Data
Generic Data import SHALL accept CSV and Excel spreadsheet files (`.csv`, `.xlsx`, and `.xls` when the platform can parse them). Parsed columns and rows MUST be stored on the Generic Data record so the table view can display them.

#### Scenario: Import a CSV file
- **WHEN** a user imports a valid `.csv` file
- **THEN** the record is saved with source type CSV
- **AND** the stored detail includes columns and row values from the file

#### Scenario: Import an Excel workbook
- **WHEN** a user imports a valid `.xlsx` file
- **THEN** the record is saved
- **AND** the stored detail includes tabular rows from the workbook (first sheet if the file has several)

#### Scenario: Unsupported spreadsheet
- **WHEN** a user imports a spreadsheet type the importer cannot parse
- **THEN** the import fails with an unsupported-type error
- **AND** no empty record is saved as a successful import

### Requirement: Successful import persists content
A successful Generic Data import MUST persist both the list row (title, source type, filename, record count) and the imported content in the record's detail store. If detail storage fails after the row is created, the API MUST return an error that the content was not saved (not a silent empty record).

#### Scenario: Import saves list row and detail
- **WHEN** a user imports a supported file and parsing succeeds
- **THEN** the new Generic Data item appears in the list
- **AND** opening it shows the imported content (markdown, table, or JSON payload)

#### Scenario: Detail store write fails
- **WHEN** parsing succeeds but storing imported content fails
- **THEN** the client is told the import did not fully save
- **AND** the user is not shown a successful empty record as if content were stored

### Requirement: Paste text and extract JSON before file import
The Generic Data import flow SHALL let the user paste plain text (without a file) and run AI extract to structured JSON. The user MUST be able to review the JSON and save it as a Generic Data record.

#### Scenario: Extract JSON from pasted text and save
- **WHEN** a user opens Generic Data import
- **AND** they paste descriptive text and choose AI extract
- **THEN** the system returns a JSON object derived from that text
- **AND** after the user confirms, a Generic Data record is saved with that JSON as content

#### Scenario: Extract without saving
- **WHEN** a user pastes text and runs AI extract
- **AND** they cancel without saving
- **THEN** no Generic Data record is created

#### Scenario: AI extract unavailable
- **WHEN** a user requests AI extract
- **AND** the AI service is not configured
- **THEN** the flow shows that AI extract is unavailable
- **AND** no record is created
