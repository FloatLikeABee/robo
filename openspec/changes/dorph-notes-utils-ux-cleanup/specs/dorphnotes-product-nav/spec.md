## Purpose

Renames Dorph Data to DorphNotes, drops Assets and User Settings, labels Configuration as Settings, and places the Dorph AI header control beside Notes and left of the theme toggle.

## ADDED Requirements

### Requirement: Product name is DorphNotes

User-visible Dorph Data product chrome SHALL say **DorphNotes** (not Dorph Data or DorphData). This includes the DorphNotes drawer title, browser document title for that app, Dorph AI header app link, and the landing Dorph Data section label.

#### Scenario: Drawer and header show DorphNotes

- **WHEN** a user opens DorphNotes
- **THEN** the navigation product title is DorphNotes

#### Scenario: Dorph AI app link

- **WHEN** a user views Dorph AI header app links
- **THEN** the data app is labeled DorphNotes

### Requirement: Assets module is not a primary surface

DorphNotes SHALL NOT list Assets in the left nav. The default DorphNotes route SHALL be Generic data. Paths that used to open Assets SHALL send the user to Generic data.

#### Scenario: Nav has no Assets

- **WHEN** a user views DorphNotes left navigation
- **THEN** there is no Assets item
- **AND** Generic data is present

#### Scenario: Old Assets URL

- **WHEN** a user opens the former Assets path
- **THEN** they land on Generic data

### Requirement: Configuration is labeled Settings

The DorphNotes nav group formerly labeled Configuration SHALL be labeled **Settings**. Users and File import remain inside it when the user is allowed to see them.

#### Scenario: Settings group

- **WHEN** a user views DorphNotes left navigation
- **THEN** they see Settings (not Configuration) as that group’s label

### Requirement: User Settings page and header control are gone

DorphNotes SHALL NOT offer a User Settings page or a header control that opens it.

#### Scenario: No person shortcut

- **WHEN** a user views the DorphNotes header
- **THEN** there is no User Settings / person control

#### Scenario: Old User Settings URL

- **WHEN** a user opens the former User Settings path
- **THEN** they are sent to a remaining DorphNotes screen (Generic data or Settings)

### Requirement: Header order Notes, Dorph AI, theme

The DorphNotes header SHALL show Notes, then the Dorph AI shortcut, then the theme toggle, adjacent in that order. Dorph AI MUST sit immediately left of the theme toggle.

#### Scenario: Header control order

- **WHEN** a user views the DorphNotes header
- **THEN** Notes is next to Dorph AI
- **AND** Dorph AI is immediately left of the theme button
