## ADDED Requirements

### Requirement: Startup can read SQL migrations from the working directory
The image MUST contain the Data Access SQL migrations at `migrations` relative to the process working directory `/app` (`/app/migrations`). Those files MUST be the SQL files from `SharpReport/backend/migrations`. The container contract MUST fail when the image does not copy that directory to `/app/migrations`, or when that source directory has no `.sql` file. Startup MUST keep reading that relative directory. The contract MUST NOT require a value for `USERS_PANEL_BASE_URL`.

#### Scenario: Image includes the migrations directory
- **WHEN** a reviewer reads the Data Access Dockerfile and `SharpReport/backend/migrations`
- **THEN** the Dockerfile copies that directory to `/app/migrations`
- **AND** the source directory contains at least one `.sql` file
- **AND** the container contract fails if either of those is missing
