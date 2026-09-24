# Spec Delta

## ADDED Requirements

### Requirement: MorphUtils URL is a non-secret build argument
The Morph image build MUST accept `REACT_APP_MORPH_UTILS_URL` as a build argument that the UI build reads. That argument MUST NOT have a default production host. The committed Dockerfile, compose file, and production env example MUST NOT contain a production MorphUtils origin. Compose MUST pass the argument from the shell, and an empty value MUST be allowed so an unset build still omits the header chip.

#### Scenario: Unset build argument has no host
- **WHEN** a reviewer reads the Morph Dockerfile build argument `REACT_APP_MORPH_UTILS_URL`
- **THEN** it has no default production host

#### Scenario: Compose does not bake a host
- **WHEN** a reviewer reads the Morph service build in `deploy/docker-compose.yml`
- **THEN** `REACT_APP_MORPH_UTILS_URL` is a build argument
- **AND** the file does not hardcode a production hostname
