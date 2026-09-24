## RENAMED Requirements

- FROM: `### Requirement: One Docker web service in a flat Blueprint`
- TO: `### Requirement: Morph Docker web service in a flat Blueprint`

## MODIFIED Requirements

### Requirement: Morph Docker web service in a flat Blueprint
The repository MUST contain `render.yaml` at the repo root. The file MUST declare a top-level `services` list and MUST NOT declare a `projects` block. That list MUST contain a service with `type: web`, `runtime: docker`, `name: morph`, `region: singapore`, `plan: starter`, and `branch: main`. The service MUST build `dockerfilePath: ./Dockerfile` with `dockerContext: .`. It MUST set `healthCheckPath: /health` and `autoDeployTrigger: checksPass`. Other services MAY appear in the same list. The `morph` fields above stay on the `morph` service.

#### Scenario: Blueprint matches the Morph image
- **WHEN** a reviewer reads the `morph` service in `render.yaml`
- **THEN** it is a Docker web service named `morph` in `singapore` on the `starter` plan, tracking `main`, built from the root Dockerfile and context
- **AND** its health check path is `/health` and it auto-deploys only after checks pass
