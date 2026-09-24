## Purpose

Keep the main Morph SPA shell inside a phone viewport, including safe areas, so the page does not scroll sideways and chrome is not hidden under the status bar or home indicator.

## ADDED Requirements

### Requirement: Phone viewport does not scroll the shell sideways
On a layout viewport about 360–430px wide, every main Morph route (sign-in, Morph AI chat, skills, and MorphNotes) MUST lay out so the document does not scroll horizontally for shell chrome or the primary content column. Shell chrome MUST NOT provide its own horizontal scrollbar. A nested region such as a code block, markdown table, or data grid MAY scroll inside a box that itself fits in the column.

#### Scenario: Morph AI at phone width
- **WHEN** an operator opens Morph AI chat at about 390px CSS width
- **THEN** the document scroll width is not greater than the layout viewport
- **AND** the header, including its actions, does not scroll horizontally

#### Scenario: MorphNotes at phone width
- **WHEN** an operator opens a MorphNotes route at about 390px CSS width
- **THEN** the document scroll width is not greater than the layout viewport
- **AND** the navigation chrome and primary column fit that width

#### Scenario: Widths from 360 to 430
- **WHEN** the portrait viewport width changes between about 360px and 430px
- **THEN** the shell reflows inside the viewport
- **AND** fixed chrome does not cover the primary column

### Requirement: Safe areas stay visible on first paint
With `viewport-fit=cover`, top and bottom content on main Morph routes MUST stay clear of `env(safe-area-inset-top)` and `env(safe-area-inset-bottom)`. That padding MUST be in CSS that applies on first paint, not only after a client media-query hook updates.

#### Scenario: Notched phone
- **WHEN** the viewport reports non-zero safe-area insets
- **THEN** the top of the shell content is inset by at least the top safe area
- **AND** the bottom of the shell content is inset by at least the bottom safe area

#### Scenario: MorphNotes before hydration effects
- **WHEN** MorphNotes first paints at a phone breakpoint
- **THEN** the home-indicator spacer is already in the layout
- **AND** it does not wait for a JavaScript media-query state update

### Requirement: First paint does not flash a desktop-wide shell
The initial stylesheet MUST constrain the document to the viewport width before route content paints. The page MUST NOT depend on a class added after mount to prevent a desktop-wide horizontal overflow.

#### Scenario: Empty or loading shell
- **WHEN** the shell paints before chat messages or MorphNotes rows are loaded
- **THEN** the document does not overflow horizontally

### Requirement: Error banners fit the phone width
An error, success, or offline-style banner on a main Morph route MUST fit the viewport width and MUST NOT force the document to scroll horizontally.

#### Scenario: Chat error bubble
- **WHEN** Morph AI shows an error message at phone width
- **THEN** the error fits inside the primary column
- **AND** a long unbroken error string wraps instead of widening the document

#### Scenario: MorphNotes alert
- **WHEN** MorphNotes shows an alert at phone width
- **THEN** the alert fits the primary column width
