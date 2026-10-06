## Purpose

Makes the MorphUtils Project module use the same dark-blue chrome as the rest of the product.

## ADDED Requirements

### Requirement: Project chrome is dark blue
The Project page background, header, cards, selected section tab, primary buttons, and create modal MUST use the shared dark-blue surfaces and blue accents. Those surfaces MUST NOT use deep violet as the active color or neutral gray as the page or modal fill. The module MUST stay dark-only and MUST NOT add a light/dark switch.

#### Scenario: Projects page reads as dark blue
- **WHEN** an operator opens Project in MorphUtils
- **THEN** the page background and cards are dark blue
- **AND** the selected section tab and primary button use a blue accent
- **AND** the page is not a violet or neutral-gray theme

#### Scenario: Create modal matches
- **WHEN** the operator opens the create-project modal
- **THEN** the modal surface is the same dark-blue family as the page

### Requirement: Danger confirm stays distinct
A control that confirms a destructive action MAY use a danger color. Ordinary chrome, including non-confirm delete links in the page, MUST NOT use that danger color as a theme.

#### Scenario: Delete confirm is the danger color
- **WHEN** a destructive confirm dialog is open
- **THEN** its confirm control may be a danger color
- **AND** the surrounding Project page remains dark blue
