# Spec Delta

## Purpose

Defines the visual behavior of Dorph AI chat so the interface presents a consistent twilight glass aesthetic—mid-luminance, purple and orange accents, frosted surfaces, and accessible fallbacks—across standalone and embedded chat.

## ADDED Requirements

### Requirement: Twilight glass preset is the default chat theme

The Dorph AI chat shell SHALL apply a single default visual preset (twilight glass) on load for standalone and embedded chat. The preset SHALL use a mid-luminance page background (neither near-black nor near-white). Primary accent color SHALL read as purple; secondary highlight color SHALL read as orange on interactive emphasis (send button, active session, links).

#### Scenario: Standalone chat loads

- **WHEN** a user opens Dorph AI chat without a stored theme override
- **THEN** the chat root sets the twilight glass preset on the document chat theme attribute
- **AND** the page background and accents match the purple/orange twilight glass token set

#### Scenario: Embedded chat inherits preset

- **WHEN** Dorph AI chat is embedded inside DorphNotes or another host shell
- **THEN** the same twilight glass preset applies to the embedded chat root
- **AND** host chrome outside the chat iframe is unchanged

### Requirement: Glass surfaces on primary chat regions

Primary chat regions (header bar, session list panel, message column backdrop, composer/input bar, and modal sheets that reuse chat surface tokens) SHALL use semi-transparent fills with backdrop blur so underlying dispersion shows through. Borders SHALL be soft (low-contrast) rather than hard opaque lines.

#### Scenario: Message area readability

- **WHEN** assistant or user messages render on the glass message column
- **THEN** message bubbles remain readable against the dispersed background
- **AND** bubble backgrounds use frosted or tinted glass distinct from the page mesh

### Requirement: Dispersed background without layout change

The chat page background SHALL include at least two soft color blooms (purple-leaning and orange-leaning) that visually disperse across the viewport. The blooms SHALL NOT shift chat layout, scroll containers, or touch targets.

#### Scenario: Reduced motion preference

- **WHEN** the user agent reports `prefers-reduced-motion: reduce`
- **THEN** any optional background animation is disabled or replaced with a static gradient
- **AND** glass styling and readability remain unchanged

### Requirement: Accessibility fallbacks for transparency

When the user agent reports `prefers-reduced-transparency: reduce`, or when backdrop blur is unsupported, chat surfaces SHALL fall back to opaque or high-opacity fills that preserve text contrast. Body text and primary button labels SHALL meet WCAG 2.1 AA contrast against their immediate background in the default preset.

#### Scenario: Reduced transparency

- **WHEN** `prefers-reduced-transparency: reduce` is active
- **THEN** header, sidebar, and composer use opaque surface colors from the same token set
- **AND** primary button label contrast remains at least 4.5:1 against its button background

### Requirement: Phone chrome constraints preserved

At viewport width 768px or less, twilight glass styling SHALL NOT reduce touch targets or drawer behavior required by existing phone chrome specs. Session rows, header controls, and sheet dismiss controls SHALL remain at least 44×44 CSS pixels.

#### Scenario: Sessions sheet on phone

- **WHEN** the user opens the sessions sheet on a 768px-wide viewport
- **THEN** session rows remain at least 44px tall
- **AND** the sheet uses the glass preset without horizontal overflow
