## Purpose

Defines the slim Booki navigation set exposed inside Morph Utils: Booki AI, Bookings, and Flow log only.

## ADDED Requirements

### Requirement: Booki nav limited to three surfaces
Booki SHALL expose only **Booki AI** (formerly Data AI), **Bookings**, and **Flow log** in primary navigation. Accounting, Warehouse, Assets, module Settings, and any other Booki primary nav items MUST be removed.

#### Scenario: Slim Booki drawer
- **WHEN** a signed-in user opens Booki
- **THEN** primary navigation includes Booki AI, Bookings, and Flow log
- **AND** Accounting, Warehouse, Assets, and Settings are not listed

### Requirement: Data AI renamed Booki AI
The former Data AI home experience SHALL be labeled **Booki AI**.

#### Scenario: Home label
- **WHEN** a user views Booki primary navigation
- **THEN** the AI home entry is labeled Booki AI
