## Purpose

Defines Morph Data Tasks after Activities are removed, with no first-class assignee field so optional assignment detail lives in description or JSON.

## ADDED Requirements

### Requirement: Activities module removed
Morph Data MUST NOT expose an Activities (trips) primary navigation item, route, or File import entity type for Activities.

#### Scenario: Activities absent from drawer
- **WHEN** an authenticated user opens Morph Data navigation
- **THEN** Activities is not listed
- **AND** File import does not offer Activities as an import entity

### Requirement: Tasks without assignee
Morph Data Tasks (case/tasks) MUST NOT require or present a first-class assignee field in create/edit UI. The system MUST allow creating and updating a task with title/description (and existing non-assignee fields) without selecting assignees.

#### Scenario: Create task without assignee
- **WHEN** a user creates a task with a title and description and no assignee selection
- **THEN** the task is saved successfully
- **AND** the UI does not block save for missing assignees

### Requirement: Assignment detail via description or JSON
If users need assignment or similar metadata, they SHALL record it in the task description or free-form JSON/detail fields. The product MUST NOT introduce a replacement dedicated assignee picker as part of this change.

#### Scenario: Assignment noted in description
- **WHEN** a user writes assignee information in the task description
- **THEN** that text is stored and shown with the task like other description content
