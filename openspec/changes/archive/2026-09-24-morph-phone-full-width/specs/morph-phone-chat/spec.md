# Spec Delta

## ADDED Requirements

### Requirement: Phone chat column fills the viewport
On a portrait viewport up to 430px wide, the Morph AI agent shell (header, transcript, and composer) MUST use the full layout width when the Notes/Knowledge workspace is closed. Ordinary reply prose MUST wrap on word boundaries inside that column. The shell MUST NOT leave the conversation in a narrow left column with an empty region beside it.

#### Scenario: Collapsed workspace at phone width
- **WHEN** Morph AI chat loads at about 390px width with the workspace closed
- **THEN** the header, transcript, and composer each span the layout width
- **AND** a normal reply wraps as words, not as a few letters per line

#### Scenario: Same column at 430px
- **WHEN** the viewport width is 430px and the workspace is closed
- **THEN** the header, transcript, and composer still span that width

### Requirement: Phone workspace stays closed until asked
On a phone, the Notes & TODOs / Context & Knowledge workspace MUST NOT take a bottom band of the chat on first load when the operator has not chosen to leave it open. The operator MUST be able to open and close it. A choice already saved as open or closed MUST be kept. While it is closed, it MUST NOT reduce the transcript height.

#### Scenario: First phone visit
- **WHEN** an operator opens Morph AI chat on a phone with no saved workspace choice
- **THEN** the workspace is closed
- **AND** the transcript is not shortened by a workspace band

#### Scenario: Saved choice
- **WHEN** the operator has saved the workspace open, then reloads on a phone
- **THEN** the workspace stays open
- **AND** when they have saved it closed, a reload keeps it closed

### Requirement: Phone composer does not repeat the workspace tabs
On a phone, the composer MUST NOT show Notes and Knowledge controls that repeat the workspace tab labels. The workspace tabs remain the labeled path to Notes & TODOs and Context & Knowledge. Sending a message MUST still be able to include notes and knowledge.

#### Scenario: Composer on a phone
- **WHEN** the Morph AI composer is shown at about 390px width
- **THEN** Notes and Knowledge chips are not shown above the composer
- **AND** the workspace, once opened, still offers Notes & TODOs and Context & Knowledge
