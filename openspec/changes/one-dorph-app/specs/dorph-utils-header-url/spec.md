## REMOVED Requirements

### Requirement: A public DorphUtils URL is inlined into the production bundle
**Reason**: DorphUtils is not a product entry. DorphNotes opens inside Dorph AI.
**Migration**: Remove the DorphUtils header chip. Do not render a chip from `REACT_APP_DORPH_UTILS_URL`.

### Requirement: An unset or loopback URL omits the chip
**Reason**: There is no DorphUtils chip to omit.
**Migration**: The header never shows DorphUtils, in development or production.

## ADDED Requirements

### Requirement: The header has no DorphUtils chip
The Dorph AI header MUST NOT render a DorphUtils chip in development or production, including when `REACT_APP_DORPH_UTILS_URL` is a public URL. DorphNotes MUST open the in-app DorphNotes modal.

#### Scenario: A public DorphUtils URL does not add a chip
- **WHEN** the UI is built with `REACT_APP_DORPH_UTILS_URL` set to a public HTTPS origin
- **THEN** the header does not show a DorphUtils chip
- **AND** activating DorphNotes opens the in-app modal
