## Context

See proposal.md for why the docs are wrong. The facts below were read from main (`d862013`), not from other markdown or from unarchived OpenSpec proposals.

Live MorphNotes drawer (`morph/frontend/src/components/admin/AppDrawer.js`) is five buttons: Tasks, Timelines, Big notes, Research, Generic data. Routes are in `morph/frontend/src/appRouter.js` under `ADMIN_BASE_PATH` (`/morphdata` in `adminPaths.js`). `settings` and `configuration` redirect to `/morphdata/generic-data`. There is no `configuration/users` route and no `UsersAdmin` page. The catch-all `*` also redirects to generic data. `/morphdata/places` still renders a page and is not in the drawer; this change does not add it to the nav list.

Invite Signup is `invite-signup/frontend` (Vite, `server.port` 3051, `/api` proxied to `127.0.0.1:9090`). `start-all.sh` lists `invite-signup-ui` in `ALL_SERVICES`, starts `npm run dev` in that frontend, and `service_url` prints `http://localhost:3051`. `resolve_services` has no invite alias. UI routes are `/`, `/redeem`, `/admin`. Redeem calls `POST /api/invite/redeem`. Admin calls `POST /api/auth/login`, `GET /api/auth/me`, `GET/POST /api/admin/invite-codes`.

Research UI is `/morphdata/research`. New jobs use five rounds (`researchRoundCount = 5` in `morph/handlers/tran_research.go`; the page subtitle says five). Optional files are `.txt`, `.pdf`, `.csv`, `.json`. Publish is `POST /api/tran/research/:id/publish`, which stores `published_path` as `/api/tran/public/research/<slug>`. Public HTML is `GET /api/tran/public/research/:slug`. In dev, `morph/frontend/src/setupProxy.js` proxies `/api` (including HTML) to port 9090, so a link opened from the Research page on `:3031` reaches that handler.

Morph AI agent workspace tabs in `AgentWorkspace.js` are Notes & TODOs and Context & Knowledge. A stored tab id `files` is read back as `knowledge`. `AgentFilesTab.js` and `filesWorkspaceStore.js` are gone. The header **AI tools** button still opens `AiToolsWorkspaceDrawer` (`bk`). That drawer is not the removed Files tab.

`/api/admin/users` CRUD is still registered. MorphUtils still has a **Your account** modal for the signed-in user (`PATCH /api/auth/me`). Bootstrap admin (`EnsureBootstrapAdmin`, README `morphadmin` / `admin123`) is unchanged. Invite Signup is how new accounts are provisioned in the product UI.

Open PRs also edit `docs/agents/12-build-deploy.md`: #61 adds a `## CI` section, #72 adds two lines about `start-all.sh` restart/stop. This change must not write those topics.

## Goals / Non-Goals

**Goals:**

- Make the five listed docs agree with the drawer, router, invite app, and launcher.
- Put the long Research and Invite Signup explanation in `docs/agents/03-morph.md`, and put only the service name, port, and folder in the README, architecture map, and build-deploy table.
- Leave a checker that fails if those claims regress or if a new relative link or repo path in those files does not resolve.

**Non-Goals:**

- Editing `docs/agents/02-ai-integration.md` (it still says "Files workspace tabs"). Issue #12 does not list it.
- Rewriting unarchived OpenSpec proposals, including the Research spec that still says twenty rounds.
- Changing `start-all.sh`, routes, or any app.
- Adding a CI section, process-group stop notes, or per-app build lines to `12-build-deploy.md`.
- Documenting `/morphdata/places` as a nav item.
- Removing or hiding `/api/admin/users` in the API table.

## Decisions

### 1. Code wins over older specs and proposals

Document five research rounds, port **3051**, and no Files tab, because that is what the code does.

Rejected: copy `openspec/changes/morphnotes-research-md-utils-polish/` (twenty rounds). The later thesis change and `researchRoundCount` are five. An older doc would send operators to wait for twenty rounds.

Rejected: copy the invite proposal's "~3050". Port 3050 is the usual Data Access API (`SHARPREPORT_PORT`). Vite and `service_url` are 3051.

Rejected: delete the phrase "AI tools" because the Files workspace was removed. The Files tab is gone; the AI tools drawer is still in `SkoolAiChat.js`. Describe them with those two names.

### 2. One deep page, short pointers elsewhere

`docs/agents/03-morph.md` owns the nav table, the redirect note for `/morphdata/settings` and `/morphdata/configuration`, Research (new jobs: five rounds; file types; publish; public GET), Invite Signup flows and API routes, the workspace tabs, the `npm run build` command for that frontend, and the note that `/api/admin/users` remains an API while the Users page does not. Launcher examples use the service name `invite-signup-ui`, not `invite-signup` (`resolve_services` does not alias it).

`morph/README.md` lists the five modules in drawer order and points at `03-morph.md` and `http://localhost:3051`. The root README folder cell for `morph/` names those five modules, and the URL and service tables gain Invite Signup. `00-architecture-overview.md` names the same five modules on the Morph row, adds Invite Signup to the diagram, tech table, port map, and user-facing names. Neither file repeats the redeem walkthrough; the spec's "search the listed docs" scenarios are satisfied by `03-morph.md` plus the service/URL rows.

Rejected: paste the same Research and redeem sections into all five files. The next IA change would have to be edited in five places and the copies would drift.

Rejected: put the redeem walkthrough only in a new `invite-signup/README.md`. The issue names the existing five files, and a sixth file is a new surface this story did not ask for.

Rejected: leave the five module names only in `03-morph.md`. The spec requires `README.md` and `00-architecture-overview.md` to describe that nav too, or a reader of the root page still sees an unnamed MorphNotes.

### 3. `12-build-deploy.md` gets two hunks

Change the opening "Remaining services" sentence so it includes `invite-signup`, and add one service-table row: API `—`, UI `` `invite-signup-ui` ``, alias `—`. Do not touch prerequisites, quick start, environment, per-app build, or production.

Rejected: also add `cd invite-signup/frontend && npm run build` to the per-app build list. That is a third region of the file and is the kind of edit that collides with the open CI and launcher PRs. The same command is one sentence in `03-morph.md`. Day-to-day start is `./start-all.sh start invite-signup-ui` (exact service name).

Rejected: generate the service table from `start-all.sh`. That would restructure a conflict-prone file for one missing row.

### 4. Say what the drawer shows, not every leftover route

The nav table matches `AppDrawer.js` order: Tasks, Timelines, Big notes, Research, Generic data. A sentence states that `/morphdata/settings` and `/morphdata/configuration` redirect to generic data, so the old Users URL is not a destination. Do not list `/morphdata/places` or the unused `platformUiDefaults.js` keys (`nav_user_settings`, People, Places). Those defaults are not the drawer.

Rejected: document every `appRouter.js` child, including `places` and the legacy redirects. The acceptance criterion is the drawer. A full route dump would read as if Places and User settings were back in the product.

Rejected: write "the only MorphNotes routes are these five." That is false because `places` still mounts.

### 5. Provisioning, self-service, and the admin API stay distinct

New accounts: Invite Signup redeem. The signed-in user's own username and password: MorphUtils **Your account** (`PATCH /api/auth/me`). The first admin is still the bootstrap account in the root README. `03-morph.md` keeps `/api/admin/users` in the API table and labels it as an API with no MorphNotes page.

Rejected: delete the `/api/admin` row. The handlers are still registered. A doc that says user admin is gone would invite someone to remove a live API.

Rejected: tell readers to create users with `POST /api/admin/users`. The product path the issue asks for is Invite Signup.

### 6. A checker is the test

`scripts/check-operator-product-docs.py` is the regression test. It fails unless:

- The five docs do not contain `Settings → Users` or `/morphdata/configuration/users`.
- `README.md`, `docs/agents/00-architecture-overview.md`, `docs/agents/03-morph.md`, and `morph/README.md` each contain `Tasks`, `Timelines`, `Big notes`, `Research`, and `Generic data`.
- `docs/agents/03-morph.md` contains the five `/morphdata/` routes, `POST /api/invite/redeem`, `POST /api/admin/invite-codes`, `` `/redeem` ``, `` `/admin` `` (not merely the substring inside `/api/admin`), `POST /api/tran/research/:id/publish`, `GET /api/tran/public/research/:slug`, `Notes & TODOs`, `Context & Knowledge`, `A new job runs five verified online rounds`, and the sentence `` `/morphdata/settings` and `/morphdata/configuration` redirect to `/morphdata/generic-data` ``.
- `03-morph.md` does not contain `Files workspace`.
- README and `12-build-deploy.md` contain `invite-signup-ui` and the port number parsed from `invite-signup/frontend/vite.config.ts` (`port: 3051`). The same port string must appear in `start-all.sh` `service_url` for `invite-signup-ui`.
- The checker does not look for or forbid a `## CI` heading or restart/stop prose. Those topics are owned by other PRs; if they land on main, this check must still pass.
- Every relative markdown link in the five files resolves to a file or directory.
- Backtick tokens that look like tracked repo paths resolve. A token is a candidate only when it equals `start-all.sh` or has at least one slash, no spaces, and does not start with `/`, `./`, or `../`, and does not contain `=`, `<`, `>`, `*`, or `http`. Gitignored runtime and build paths (`.robo-dev/`, `**/dist`, `deploy/.env.production`) are not required to exist on a fresh checkout when the parent exists or is itself gitignored. If the parent is missing and not gitignored, the same paragraph must say the tree is absent or that the path applies only when that tree exists. A missing tracked path is allowed only when the same paragraph says it is absent.

Rejected: no automated check, only a manual read. The story's failure mode is exactly this kind of drift.

Rejected: snapshot the whole markdown files. Wording can change; the invariants cannot.

Rejected: treat every backtick span with a slash as a file path. That flags `POST /api/tran/research/:id/publish` and fails a correct doc.

## Risks / Trade-offs

- [PR #61 or #72 edits the same service table or opening sentence before this merges] → Keep the `12-build-deploy.md` diff to those two hunks and rebase on latest `main` immediately before opening the PR. The checker must not forbid headings those PRs add.
- [`docs/agents/02-ai-integration.md` still says Files workspace tabs] → Left out on purpose. Called out here so a later docs pass can fix it. This change does not cite that sentence as current.
- [A reader treats the five nav labels as the full route table and is surprised by `/morphdata/places`] → The nav section says these are the drawer items, and that settings/configuration redirect. It does not say no other route exists.
- [Checker false-fails on an API path, an env assignment, a runtime log path, or a shell command in backticks] → Relative `](...)` links are resolved. Repo-path checks skip tokens that start with `/`, `./`, or `../`, and tokens that contain spaces, `=`, `<`, `>`, `*`, or `http`. Gitignored outputs (`.robo-dev/`, `dist/`, `deploy/.env.production`) need a real parent directory, not a checked-in file. `start-all.sh` alone is checked because the docs name that file.
- [Documenting five rounds disagrees with the unarchived twenty-round spec] → The unarchived spec is stale relative to `tran_research.go`. This change does not rewrite it (out of scope). The operator doc follows the code.
- [Invite Signup dev server is useless without Morph API] → `03-morph.md` says the Vite app proxies `/api` to `:9090`, so `morph-api` must be up.

## Migration Plan

Docs-only. No data migration and no deploy step. Rollback is reverting the five markdown files and the checker.

## Open Questions

None. Port, nav, publish routes, and the `12-build-deploy.md` edit boundary are fixed by the code and by the coordination note on issue #12.
