## REMOVED Requirements

### Requirement: A public MorphUtils URL is inlined into the production bundle
**Reason**: MorphUtils is not a product entry. MorphNotes opens inside Morph AI.
**Migration**: Remove the MorphUtils header chip. Do not render a chip from `REACT_APP_MORPH_UTILS_URL`.

### Requirement: An unset or loopback URL omits the chip
**Reason**: There is no MorphUtils chip to omit.
**Migration**: The header never shows MorphUtils, in development or production.

## ADDED Requirements

### Requirement: The header has no MorphUtils chip
The Morph AI header MUST NOT render a MorphUtils chip in development or production, including when `REACT_APP_MORPH_UTILS_URL` is a public URL. MorphNotes MUST open the in-app MorphNotes modal.

#### Scenario: A public MorphUtils URL does not add a chip
- **WHEN** the UI is built with `REACT_APP_MORPH_UTILS_URL` set to a public HTTPS origin
- **THEN** the header does not show a MorphUtils chip
- **AND** activating MorphNotes opens the in-app modal
