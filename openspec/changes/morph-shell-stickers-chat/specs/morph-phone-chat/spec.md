## MODIFIED Requirements

### Requirement: Phone chat column fills the viewport
On a portrait viewport up to 430px wide, the Morph AI agent shell (header, transcript, and composer) MUST use the full layout width. The workspace band MAY sit below the chat and share vertical space. Ordinary reply prose MUST wrap on word boundaries inside that column. The shell MUST NOT leave the conversation in a narrow left column with an empty region beside it.

#### Scenario: Phone width uses the full column
- **WHEN** Morph AI chat loads at about 390px width
- **THEN** the header, transcript, and composer each span the layout width
- **AND** a normal reply wraps as words, not as a few letters per line

#### Scenario: Same column at 430px
- **WHEN** the viewport width is 430px
- **THEN** the header, transcript, and composer still span that width

### Requirement: Phone composer does not repeat the workspace tabs
On a phone, the composer MUST NOT show Notes and Knowledge controls that repeat the workspace tab labels. The workspace tabs remain the labeled path to Notes & TODOs and Context & Knowledge. Sending a message MUST still be able to include notes and knowledge.

#### Scenario: Composer on a phone
- **WHEN** the Morph AI composer is shown at about 390px width
- **THEN** Notes and Knowledge chips are not shown above the composer
- **AND** the workspace still offers Notes & TODOs and Context & Knowledge

## REMOVED Requirements

### Requirement: Phone workspace stays closed until asked
**Reason**: The workspace pane is always visible, including the phone band, and the hide/show control is gone.
**Migration**: Ignore a saved closed preference. The phone layout keeps the chat at full width with the workspace in the lower band.

## ADDED Requirements

### Requirement: Phone workspace stays visible
On a phone, the Notes & TODOs / Context & Knowledge workspace MUST be visible in the lower band on first load and after reload. The shell MUST NOT provide a control to close it. A saved closed preference MUST NOT hide it.

#### Scenario: First phone visit
- **WHEN** an operator opens Morph AI chat on a phone with no saved workspace choice
- **THEN** the workspace band is visible below the chat
- **AND** there is no control to hide it

#### Scenario: Saved closed preference
- **WHEN** the operator has a saved workspace-closed preference and reloads on a phone
- **THEN** the workspace band is still visible
