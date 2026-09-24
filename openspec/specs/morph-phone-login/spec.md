# morph-phone-login Specification

## Purpose

Make the main Morph SPA sign-in screen usable on a phone: the form fits, the on-screen keyboard does not trap the fields or Sign in, errors stay readable, and the page after sign-in is still a phone-width shell.

## Requirements

### Requirement: Sign-in form fits a phone portrait
At a layout viewport about 390px wide, the `/login` screen MUST show the username field, password field, their labels, and the Sign in control without the document scrolling horizontally. Each text field MUST use a font size of at least 16px so focusing it does not trigger a browser zoom that pans the page. The shell MUST keep `env(safe-area-inset-*)` padding from CSS on first paint.

#### Scenario: Portrait phone
- **WHEN** an operator opens `/login` at about 390px CSS width in portrait
- **THEN** the document scroll width is not greater than the layout viewport
- **AND** the username field, password field, labels, and Sign in control are inside that width

#### Scenario: Field font does not zoom the page
- **WHEN** an operator focuses the username or password field on a phone
- **THEN** the field's computed font size is at least 16px

### Requirement: Focused fields and Sign in stay reachable with the keyboard
While a sign-in field is focused and the on-screen keyboard covers the bottom of the layout viewport, the focused field MUST remain inside the visible viewport, or the sign-in shell MUST scroll so the Sign in control is reachable. On a phone-width portrait layout the form MUST sit at the top of the shell so a typical keyboard does not cover it before any script runs.

#### Scenario: Keyboard covers the lower half
- **WHEN** an operator focuses a sign-in field and the visible viewport is shorter than the layout viewport by at least a keyboard
- **THEN** the focused field is inside the visible viewport or the operator can scroll the shell until Sign in is inside the visible viewport

#### Scenario: Phone layout before script
- **WHEN** `/login` first paints at about 390px width
- **THEN** the sign-in card is aligned to the start of the shell rather than vertically centered

### Requirement: Failed sign-in text stays on screen
A failed sign-in MUST show the error text inside the card. A long or unbroken error string MUST wrap instead of widening the document. The error MUST be exposed as an alert so it is announced.

#### Scenario: Long error
- **WHEN** sign-in fails with a long error message at about 390px width
- **THEN** the message is visible inside the card
- **AND** the document scroll width is not greater than the layout viewport

### Requirement: Sign in is a 44px touch target
The Sign in control MUST have a hit area of at least 44 by 44 CSS pixels, including while it is disabled during a request. This screen MUST NOT add a password-reset, invite, or OAuth link.

#### Scenario: Touch Sign in
- **WHEN** an operator views `/login` at phone width
- **THEN** the Sign in control is at least 44px wide and 44px tall

### Requirement: Post-login landing stays a phone shell
After a successful sign-in, navigation to the return path (default `/`) MUST land on the existing Morph shell without a leftover modal backdrop covering the page. At about 390px width that landing MUST NOT scroll the document horizontally. This requirement does not change header, drawer, or composer layout.

#### Scenario: Land on Morph AI
- **WHEN** sign-in succeeds and the app opens the default landing at about 390px width
- **THEN** the document scroll width is not greater than the layout viewport
- **AND** no hidden modal backdrop covers the shell
