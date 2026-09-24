# morph-container Specification Delta

## MODIFIED Requirements

### Requirement: Operators have a short runbook
`deploy/README.md` MUST describe how to build, run, set the env file, mount `/data`, back up `/data`, and upgrade the image. `docs/agents/12-build-deploy.md` MUST point at that runbook. The runbook MUST document creating the Morph service on Render from the root Blueprint, and MUST NOT tell the operator to follow `scripts/deploy.sh`.

#### Scenario: Reader follows the pointer
- **WHEN** a reader opens the production section of `docs/agents/12-build-deploy.md`
- **THEN** it links to `deploy/README.md` for the Morph image and for the Render Blueprint

#### Scenario: Deploy script is not the runbook
- **WHEN** a reader follows `deploy/README.md` to host Morph
- **THEN** the steps create the service from `render.yaml`
- **AND** they do not run `scripts/deploy.sh`

### Requirement: Image build is checked without a new required check name
CI MUST build the image on pull requests and on pushes to `main`. That build MUST NOT add or rename any of the five required check names (`Go / Morph API`, `Go / Event Logs`, `Go / Content Maker`, `Go / morphai`, `Morph frontend`). The platform CI workflow MUST still report exactly those five names. The Morph image workflow MUST validate `render.yaml` against the Render Blueprint schema in a step of that same workflow.

#### Scenario: Image workflow is separate
- **WHEN** a pull request is opened
- **THEN** an image build runs
- **AND** the platform CI workflow still reports only the five existing check names

#### Scenario: Blueprint schema is part of the image workflow
- **WHEN** the Morph image workflow runs
- **THEN** it validates `render.yaml` against the Render Blueprint schema
- **AND** it does not add a job or check name to the platform CI workflow
