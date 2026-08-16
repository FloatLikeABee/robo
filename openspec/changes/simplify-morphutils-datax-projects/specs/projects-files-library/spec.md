## Purpose

Limit the Morph Engi / Morph Utils Projects app to Projects and Files only, with Files as a simple uploaded-content library.

## ADDED Requirements

### Requirement: Primary nav is Projects and Files only
The Projects product primary navigation SHALL expose only **Projects** and **Files**. **People** and **Flow log** MUST NOT appear as primary nav modules.

#### Scenario: Nav shows two modules
- **WHEN** an authenticated user opens the Projects app
- **THEN** navigation lists Projects and Files and does not list People or Flow log

#### Scenario: Legacy people/flow paths retired from primary UX
- **WHEN** a user attempts to open former People or Flow log primary routes/tabs
- **THEN** they are redirected to Projects or Files, or the modules are otherwise unavailable in the primary UI

### Requirement: Files module lists retained content files
The Files module SHALL list uploaded files and paste-saved content files so the user can browse them. Files is a retention/list surface — not a full project CRM.

#### Scenario: Empty Files list
- **WHEN** the user opens Files and no files exist
- **THEN** the UI shows an empty state indicating no files yet

#### Scenario: User sees uploaded file in list
- **WHEN** at least one file has been uploaded or saved from paste
- **THEN** it appears in the Files list with an identifiable name and timestamp

#### Scenario: User can open or download a listed file
- **WHEN** the user selects a file from the list
- **THEN** they can view or download its content according to the product’s file handling

### Requirement: Morph Utils Project module description matches two-module scope
Morph Utils SHALL describe the Project embed as projects and files (not people/flow-log job tracking).

#### Scenario: Shell copy updated
- **WHEN** a user views the Morph Utils Project module description
- **THEN** it does not advertise People or Flow log as primary features
