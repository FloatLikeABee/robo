## 1. Schema and scan

- [x] 1.1 Add `plat_users.pool_alias` and `pool_share` / tag / topic / FTS / report tables. Seed a short topic list. morph-mcp does not migrate.
- [x] 1.2 Rule-based secret/PII scanner with tests. Secrets and emails block.

## 2. Share and pool tools

- [x] 2.1 `preview_share` / `share_note` / `unshare_note` / `list_my_shares` with confirm_token, alias, snapshot. Stick shareable. Other-user unshare is not found.
- [x] 2.2 `pool_topics` / `pool_search` / `pool_get` / `pool_save_copy` / `pool_report`. Untrusted envelope. Three reports hide. Rate limits. Search/get leak no email.

## 3. Docs and check

- [x] 3.1 Remove phase 1B markers from `docs/agents/16-morph-for-agents.md` and `morph/mcp/agent_guide.md` (keep them identical). Update agentpack and handshake counts.
- [x] 3.2 `cd morph && go vet ./... && go test ./...`. Morph frontend tests still pass.
