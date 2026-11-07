# Spec Delta

## Purpose

Makes the Skills modal and the DorphTools app use the same twilight glass as Dorph AI, with no theme switch.

## ADDED Requirements

### Requirement: Skills modal uses twilight glass
The Dorph AI Skills modal SHALL use twilight glass surfaces, light text, and a purple accent. It SHALL NOT use a slate blue panel.

#### Scenario: Open Skills
- **WHEN** an operator opens the Skills modal in Dorph AI
- **THEN** the modal panel is purple glass
- **AND** its text is light

### Requirement: DorphTools uses twilight glass
The DorphTools app, including when it is embedded from Dorph AI, SHALL use twilight glass surfaces and a purple accent. It SHALL NOT offer a light or dark theme switch.

#### Scenario: Open DorphTools
- **WHEN** an operator opens DorphTools
- **THEN** the app background is dark purple
- **AND** primary actions use the purple accent
- **AND** there is no theme switch
