## 1. Checker first

- [x] 1.1 Add `scripts/check-operator-product-docs.py` implementing the invariants in design.md decision 6, including `/redeem`, `/admin`, and a service-table row whose UI cell is `invite-signup-ui` and whose API cell is `—`
- [x] 1.2 Run the checker against the current docs and confirm it fails because Research, Invite Signup, and the live nav are missing

## 2. Operator docs

- [x] 2.1 Update `docs/agents/03-morph.md`: drawer nav table, settings/configuration redirect, Research publish paths and five-round new jobs, Invite Signup redeem/admin flows, agent workspace tabs, AI tools drawer, `/api/admin/users` kept as an API
- [x] 2.2 Update `morph/README.md` and `README.md` so MorphNotes names the five drawer modules and Invite Signup is a folder, URL, and `invite-signup-ui` service
- [x] 2.3 Update `docs/agents/00-architecture-overview.md` with the five modules and Invite Signup on the map, tech table, port map, and names table
- [x] 2.4 Update only the opening services sentence and the service table in `docs/agents/12-build-deploy.md` to include `invite-signup` / `invite-signup-ui` with an empty API and alias

## 3. Verify

- [x] 3.1 Re-run `scripts/check-operator-product-docs.py` and confirm it exits 0
- [x] 3.2 Confirm the `12-build-deploy.md` diff does not add a CI section or restart/stop behavior
