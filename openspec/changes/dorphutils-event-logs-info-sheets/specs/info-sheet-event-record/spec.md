## Purpose

Turns a completed Info Sheet collection into an Events & Info record via AI summary, while still keeping the original answers on the Info Sheets list.

## ADDED Requirements

### Requirement: Collected answers still save on Info Sheets
Publishing an Info Sheet and collecting a response SHALL continue to store the completed answers on Info Sheets. Collection MUST succeed even when Events & Info recording fails.

#### Scenario: Collect saves the sheet result
- **WHEN** a respondent completes a published Info Sheet
- **THEN** the answers appear in the Info Sheets results list
- **AND** the respondent is shown a successful collect

#### Scenario: Collect succeeds if AI is down
- **WHEN** a respondent completes a published Info Sheet
- **AND** the AI service is not configured or the summary call fails
- **THEN** the Info Sheet result is still saved
- **AND** Events & Info still gets a record built from the sheet title and answers
- **AND** the respondent is shown a successful collect

### Requirement: AI summary becomes an Events & Info record
When a published Info Sheet completion is saved and AI is available, the system MUST create one Events & Info record whose title and detail summarize that response (sheet title, answers, and time). The original answers MUST remain on Info Sheets. Completing the same result twice MUST NOT create duplicate events.

#### Scenario: New collection records an event
- **WHEN** a respondent completes a published Info Sheet
- **AND** AI is configured
- **THEN** Events & Info contains a new record summarizing that collection
- **AND** the Info Sheets result still exists

#### Scenario: No duplicate event for the same result
- **WHEN** the same Info Sheet result is processed again
- **THEN** a second Events & Info record is not created for that result

### Requirement: Info Sheets publish-and-collect stays
Operators MUST still design, publish, unpublish, and review Info Sheets the same way as today’s AI Surveys flow, aside from naming and the Events & Info recording above.

#### Scenario: Publish a sheet
- **WHEN** an operator publishes an Info Sheet
- **THEN** a public collect link is available
- **AND** respondents can submit without an Event Logs login
