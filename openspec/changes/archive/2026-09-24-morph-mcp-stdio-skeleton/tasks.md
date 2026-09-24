## 1. Module and command

- [x] 1.1 Pin `github.com/modelcontextprotocol/go-sdk` v1.8.0 in `morph/go.mod` and keep the Morph module building
- [x] 1.2 Add `morph/mcp` plus `morph/cmd/morph-mcp` so the process speaks stdio and does not listen on a port

## 2. Identity

- [x] 2.1 Add failing tests for a missing token, an invalid token that must not be echoed, an expired token, a token with no subject, and an empty `JWT_SECRET`
- [x] 2.2 Verify `MORPH_MCP_TOKEN` with `auth.DecodeToken` and refuse to start the stdio process when verification fails

## 3. Handshake

- [x] 3.1 Test `initialize` at protocol 2025-06-18 for serverInfo plus tools and resources capabilities, and that a newer SDK revision can still list tools
- [x] 3.2 Test that `tools/list` returns only read-only `whoami`, that `whoami` returns the verified user without the token, and that `resources/list` returns no resources
- [x] 3.3 Advertise the resources capability with an empty resource list and register `whoami` on the shared server

## 4. Stores and stdout

- [x] 4.1 Test that `morph/mcp` and `morph/cmd/morph-mcp` do not import `idongivaflyinfa/db` or Badger
- [x] 4.2 Test the stdio process: initialize, tools/list, and whoami on stdout as JSON-RPC; identity failures leave stdout empty; logs and the token stay off the wrong streams
- [x] 4.3 Confirm a running stdio session does not hold a listening TCP socket

## 5. Docs

- [x] 5.1 Document the build command, `MORPH_MCP_TOKEN` and `JWT_SECRET`, a Cursor `mcp.json` example, that HTTP tool catalogs are not MCP, and the Badger and identity decisions

## 6. Verification

- [x] 6.1 Run `go build ./cmd/morph-mcp` and `go test ./...` in `morph/`
