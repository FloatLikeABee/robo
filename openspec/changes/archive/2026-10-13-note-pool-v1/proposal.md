## Why

Bridge A landed PATs and owner-scoped notes. Agents still cannot publish a snapshot into a shared pool or find other users' notes through MCP. The playbook still labels those tools "phase 1B".

## What Changes

- Sharing tools: `preview_share`, `share_note`, `unshare_note`, `list_my_shares`. A share needs a preview and a `confirm_token`. No token is `consent_required`.
- Shared notes show a public alias, never the owner's email or user id.
- Pool index tools: `pool_topics`, `pool_search`, `pool_get`, `pool_save_copy`. Saved copies keep attribution to the original alias and source share id.
- A rule-based secret/PII scanner blocks a share when it finds secrets or personal data.
- In-process rate limits. `pool_report` auto-hides a share after three distinct reporters.
- No pool web page. Stick notes may be shared. Anyone with `pool:read` may `pool_get` a live share id.
- `docs/agents/16-dorph-for-agents.md` and the embedded `agent_guide.md` drop "phase 1B" markers.

Out of scope: subscribe/feed/mute, `update_share`, admin queue, MCP Apps share/pool UI, HTTP/OAuth, LLM classifier.

## Capabilities

### New Capabilities

- `note-pool`: Snapshot shares, consent, alias attribution, MCP-only index, scanner, reports, rate limits.

### Modified Capabilities

- `dorph-mcp-stdio`: Advertise the nine pool tools. Do not reuse `ExposeRecord` for pool reads.
- `dorph-for-agents`: Pool tools are current, not "coming in phase 1B". Stick notes may be shared.

## Impact

- Schema: `pool_*` tables, `plat_users.pool_alias`. dorph-mcp does not migrate.
- `dorph/mcp` share/pool tools. Scanner and in-process limiter.
- Docs: `16-dorph-for-agents.md` and `dorph/mcp/agent_guide.md` (must stay identical).
