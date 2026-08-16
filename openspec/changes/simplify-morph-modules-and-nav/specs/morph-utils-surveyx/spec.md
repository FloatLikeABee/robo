## Purpose

Defines Morph Utils SurveyX (formerly SheetX) as an AI Surveys product without the Forms product surface.

## ADDED Requirements

### Requirement: SheetX is branded SurveyX
Morph Utils SHALL present the former SheetX module as **SurveyX** in module labels and primary product chrome for that app.

#### Scenario: Utils module list shows SurveyX
- **WHEN** a user views Morph Utils apps
- **THEN** the module formerly labeled SheetX is labeled SurveyX

### Requirement: Forms product surface removed
SurveyX SHALL NOT expose the Forms product navigation or primary Forms management UI. Users MUST NOT be able to open the Forms list/create/edit flows from SurveyX nav.

#### Scenario: Forms nav absent
- **WHEN** a signed-in user opens SurveyX
- **THEN** Forms is not present in primary navigation
- **AND** the default product path is the AI Surveys experience

### Requirement: AI Sheets renamed AI Surveys
The AI Sheets capability SHALL be labeled **AI Surveys** in SurveyX navigation and primary UI.

#### Scenario: AI Surveys label
- **WHEN** a user views SurveyX primary navigation
- **THEN** the AI sheets feature is labeled AI Surveys
