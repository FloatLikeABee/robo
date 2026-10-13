## ADDED Requirements

### Requirement: Grill start accepts every note kind
`note_grill` start MUST accept `kind` values `note`, `todo`, `big_note`, `timeline`, `research`, `stick`, and `task`. An unknown kind MUST be `invalid kind`. The interview MUST stay in memory and MUST NOT insert a row. The draft MUST be body text for that kind (a TODO stays bullets; a timeline draft lists dated entries; research names findings). Another user's grill id MUST still be not found.

#### Scenario: Start a timeline grill
- **WHEN** a client starts `note_grill` with kind `timeline` and a title
- **THEN** the result is grill state for that session user
- **AND** the state kind is `timeline`
- **AND** no timeline row is written

#### Scenario: Unknown kind
- **WHEN** a client starts `note_grill` with kind `banana`
- **THEN** the result is an invalid-kind error
- **AND** no session is stored
