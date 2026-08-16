## Purpose

Defines Morph Data navigation and module surface as Assets-only, removing the People module and the combined Resources label.

## ADDED Requirements

### Requirement: Assets is the sole resources nav item
Morph Data navigation SHALL expose a single **Assets** item for former People/Assets/Resources surfaces. The primary nav label MUST be **Assets**, not Resources. A separate People module or People tab MUST NOT remain.

#### Scenario: Drawer shows Assets
- **WHEN** an authenticated user opens Morph Data navigation
- **THEN** they see **Assets** as the menu label
- **AND** they do not see a top-level **Resources** or **People** item

### Requirement: Assets page has no People tab
The Assets module surface SHALL present assets only. Combined People + Assets tabs MUST NOT remain.

#### Scenario: Open Assets module
- **WHEN** a user opens the Assets route
- **THEN** they can browse and manage assets
- **AND** there is no People tab or People list embedded in that module

### Requirement: Legacy People and Resources paths resolve to Assets
Routes previously used for People or Resources SHALL navigate users to Assets (redirect or equivalent) so bookmarks do not land on a removed People UI.

#### Scenario: Legacy resources or people URL
- **WHEN** a user opens a legacy `/resources` or `/people` admin path
- **THEN** they are taken to the Assets experience

### Requirement: File import drops People as a peer resources target
File import entity options MUST NOT offer a separate "Resources — People" (or equivalent) target. Asset import MAY remain under an Assets-oriented label.

#### Scenario: Import entity list
- **WHEN** a user opens File import entity selection
- **THEN** People is not offered as an import entity under Resources
- **AND** Assets import remains available under an Assets-oriented label
