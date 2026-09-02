## Purpose

Renames Morph Data to MorphNotes, drops Assets and User Settings, labels Configuration as Settings, and places the Morph AI header control beside Notes and left of the theme toggle.

## ADDED Requirements

### Requirement: Product name is MorphNotes

User-visible Morph Data product chrome SHALL say **MorphNotes** (not Morph Data or MorphData). This includes the MorphNotes drawer title, browser document title for that app, Morph AI header app link, and the landing Morph Data section label.

#### Scenario: Drawer and header show MorphNotes

- **WHEN** a user opens MorphNotes
- **THEN** the navigation product title is MorphNotes

#### Scenario: Morph AI app link

- **WHEN** a user views Morph AI header app links
- **THEN** the data app is labeled MorphNotes

### Requirement: Assets module is not a primary surface

MorphNotes SHALL NOT list Assets in the left nav. The default MorphNotes route SHALL be Generic data. Paths that used to open Assets SHALL send the user to Generic data.

#### Scenario: Nav has no Assets

- **WHEN** a user views MorphNotes left navigation
- **THEN** there is no Assets item
- **AND** Generic data is present

#### Scenario: Old Assets URL

- **WHEN** a user opens the former Assets path
- **THEN** they land on Generic data

### Requirement: Configuration is labeled Settings

The MorphNotes nav group formerly labeled Configuration SHALL be labeled **Settings**. Users and File import remain inside it when the user is allowed to see them.

#### Scenario: Settings group

- **WHEN** a user views MorphNotes left navigation
- **THEN** they see Settings (not Configuration) as that group’s label

### Requirement: User Settings page and header control are gone

MorphNotes SHALL NOT offer a User Settings page or a header control that opens it.

#### Scenario: No person shortcut

- **WHEN** a user views the MorphNotes header
- **THEN** there is no User Settings / person control

#### Scenario: Old User Settings URL

- **WHEN** a user opens the former User Settings path
- **THEN** they are sent to a remaining MorphNotes screen (Generic data or Settings)

### Requirement: Header order Notes, Morph AI, theme

The MorphNotes header SHALL show Notes, then the Morph AI shortcut, then the theme toggle, adjacent in that order. Morph AI MUST sit immediately left of the theme toggle.

#### Scenario: Header control order

- **WHEN** a user views the MorphNotes header
- **THEN** Notes is next to Morph AI
- **AND** Morph AI is immediately left of the theme button
