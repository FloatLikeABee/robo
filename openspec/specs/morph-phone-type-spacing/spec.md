# morph-phone-type-spacing Specification

## Purpose

Keep body copy, stacks, empty states, and error banners on the main Morph SPA readable and tappable at phone width after the structural mobile fixes.

## Requirements

### Requirement: Phone body copy is readable without pinch-zoom
On a viewport about 390px wide, normal body copy on the Morph sign-in screen, the Morph AI chat transcript (including the empty welcome), and MorphNotes notes surfaces SHALL be at least 16px with a line-height of at least 1.5. This applies to field labels, helper lines under a title, message bubbles, and empty-state sentences. It does not require captions, badges, or monospace code to use that size.

#### Scenario: Sign-in copy at phone width
- **WHEN** the sign-in screen is shown at about 390px width
- **THEN** the field labels, the line under the title, and any error text are at least 16px with a line-height of at least 1.5

#### Scenario: Chat transcript at phone width
- **WHEN** Morph AI shows a message or the empty welcome at about 390px width
- **THEN** that body copy is at least 16px with a line-height of at least 1.5

#### Scenario: Notes copy at phone width
- **WHEN** a MorphNotes notes surface shows an empty sentence or an alert at about 390px width
- **THEN** that copy is at least 16px with a line-height of at least 1.5

### Requirement: Phone stacks do not collide
On a viewport about 390px wide, stacked sections and lists on those screens SHALL separate neighboring controls by a visible gap. The chat title SHALL ellipsize inside the header instead of painting over the header actions.

#### Scenario: Header title yields to actions
- **WHEN** the Morph AI header is shown at about 390px width with a long title
- **THEN** the title truncates with an ellipsis
- **AND** the title does not cover the header actions

#### Scenario: Stacked controls keep a gap
- **WHEN** sign-in fields, chat messages, or a notes list are stacked at about 390px width
- **THEN** neighboring controls are separated by a gap of at least 8px

### Requirement: Phone empty states fit and stay tappable
An in-app empty state for no chat messages, no skills, or no notes SHALL wrap inside the column at about 390px width. Any button in that empty state SHALL be at least 44×44 CSS pixels.

#### Scenario: Chat welcome fits
- **WHEN** Morph AI has no messages at about 390px width
- **THEN** the welcome mark and title fit the column width

#### Scenario: No-notes copy wraps and the create control is tappable
- **WHEN** a notes surface shows that there are no notes at about 390px width
- **THEN** the empty sentence wraps inside the column
- **AND** the control that creates a note is at least 44×44 CSS pixels

#### Scenario: No-skills copy wraps
- **WHEN** a skills surface shows that no skills are available at about 390px width
- **THEN** that sentence wraps inside the column and is at least 16px

### Requirement: Phone error banners are readable and dismissible
An error banner or toast on sign-in, Morph AI chat, or a MorphNotes notes surface SHALL wrap inside the column, use at least 16px type with a line-height of at least 1.5, and offer a dismiss control of at least 44×44 CSS pixels.

#### Scenario: Sign-in error can be dismissed
- **WHEN** sign-in shows an error at about 390px width
- **THEN** the error text wraps inside the card
- **AND** a dismiss control of at least 44×44 CSS pixels clears it

#### Scenario: Chat error can be dismissed
- **WHEN** Morph AI shows an error bubble at about 390px width
- **THEN** the error text wraps inside the transcript column
- **AND** a dismiss control of at least 44×44 CSS pixels removes that bubble

#### Scenario: Notes alert can be dismissed
- **WHEN** a MorphNotes notes surface shows an error alert at about 390px width
- **THEN** the alert text wraps inside the column
- **AND** a dismiss control of at least 44×44 CSS pixels clears it
