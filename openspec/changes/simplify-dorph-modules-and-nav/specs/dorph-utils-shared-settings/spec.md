## Purpose

Defines a single Dorph Utils Settings at shell level so embedded modules do not each own settings or profile screens.

## ADDED Requirements

### Requirement: Shared Settings at Dorph Utils app level
Dorph Utils SHALL provide one **Settings** entry at the same navigation level as the apps (Dorph Utils shell), not inside each embedded module.

#### Scenario: Settings beside apps
- **WHEN** a signed-in user views Dorph Utils shell navigation (desktop or mobile More/menu)
- **THEN** Settings is available at the shell/app level alongside module switching

### Requirement: Modules do not own Settings or Profile
Embedded Dorph Utils modules (SurveyX, Booki, Academi, and other Utils apps) SHALL NOT expose module-local Settings or Profile primary nav items for account/product settings that belong in the shared shell Settings.

#### Scenario: No per-module settings nav
- **WHEN** a user opens any Dorph Utils embedded module
- **THEN** that module does not show its own Settings or Profile primary nav for shared account settings
