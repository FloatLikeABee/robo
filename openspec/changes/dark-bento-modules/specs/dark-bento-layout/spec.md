# Spec Delta

## Purpose

Puts every operator module on one dark bento: modular tiles, an 8px rhythm, and the sample layout rules, without a light page.

## ADDED Requirements

### Requirement: Pages use a balanced dark bento
Operator screens SHALL sit on a page darker than the tiles. Tiles SHALL be mid-dark purple, not white, cream, or a single flat lavender slab. A primary view SHALL show at least one wide tile and smaller tiles in a grid with a 16px gap and about 16px corners.

#### Scenario: Module home
- **WHEN** an operator opens Dorph AI, DorphNotes, Content, Event Logs, or DorphTools
- **THEN** the page background is darker than the cards
- **AND** the cards are not a light-mode white or cream surface

### Requirement: Spacing uses an 8px scale
Padding and gaps on these screens SHALL use multiples of 8 pixels. A control that is meant to be tapped SHALL be at least 40 pixels tall.

#### Scenario: Card padding
- **WHEN** a bento tile shows a title and body
- **THEN** the space inside the tile and between tiles is a multiple of 8 pixels

### Requirement: Navigation shows an icon and a label
A navigation item SHALL show an icon and a text label. Navigation SHALL NOT add a third nested level.

#### Scenario: Drawer item
- **WHEN** an operator looks at a module navigation item
- **THEN** they see both an icon and a label

### Requirement: First load uses a skeleton
When a list or card grid is loading for the first time, the screen SHALL show skeleton shapes that match those cards. It SHALL NOT show only a centered spinner.

#### Scenario: List still loading
- **WHEN** a module list has not returned yet
- **THEN** the operator sees placeholder cards
- **AND** they do not see only a spinner

### Requirement: Long card text offers Read more
Card body text longer than a short paragraph SHALL clamp and offer Read more. Expanding it SHALL show the rest in place.

#### Scenario: Long description
- **WHEN** a card’s body is longer than the clamp
- **THEN** the operator can choose Read more and see the rest

### Requirement: Preview and text sit in columns
On a wide screen, a preview and its source text SHALL sit in two columns. They SHALL NOT stack the preview as a full-width block above a wrapped paragraph when both fit side by side.

#### Scenario: Wide Content preview
- **WHEN** the Content preview and source text are both visible on a wide screen
- **THEN** they sit in two columns
