# Spec Delta

## Purpose

Makes the Morph header MorphUtils chip open the hosted shell when a production build is given a public MorphUtils URL, and keeps that chip hidden when the URL is missing or loopback.

## ADDED Requirements

### Requirement: A public MorphUtils URL is inlined into the production bundle
When the Morph production UI is built with `REACT_APP_MORPH_UTILS_URL` set to a non-empty URL that is not a loopback address, the built browser bundle MUST contain that URL. The MorphUtils header chip MUST be eligible to render and MUST use that URL as its base. Application source MUST NOT default that variable to a production host when it is unset.

#### Scenario: Production build with a public URL
- **WHEN** the production UI is built with `REACT_APP_MORPH_UTILS_URL` set to a public HTTPS origin
- **THEN** the built JavaScript contains that origin
- **AND** the header treats that origin as eligible for the MorphUtils chip

### Requirement: An unset or loopback URL omits the chip
When the Morph production UI is built with `REACT_APP_MORPH_UTILS_URL` unset, empty, or set to a loopback URL (`localhost`, `127.0.0.0/8`, or `::1`), the MorphUtils header chip MUST stay omitted. A development build MAY still default the chip to `http://localhost:3040`.

#### Scenario: Unset production build
- **WHEN** the production UI is built without `REACT_APP_MORPH_UTILS_URL`
- **THEN** the built JavaScript does not contain a production MorphUtils origin
- **AND** the header treats the MorphUtils base URL as empty

#### Scenario: Loopback production value
- **WHEN** the production header is given `http://localhost:3040` as the MorphUtils URL
- **THEN** the MorphUtils base URL is empty
