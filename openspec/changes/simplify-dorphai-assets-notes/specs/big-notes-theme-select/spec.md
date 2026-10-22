## Purpose

Defines Big Notes create-flow theme selection as exactly two choices: dark and light.

## ADDED Requirements

### Requirement: Create New theme is dark or light only
When creating a new Big Note, the theme control SHALL offer exactly two selections: **dark** and **light**. Free-text theme entry and any additional named themes MUST NOT be offered in that create dialog.

#### Scenario: Open create dialog
- **WHEN** a user opens Create New for Big Notes
- **THEN** the theme control presents dark and light only

#### Scenario: Default theme
- **WHEN** the create dialog opens
- **THEN** the selected theme defaults to dark unless product UI already stores a different allowed value of dark or light

#### Scenario: Submit create with selected theme
- **WHEN** the user generates a Big Note with theme dark or light
- **THEN** the create request uses that theme value for generated HTML styling
