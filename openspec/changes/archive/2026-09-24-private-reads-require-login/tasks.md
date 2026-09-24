## 1. Private reads and published visibility

- [x] 1.1 Add failing tests: anonymous `GET`/`HEAD` on private tran, forms, knowledge, graph, personal-data lists, and a download path return 401; a valid JWT `GET` is not 401; published research HTML stays 200 without a session; an unpublished research row is not returned by slug.
- [x] 1.2 Add `publish.PageRoute` and `publish.Visible`, and `mcp.ExposeRecord`. Tests fail while `Visible("")` is unimplemented or while the middleware still allows anonymous private reads.
- [x] 1.3 Require a session for every `isMorphDataAPI` method except `publish.PageRoute`. Public handlers refuse a slug that is not `Visible`. `mcp.ExposeRecord` uses `Visible` when the caller has no user id.

## 2. Graph health, JWT iat, chat fail-closed

- [x] 2.1 Add a failing test that `GraphHealth` JSON has no Neo4j URI and no raw error field, and a failing production test that a future `iat` beyond one minute is rejected while an `iat` inside the leeway is accepted.
- [x] 2.2 Stop `GraphHealth` from ensuring schema or returning `neo4j_uri` / `neo4j_error`. Production `DecodeToken` uses `jwt.WithIssuedAt` and a one-minute leeway.
- [x] 2.3 Chat and chat-session handlers return 401 when the user id is empty. Remove unused `IsAdminRoles`. Test the chat handler.

## 3. MorphNotes login return

- [x] 3.1 Add `safeReturnPath` and a unit test that keeps `/morphdata/research` and rejects `https://evil.example/phish` and `//evil.example`.
- [x] 3.2 Wrap `/morphdata` in `ProtectedLayout`. Send 401s from `tranClient`, including on Morph Data, to `/login?returnTo=`. `LoginPage` navigates to the validated path.

## 4. Docs and verification

- [x] 4.1 Update `docs/agents/01-auth-flow.md`, `docs/agents/03-morph.md`, `morph/README.md`, `docs/agents/14-morph-mcp.md`, and the hosting checklist so the public route list matches this change.
- [x] 4.2 Run `cd morph && go vet ./... && go test ./...`, the return-path unit test, and the Morph frontend production build.
