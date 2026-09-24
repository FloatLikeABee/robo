## Purpose

Makes the main Morph SPA header, app chips, and drawers usable at phone width so a person can reach every primary control and dismiss a sheet without sideways scrolling.

## ADDED Requirements

### Requirement: Phone header keeps primary actions on screen
At a viewport width of 768px or less, the Morph AI header SHALL show the sessions menu (when the session list exists), the sign-out control (when the chat is not embedded), and a More control, each at least 44×44 CSS pixels. App and module chips SHALL be reachable from that More control. The header SHALL NOT scroll horizontally.

#### Scenario: Primary actions fit at phone width
- **WHEN** the Morph AI header is shown at about 390px width and the chat is not embedded
- **THEN** the sessions menu, sign-out control, and More control are visible without scrolling the header sideways
- **AND** each of those controls is at least 44×44 CSS pixels

#### Scenario: App chips open from More
- **WHEN** the person activates More at phone width
- **THEN** Skills, AI tools, and MorphNotes are listed with visible text labels
- **AND** MorphUtils is listed only when a non-loopback MorphUtils URL is configured
- **AND** each listed target is at least 44 CSS pixels tall
- **AND** activating a listed target runs the same action as the desktop chip

#### Scenario: More dismisses without leaving the header scrolling
- **WHEN** More is open
- **THEN** it can be dismissed by a close control or by activating the backdrop behind the menu
- **AND** the header still does not scroll horizontally

### Requirement: Phone touch targets for module navigation
On a viewport of 768px or less, each MorphNotes module navigation row and each Morph AI session row in the open sessions sheet SHALL be at least 44 CSS pixels tall. The phone close control on those sheets SHALL be at least 44×44 CSS pixels.

#### Scenario: MorphNotes modules are tappable
- **WHEN** the MorphNotes navigation sheet is open at about 390px width
- **THEN** each module row is at least 44 CSS pixels tall
- **AND** the close control is at least 44×44 CSS pixels

#### Scenario: Session rows are tappable
- **WHEN** the Morph AI sessions sheet is open at about 390px width
- **THEN** each session row is at least 44 CSS pixels tall

### Requirement: Phone drawers are overlay sheets
On a viewport of 768px or less, an opened MorphNotes navigation drawer, Morph AI sessions sheet, context/knowledge sheet, AI tools sheet, or MorphNotes case/task detail drawer SHALL cover the full width of the layout viewport (100%, not `100vw`). Each of those sheets SHALL be dismissible by its close control. A sheet that leaves any backdrop visible SHALL also dismiss when that backdrop is activated. While a sheet is open and the soft keyboard is not involved, the background page SHALL NOT scroll sideways, and focus SHALL stay inside the sheet until it closes. When every such sheet is closed, chrome SHALL NOT permanently reserve more than 40% of the viewport width.

#### Scenario: MorphNotes navigation is a full-width sheet
- **WHEN** the person opens MorphNotes navigation at about 390px width
- **THEN** the sheet covers the full layout width
- **AND** the page behind it does not scroll sideways
- **AND** the close control or the backdrop dismisses it

#### Scenario: Sessions sheet traps focus
- **WHEN** the Morph AI sessions sheet is open at phone width
- **THEN** it covers the full layout width
- **AND** focus is inside the sheet
- **AND** Escape or the close control dismisses it
- **AND** the background does not scroll sideways

#### Scenario: Closed chrome does not reserve a wide rail
- **WHEN** no drawer or sheet is open at about 390px width
- **THEN** neither the MorphNotes navigation nor the Morph AI sessions list permanently occupies more than 40% of the viewport width
