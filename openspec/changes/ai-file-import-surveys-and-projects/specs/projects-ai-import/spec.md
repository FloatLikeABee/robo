## Purpose

Lets a Projects user hand over a small bundle of description files and get AI-drafted projects — complete with site logs, people, and flow-log entries — to review and confirm, instead of re-entering a handover pack row by row.

## ADDED Requirements

### Requirement: Analyze uploaded description files

The system SHALL provide an authenticated endpoint that accepts up to three uploaded files and returns an AI-generated draft plan of projects derived from their combined content.

Accepted file types are PDF, TXT, CSV, and MD. Requests containing more than three files SHALL be rejected. An optional free-text instruction field SHALL be accepted to steer the draft.

All uploaded content and all resulting records SHALL be scoped to the authenticated user's organization.

#### Scenario: Draft produced from a single description file

- **WHEN** an authenticated user uploads one PDF describing a construction job
- **THEN** the response contains a draft plan with at least one project

#### Scenario: Three mixed files analyzed together

- **WHEN** a user uploads a PDF brief, a CSV of expenses, and a MD contact list in one request
- **THEN** the content of all three files is combined into a single AI request
- **AND** the draft plan reflects information from each file

#### Scenario: Fourth file rejected

- **WHEN** a request contains four files
- **THEN** the request fails with a message stating at most three files are allowed per import
- **AND** no import session is created

#### Scenario: Unsupported type rejected

- **WHEN** a request includes a `.png` file
- **THEN** the request fails with a message naming PDF, TXT, CSV, and MD as the accepted types

#### Scenario: Instruction steers the draft

- **WHEN** the request includes the instruction `treat each site as its own project`
- **THEN** the draft plan reflects that instruction

#### Scenario: Organization scoping

- **WHEN** a user from organization A completes an import
- **THEN** every created record belongs to organization A
- **AND** users of other organizations cannot read that import session

### Requirement: Draft plan supports multiple projects with nested records

The draft plan SHALL be a list of proposed projects. Each proposed project SHALL carry a name and MAY carry a code, client, location, status, start date, end date, budget total, and description.

Each proposed project MAY additionally carry:

- **site logs** — each with a log date and a summary, and optionally weather, crew count, and issues
- **people** — each with a name, and optionally trade, contact name, phone, email, and description
- **flow log entries** — each with an entry date, a direction of income or expense, and a positive amount, and optionally currency, category, title, notes, and tags

Where the source files describe several distinct jobs, the plan SHALL contain one proposed project per job rather than merging them.

#### Scenario: Two jobs yield two proposed projects

- **WHEN** the uploaded files describe two separately named jobs at different locations
- **THEN** the draft plan contains two proposed projects

#### Scenario: Site logs attached to their project

- **WHEN** a file contains dated daily site entries for a job
- **THEN** those entries appear as proposed site logs nested under that job's proposed project

#### Scenario: People extracted with contact details

- **WHEN** a file lists a contractor with a trade and a phone number
- **THEN** that person appears as a proposed person under the relevant project with the trade and phone populated

#### Scenario: Expense rows become flow log entries

- **WHEN** a CSV contains dated amounts labelled as costs
- **THEN** those rows appear as proposed flow log entries with direction `expense` and positive amounts

#### Scenario: Missing optional fields are omitted

- **WHEN** the source content gives no client or budget for a job
- **THEN** the proposed project omits those fields rather than inventing values

#### Scenario: Project without children

- **WHEN** the source content describes a job with no logs, people, or financial rows
- **THEN** the proposed project is returned with empty lists for those children

### Requirement: Nothing is created until the user confirms

The analyze step SHALL NOT create projects, site logs, people, or flow log entries. Records SHALL be created only by a separate confirm step that names which proposed projects to accept.

The confirm step SHALL allow the user to submit an edited version of the draft, and SHALL create records from the submitted values rather than from the original AI output.

The confirm step SHALL allow a subset of proposed projects to be accepted, and SHALL allow individual nested records to be excluded.

#### Scenario: Analyze creates no records

- **WHEN** an analyze request succeeds with three proposed projects
- **THEN** no project, site log, person, or flow log entry has been created

#### Scenario: Confirm creates the accepted projects

- **WHEN** the user confirms two of three proposed projects
- **THEN** those two projects are created with their included nested records
- **AND** the third proposed project is not created

#### Scenario: Edited values are used

- **WHEN** the user changes a proposed project's name before confirming
- **THEN** the created project uses the edited name

#### Scenario: Excluded child records are skipped

- **WHEN** the user removes one proposed site log before confirming
- **THEN** that site log is not created while the project's other site logs are

#### Scenario: Nested records linked to their new project

- **WHEN** a confirmed project is created with site logs, people, and flow log entries
- **THEN** each created site log and person references the newly created project

#### Scenario: Confirming twice does not duplicate

- **WHEN** a confirm request is submitted for an import session that has already been completed
- **THEN** the request fails with a message stating the session is already completed
- **AND** no duplicate records are created

### Requirement: Created records satisfy existing validation

Records created by the confirm step SHALL satisfy the same validation as records created through the normal create endpoints. A proposed record that would fail validation SHALL be reported as a per-record error rather than silently dropped or written in an invalid state.

A proposed project with a blank name SHALL be rejected. A proposed project with no code SHALL receive a generated code derived from its name. A proposed flow log entry SHALL have a positive amount and a direction of income or expense.

#### Scenario: Blank project name rejected

- **WHEN** a confirmed project has an empty name
- **THEN** that project is reported as an error and is not created

#### Scenario: Missing code is generated

- **WHEN** a confirmed project has a name but no code
- **THEN** a code is generated from the name and the project is created

#### Scenario: Invalid flow log entry reported

- **WHEN** a proposed flow log entry has a zero amount
- **THEN** that entry is reported as an error and is not created
- **AND** the project and its valid children are still created

#### Scenario: Partial success reported

- **WHEN** a confirm request creates a project but one nested person fails validation
- **THEN** the response reports what was created and what failed

### Requirement: Import sessions are persisted and retrievable

The system SHALL persist each import session with its uploaded file metadata, extracted excerpts, the AI draft plan, and a status. Status SHALL progress through analyzing, drafted, completed, and failed states.

An authenticated user SHALL be able to list their organization's import sessions and retrieve a single session including its draft plan, so a draft can be reviewed later.

A completed session SHALL record which records were created from it.

#### Scenario: Session listed after analysis

- **WHEN** an analyze request succeeds
- **THEN** the session appears in the user's session list with status `drafted`

#### Scenario: Draft reopened later

- **WHEN** the user retrieves a previously drafted session by id
- **THEN** the response includes the stored draft plan and file metadata

#### Scenario: Failed analysis recorded

- **WHEN** the AI request fails during analysis
- **THEN** the session is stored with status `failed` and the failure message
- **AND** the endpoint returns an error to the caller

#### Scenario: Completed session records its output

- **WHEN** a session is confirmed and records are created
- **THEN** the session status becomes `completed` and it records the identifiers of the created projects

#### Scenario: Cross-organization access denied

- **WHEN** a user requests a session id belonging to another organization
- **THEN** the request fails as not found

### Requirement: Projects module offers AI import

The Projects module SHALL provide an AI import view where the user can select up to three files, submit them for analysis, review the resulting draft, edit it, and confirm creation.

The view SHALL show a busy state during analysis, list each uploaded file with its read status, and display per-file errors. After a successful confirm it SHALL refresh the projects list so the new projects appear.

#### Scenario: Import view reachable from Projects

- **WHEN** the user opens the Projects module
- **THEN** an AI import view is available alongside the existing create and list views

#### Scenario: File count enforced in the UI

- **WHEN** the user tries to add a fourth file
- **THEN** the interface prevents it and explains the three-file limit

#### Scenario: Draft reviewed and edited before confirming

- **WHEN** analysis returns two proposed projects
- **THEN** both are shown with their nested logs, people, and flow-log entries
- **AND** each project and nested record can be edited or excluded before confirming

#### Scenario: Busy state during analysis

- **WHEN** an analyze request is in flight
- **THEN** the submit control is disabled and shows a busy label

#### Scenario: Per-file read errors shown

- **WHEN** one uploaded PDF yields no extractable text
- **THEN** that file is listed with its error while the draft from the remaining files is still shown

#### Scenario: Projects list refreshed after confirm

- **WHEN** the user confirms an import that creates two projects
- **THEN** the projects list shows the two new projects without a manual reload

#### Scenario: AI unavailable message

- **WHEN** the backend reports that AI is not configured
- **THEN** the view displays that message and does not present an empty draft
