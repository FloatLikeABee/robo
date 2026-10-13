# note-pool Specification

## Purpose

Lets a signed-in agent publish a consented snapshot of an owned note into a Morph-only pool and lets other agents find and copy those snapshots without seeing the owner's email.

## Requirements

### Requirement: Share requires preview then confirm
`preview_share` MUST load the caller's note, scan it, and return a snapshot plus a `confirm_token` that expires in about five minutes. `share_note` MUST refuse a missing, expired, or foreign token with `consent_required` and MUST NOT insert a row. A successful share MUST store a snapshot so later private edits do not leak.

#### Scenario: Confirm token required
- **WHEN** the caller invokes `share_note` without a live `confirm_token` from `preview_share`
- **THEN** the result is `consent_required`
- **AND** no `pool_share` row is inserted

#### Scenario: Preview then share
- **WHEN** the owner previews a clean note and shares with that token
- **THEN** a live share exists
- **AND** `list_my_shares` includes it for that owner only

### Requirement: Scanner blocks secrets and personal data
`share_note` MUST refuse a snapshot that contains a secret or personal-data finding. The error MUST name the block. Preview MUST still return the findings.

#### Scenario: Secret blocks publish
- **WHEN** the note body contains an API key or JWT
- **THEN** `share_note` fails
- **AND** no live share is stored

#### Scenario: Email blocks publish
- **WHEN** the note body contains an email address
- **THEN** `share_note` fails
- **AND** no live share is stored

### Requirement: Alias only, never email
Share records, pool search hits, and `pool_get` MUST include a public alias and MUST NOT include the owner's email or plat user id. Attribution MUST look like `Shared by @alias · License CC BY 4.0 · morph share <id>`.

#### Scenario: Search does not leak email
- **WHEN** user A shared a pool note
- **THEN** user B's `pool_search` and `pool_get` do not contain A's email
- **AND** they include A's alias

### Requirement: Pool index is MCP-only
`pool_topics` MUST list curated topics. `pool_search` MUST return live `visibility=pool` index rows only (title, summary, tags, topic, kind, author alias, license, share_id, published_at). `pool_get` MUST wrap content in `<shared_content untrusted="true">`. A revoked or hidden share MUST be not found. There MUST be no public HTTP pool page in this change.

#### Scenario: Revoked share disappears
- **WHEN** the owner unshares
- **THEN** `pool_get` and `pool_search` treat the id as not found

#### Scenario: Stick note may be shared
- **WHEN** the owner previews and shares a stick note
- **THEN** `pool_get` returns that snapshot for a reader with `pool:read`

### Requirement: Save copy keeps attribution
`pool_save_copy` MUST insert a private note for the caller whose body starts with the attribution line and names the source share id.

#### Scenario: Copy attributes the author
- **WHEN** a reader saves a copy of a live share
- **THEN** the new note is owned by the reader
- **AND** the body includes the author's alias and the share id

### Requirement: Reports hide and rate limits apply
Three distinct reporters MUST set the share `status` to `hidden`. A later `pool_get` is not found. Crossing a per-user action limit MUST return `rate_limited`.

#### Scenario: Three reports hide
- **WHEN** three different users report the same live share
- **THEN** `pool_get` is not found

### Requirement: Pool read is not ExposeRecord
Pool authorization MUST NOT call `ExposeRecord`. Another user's share id on owner tools (`unshare_note`) MUST be not found.

#### Scenario: Other user cannot unshare
- **WHEN** user B calls `unshare_note` with user A's share id
- **THEN** the result is not found
- **AND** A's share stays live
