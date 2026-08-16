## Purpose

Defines Morph Data file import so operators can load CSV, Excel, or JSON into the approved Work data entity types.

## ADDED Requirements

### Requirement: File import replaces Data import naming
Morph Data SHALL present the import feature as **File import** (not Data import) in navigation and primary UI chrome.

#### Scenario: Nav label is File import
- **WHEN** an admin opens Morph Data Configuration
- **THEN** the import entry is labeled File import

### Requirement: Supported file formats
File import SHALL accept CSV, Excel (.xls/.xlsx), and JSON files for import.

#### Scenario: User selects allowed formats
- **WHEN** a user opens File import and chooses a file
- **THEN** the system accepts CSV, Excel, or JSON
- **AND** rejects unsupported types with a clear error

### Requirement: Import entity types limited to Work data set
File import SHALL offer entity types Generic data, People, Assets, and Activities only.

#### Scenario: Entity type picker
- **WHEN** a user starts a File import
- **THEN** the selectable data types are Generic data, People, Assets, and Activities
- **AND** no other entity types are offered in that picker
