# Spec Delta

## MODIFIED Requirements

### Requirement: Production env is explicit and secrets are empty
The service env list MUST set `MORPH_ENV` to `production`, `PORT` to `9090`, `GIN_MODE` to `release`, and `MORPH_AI_PROVIDER` to `dashscope`. It MUST set `ADMIN_USERNAME` and `ADMIN_EMAIL` to plain values. It MUST set `DB_PATH`, `TRAN_SQLITE_PATH`, `ENTITY_DETAILS_BADGER`, `MORPH_KNOWLEDGE_DIR`, and `TRAN_ENTITY_ATTACHMENT_DIR` to the image defaults under `/data`. Every secret the Morph API can read, including `JWT_SECRET`, `ADMIN_PASSWORD`, and `MORPH_AI_API_KEY`, MUST appear as `sync: false` and MUST NOT have a `value`. The committed file MUST NOT contain a usable JWT secret, admin password, or API key. The `morph` service MUST list `REACT_APP_MORPH_UTILS_URL` with `sync: false` and MUST NOT give that key a `value`. The committed file MUST NOT contain a URL for that key.

#### Scenario: Port is pinned
- **WHEN** a reviewer reads the `PORT` entry
- **THEN** its value is `9090`

#### Scenario: Secrets are dashboard-only
- **WHEN** a reviewer reads `JWT_SECRET`, `ADMIN_PASSWORD`, and `MORPH_AI_API_KEY`
- **THEN** each has `sync: false` and no value

#### Scenario: MorphUtils URL is a dashboard prompt
- **WHEN** a reviewer reads `REACT_APP_MORPH_UTILS_URL` on `morph`
- **THEN** it has `sync: false` and no value
