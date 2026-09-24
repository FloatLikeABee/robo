# Design

## Context

See proposal.md for why. The root `Dockerfile` already has `ARG REACT_APP_MORPH_UTILS_URL` and `ENV` before `npm run build`. `morph/frontend/src/lib/headerAppLinks.js` returns an empty base in production when the URL is empty or loopback (`localhost`, `127.*`, `::1`). `SkoolAiChat.js` reads `process.env.REACT_APP_MORPH_UTILS_URL` at the call site, which is what CRA inlines. `render.yaml` only comments that the arg is optional. It does not declare the env key, so Render has nothing to pass into the Docker build. `deploy/check-stack-runbook.sh`, `formx/deploy/check-container-contract.sh`, and `composerx/deploy/check-container-contract.sh` reject `key: REACT_APP_MORPH_UTILS_URL` anywhere in the Blueprint. Those checks were written when story #106 had not wired the prompt yet.

## Goals / Non-Goals

**Goals:**

- The `morph` service names `REACT_APP_MORPH_UTILS_URL` as a dashboard prompt with no value.
- A production UI build with a public HTTPS URL inlines that URL, so the existing header gate shows the MorphUtils chip.
- A production UI build with the variable unset still omits the chip.
- The product owner runbook says to fill the prompt with `https://<morph-utils public host>` and rebuild.

**Non-Goals:**

- Calling Render, creating services, or putting `https://morph-utils.onrender.com` in source.
- A runtime config file, embed `VITE_*` keys (#114), or a Quill redesign.
- Changing the header gate's loopback rules.

## Decisions

### 1. Dashboard prompt on `morph`, no value

**Choice:** Add this to the `morph` env list and nowhere else:

```yaml
- key: REACT_APP_MORPH_UTILS_URL
  sync: false
```

Render passes service env vars into the Docker build. `sync: false` keeps a dashboard value across Blueprint syncs and stores nothing in git. An empty prompt inlines an empty string. The header gate already treats that as omitted.

**Rejected: hardcode `https://morph-utils.onrender.com` as the Dockerfile default.** That is an invented production host. Local image builds would open production, and a renamed service would require a code change. The issue forbids a required production host in app code.

**Rejected: a MorphUtils-style runtime `/config.js`.** A restart could then change the chip, but PR #93 and the stack runbook already treat this as a CRA build arg. A second config channel is more code and a different failure mode (the shell iframe bug when an origin is empty). Out of scope.

**Rejected: leave the key out of `render.yaml` and only repeat the #113 dashboard instructions.** That runbook already tells the product owner to set the variable by hand. The chip is still absent because the Blueprint does not declare the arg, and a later sync can drop a dashboard-only variable. A comment is not an env key Render will pass as a build arg.

**Rejected: bake the URL in the GitHub Actions image build.** Render builds the Dockerfile from the service env. The CI image is not the production service. Putting a host in CI would also invent a default.

### 2. Keep the header helper; prove inlining on the bundle

**Choice:** Do not change `morphUtilsBaseURL` unless a bundle check shows CRA failed to inline the call site. Add `morph/frontend/scripts/check-morph-utils-bundle.js`. The Morph frontend CI job builds once with the variable unset and once with `https://utils.example.com` (a fixture, not a production host). The unset build must not contain that fixture or `https://morph-utils.onrender.com`, and must not leave `process.env.REACT_APP_MORPH_UTILS_URL` in the bundle. The set build must contain the fixture. Existing unit tests stay the loopback regression.

**Rejected: unit tests only.** They call the helper with arguments. They do not prove CRA inlined the env var. If the call site stops reading `process.env.REACT_APP_MORPH_UTILS_URL` directly, production omits the chip while the unit tests stay green.

### 3. Compose passes the arg from the shell, default empty

**Choice:** `deploy/docker-compose.yml` sets `build.args.REACT_APP_MORPH_UTILS_URL` to `${REACT_APP_MORPH_UTILS_URL:-}`. Compose interpolation does not read `env_file`. The gitignored `deploy/.env.production` is runtime-only, and the example file must not gain a host. Document `docker build --build-arg REACT_APP_MORPH_UTILS_URL=https://<morph-utils public host>`.

**Rejected: reading the URL from `deploy/.env.production`.** That file is not the compose interpolation file. Wiring it would look like a runtime setting for a value CRA already inlined.

### 4. Narrow the "key must not appear" checks

**Choice:** `deploy/check-container-contract.sh` requires the `morph` prompt with `sync: false` and no `value:`. `deploy/check-stack-runbook.sh` requires that shape and still rejects example hosts and a value on the key. Formx, Content Maker, and MorphUtils checks reject the key on their own service, not on `morph`.

**Rejected: leave the whole-file forbid in place.** The prompt could not land, so the production build could never receive the arg.

### Grill

Proposer: declare the prompt. Reviewer: #113 rejected the key in `render.yaml`. Reply: #113 rejected a committed host and a key before this story. The requirement that remains is "no value in `render.yaml`". The prompt has no value. Reviewer: an empty first sync still ships a chip-less bundle. Reply: that matches the unset regression, and the runbook says to fill the prompt, then rebuild, and to clear the build cache if the chip is still missing. Reviewer: a loopback pasted in the dashboard would show localhost. Reply: the existing gate returns empty for loopback, so the chip stays omitted. Reviewer: two production builds slow CI. Reply: the Morph frontend job timeout is 30 minutes, and the second build is the only check that fails if inlining breaks. The Docker image job stays one unset build so CI does not bake a host.

## Risks / Trade-offs

- [First Blueprint sync creates an empty prompt] → Chip stays omitted until the product owner fills `https://<morph-utils public host>` and rebuilds. Same as today until that rebuild.
- [Restart without rebuild] → Bundle is unchanged. Runbook keeps that sentence.
- [Loopback or empty dashboard value] → Header gate omits the chip. Runbook says an empty or loopback value omits it.
- [Key placed on `morph-utils`] → The shell does not read `REACT_APP_*`. Contract checks keep the key off every service except `morph`.
- [CRA minify hides the call] → The check looks for the URL string and for a leftover `process.env.REACT_APP_MORPH_UTILS_URL`, not for a function name.

## Migration Plan

After merge, the product owner syncs the Blueprint, sets `REACT_APP_MORPH_UTILS_URL` on `morph` to the MorphUtils origin with no path, and deploys so the image rebuilds. If the chip is still missing, clear the build cache and deploy again. Rollback is clearing that dashboard value and rebuilding; the chip is omitted again. This repository does not call Render.

## Open Questions

None. The public origin stays in the dashboard. Docs use `https://<morph-utils public host>` and label `https://morph-utils.onrender.com` as an example that already exists.
