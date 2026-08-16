## Purpose

Defines Morph Data drawer navigation and fixed entity labels so Work data lives under Big notes with only the approved entity set.

## ADDED Requirements

### Requirement: Work data entities nest under Big notes
Morph Data navigation SHALL present Generic data, People, Assets, and Activities as items nested under Big notes. The system MUST NOT show a separate top-level Work data group for these entities, and MUST NOT list Places (or districts/facilities) among those nested Work data items.

#### Scenario: Drawer shows entities under Big notes
- **WHEN** an authenticated user opens the Morph Data navigation drawer
- **THEN** Generic data, People, Assets, and Activities appear under Big notes
- **AND** Places is not listed among those items

### Requirement: Fixed People label replaces Members
The Morph Data UI SHALL label the members/people entity as **People**. The system MUST NOT require or offer a configuration UI to change that display name (or other Work data entity display names).

#### Scenario: People label is hardcoded
- **WHEN** a user views Morph Data navigation for the people/members entity
- **THEN** the visible label is People
- **AND** there is no Display labels configuration entry to rename it

### Requirement: Display labels configuration removed
Morph Data SHALL NOT expose a Display labels (editable display names) configuration page or equivalent settings that let users rename nav/entity labels.

#### Scenario: Display labels route gone
- **WHEN** a user opens Morph Data Configuration
- **THEN** Display labels is not available as a navigation item
