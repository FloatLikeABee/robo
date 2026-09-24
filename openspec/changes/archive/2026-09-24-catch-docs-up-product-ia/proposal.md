## Why

Operator and agent docs still describe the previous MorphNotes information architecture. They send people to Settings → Users, omit the Research nav item and its publish path, and omit Invite Signup (`invite-signup/`, UI on port 3051, `invite-signup-ui` in `start-all.sh`). Agents planning from those pages will rebuild a Users screen that the code removed and will miss the apps that replaced it.

## What Changes

- Update `README.md`, `docs/agents/00-architecture-overview.md`, `docs/agents/03-morph.md`, `docs/agents/12-build-deploy.md`, and `morph/README.md` so they match the code on main.
- Document MorphNotes nav as Tasks, Timelines, Big notes, Research, and Generic data. Do not present Settings → Users as a primary destination.
- Document Invite Signup: folder `invite-signup/`, dev UI `http://localhost:3051`, launcher name `invite-signup-ui`, redeem (`/redeem`) and admin (`/admin`) flows, backed by Morph API invite routes.
- Document the Research module (`/morphdata/research`) and publish paths (`POST /api/tran/research/:id/publish`, public HTML `GET /api/tran/public/research/:slug`).
- Add `invite-signup-ui` to the service table in `docs/agents/12-build-deploy.md` and to the matching README service and URL tables. Keep that file's other sections untouched so open PRs that add CI and launcher stop/restart notes can land cleanly.
- State the Morph AI agent workspace tabs that exist today (Notes & TODOs, Context & Knowledge). Do not describe a Files workspace tab.

## Capabilities

### New Capabilities

- `operator-product-docs`: The listed operator docs state the live MorphNotes nav, Research publish paths, and Invite Signup service, port, and flows.

### Modified Capabilities

- (none — `openspec/specs/` has no existing capability for these docs)

## Impact

- Docs only: the five files above. No application code, routes, or launcher behavior.
- Cites implemented work in `openspec/changes/morphutils-user-profile-invite-app/`, `openspec/changes/morphnotes-research-thesis-synthesis/`, and `openspec/changes/morphai-drop-files-workspace/` without rewriting those proposals.
- Out of scope: other agent docs (including `docs/agents/02-ai-integration.md`), marketing copy, and the CI / process-group sections owned by other pull requests.
