## ADDED Requirements

### Requirement: Pool and share tools are advertised
`tools/list` MUST include `preview_share`, `share_note`, `unshare_note`, `list_my_shares`, `pool_topics`, `pool_search`, `pool_get`, `pool_save_copy`, and `pool_report`. `pool_topics`, `pool_search`, and `pool_get` MUST be marked read-only. Share writes MUST require `pool:publish`. Pool reads MUST require `pool:read`. `pool_save_copy` MUST require `pool:read` and `notes:write`.

#### Scenario: Handshake lists pool tools
- **WHEN** a client calls `tools/list` after initialize
- **THEN** the nine pool tools are present
- **AND** `pool_search` is marked read-only
