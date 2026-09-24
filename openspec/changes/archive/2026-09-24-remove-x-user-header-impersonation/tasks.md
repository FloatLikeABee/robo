# Tasks

## 1. JWT-only session

- [x] 1.1 Add failing tests in `morph/handlers/authz_middleware_test.go`: header-only `X-User-ID` + `X-User-Role: admin` is 401 on `POST /api/chat`, `GET /api/admin/users`, and `POST /api/tran/research`, and the research row count is unchanged; invalid bearer plus `X-User-ID` is 401 and the row count is unchanged; a valid employee JWT plus spoofed admin headers is 403 on `GET /api/admin/users`; a chat stub returns the token user id and role; `ResolveUserScope` on the same request ignores the headers. Run `go test ./handlers/ -count=1 -run 'TestHeaderOnly|TestInvalidBearer|TestSpoofed|TestResolveUserScope'` from `morph/` and confirm the new tests fail because the header fallback still authenticates.
- [x] 1.2 Implement `ResolveUserScope(*http.Request)` and remove the header fallback so those tests pass. Strip client `X-User-*` headers on `/api` requests and stamp `X-User-ID`, `X-User-Email`, and `X-User-Role` only from the resolved `plat_users` row. Re-run the same `go test` filter and confirm it passes.

## 2. Published path shape

- [x] 2.1 Add a table test for `isPublicMorphRead` and HTTP checks covering `%2F`, trailing slash, `//`, dot segments, kind case, and a query string on a real published slug. Run the new test and confirm the loose shapes (`research//slug`, `research/slug/`) still pass the helper before the fix.
- [x] 2.2 Require exactly `{kind}/{slug}` with no empty or dot segments. Re-run the path test and confirm only the exact slug and the query-string case stay public.

## 3. Management tool loop

- [x] 3.1 Change `TestManagementAPIMutationUsesCallerJWT` so a handler mounted without `AuthzMiddleware` requires the inner `Authorization` to equal the outer value and requires `X-User-ID` to be absent. Add the no-Authorization case that returns 401. Run it and confirm it fails while the executor still sets `X-User-ID`.
- [x] 3.2 Stop `execManagementAPI` from setting `X-User-ID`, `X-User-Role`, `X-User-Roles`, and `X-User-Permissions`. Keep forwarding `Authorization`. Re-run the management test and the existing Research create-through-middleware case and confirm both pass.

## 4. OPTIONS and CORS

- [x] 4.1 Add a test that `OPTIONS /api/chat` with no session is not 401, and a `package main` test that an `OPTIONS` response sets `Access-Control-Allow-Headers` to a list containing `Authorization`, `Content-Type`, and `Accept`, not `*`, and not `X-User-ID`. Run them and confirm the CORS assertion fails on `*`.
- [x] 4.2 Pass `OPTIONS` through `AuthzMiddleware` before the session check. In the CORS block of `morph/main.go` only, set the explicit allow-list. Do not edit `morph/auth/jwt.go` or `morph/config`. Re-run the OPTIONS and CORS tests and confirm they pass.

## 5. Dev docs

- [x] 5.1 Update `docs/agents/01-auth-flow.md`, `docs/agents/03-morph.md`, and `morph/README.md` so they say identity headers are not a session, chat and admin require a Morph JWT, and there is no header fallback. Adjust chat swagger comments that present `X-User-ID` as a client input. Verify by reading those sections: no sentence still says `X-User-ID` satisfies the session check.

## 6. Module verification

- [x] 6.1 Run `go test ./...` in `morph/` and confirm the package tests pass. No frontend or Rust tree is modified, so no production frontend build or `cargo test` is required for this change.
