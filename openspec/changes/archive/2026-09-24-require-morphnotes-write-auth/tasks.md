## 1. Tests for the session split

- [x] 1.1 Add Go tests that fail on the open carve-out: anonymous `POST`/`PATCH`/`PUT`/`DELETE` on Research (create, publish, and the other mutating methods) and anonymous `POST` on `/api/forms/templates`, `/api/knowledge/files`, and `/api/graph/search` return 401; anonymous `POST` under `/api/tran/public/research/:slug` and `GET /api/tran/public/other/:slug` return 401; an invalid bearer on Research create returns 401.
- [x] 1.2 Cover the allow paths in the same tests: anonymous `GET`/`HEAD` of the three published slug routes is not 401 (published research HTML is 200); anonymous `GET` of `/api/tran/research`, `/api/forms/templates`, `/api/knowledge/files`, and `/api/graph/health` is not 401; a valid Morph JWT creates and publishes Research and is not 401 on the sibling POSTs; `execManagementAPI` forwards that JWT and the inner create succeeds; `X-User-ID` without a JWT still creates Research.

## 2. Middleware

- [x] 2.1 Replace `isOpenMorphDataAPI`'s anonymous bypass with the design split: unsafe methods on `/api/tran/`, `/api/forms/`, `/api/knowledge/`, and `/api/graph/` require `resolveUserScope`; `GET` and `HEAD` on those prefixes do not, except `/api/tran/public/` which is only the explicit published-page allowlist. Leave `resolveUserScope` header fallback unchanged. Do not edit `.github/` or `docs/agents/12-build-deploy.md`.

## 3. Docs

- [x] 3.1 Document the public GET/HEAD routes and that writes on the four prefixes need a Morph JWT, including the remaining anonymous private reads, in `docs/agents/01-auth-flow.md`, `docs/agents/03-morph.md`, and `morph/README.md`.

## 4. Verification

- [x] 4.1 Run `cd morph && go vet ./... && go test ./...` and record a passing result.
