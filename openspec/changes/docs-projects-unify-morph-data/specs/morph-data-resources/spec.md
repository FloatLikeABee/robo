## Purpose

Defines the combined Morph Data resources module that merges People and Assets into one navigable surface and import target.

## ADDED Requirements

### Requirement: Single Resources module replaces People and Assets
Morph Data navigation SHALL expose one combined module for the former People and Assets surfaces (label may be **Resources** or an equivalent single label). Separate primary nav items for People and Assets MUST NOT remain.

#### Scenario: One nav item for people and assets
- **WHEN** an authenticated user opens Morph Data navigation under Big notes (or equivalent Work data area)
- **THEN** a single combined module is listed instead of separate People and Assets items

### Requirement: Combined list and detail
The combined module SHALL let users browse and open both people-type and asset-type records from one list/detail experience (tabs, filters, or unified rows are acceptable). Users MUST be able to reach existing people and asset records without separate top-level modules.

#### Scenario: Open a person or asset from the combined module
- **WHEN** a user opens the combined resources module
- **THEN** they can locate and open both people and asset records from that module

### Requirement: File import targets the combined module
File import entity options SHALL include the combined resources target (or explicit people/asset subtypes under that module) and MUST NOT list People and Assets as two peer top-level Work entities alongside a removed Activities type.

#### Scenario: Import entity list updated
- **WHEN** a user opens File import entity selection
- **THEN** Activities is not offered
- **AND** people/asset import is available via the combined resources model (not as two separate peer nav modules)
