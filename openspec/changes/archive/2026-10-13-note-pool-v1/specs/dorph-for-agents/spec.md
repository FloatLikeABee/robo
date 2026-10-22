## MODIFIED Requirements

### Requirement: Playbook documents every tool and error code
The playbook MUST include a worked request and response example for each current tool named in the agentic-notes plan. Implemented pool tools MUST NOT be labeled coming in phase 1B. The playbook MUST list `not_found`, `invalid_input`, `rate_limited`, `forbidden_scope`, and `consent_required` with what the agent should do.

#### Scenario: Pool tools are named and current
- **WHEN** an agent reads the tool reference
- **THEN** it finds `preview_share`, `share_note`, `unshare_note`, `list_my_shares`, `pool_topics`, `pool_search`, `pool_get`, `pool_save_copy`, and `pool_report`
- **AND** those tools are not marked coming in phase 1B

#### Scenario: Error codes have a next step
- **WHEN** an agent reads the errors section
- **THEN** each of `not_found`, `invalid_input`, `rate_limited`, `forbidden_scope`, and `consent_required` has a stated next action
