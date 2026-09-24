## 1. Shell URL and session helpers

- [x] 1.1 Add failing tests for HTTPS embed `src` values, a blank production origin, bearer query handoff, host-only `SameSite=Lax` cookie attributes, and a Morph sign-in href that ignores a blank or loopback production origin
- [x] 1.2 Implement those helpers in the MorphUtils shell and keep localhost defaults on dev only

## 2. Unauthenticated shell

- [x] 2.1 Show the Morph sign-in link and do not mount embed iframes until a bearer is present
- [x] 2.2 Keep the local `./start-all.sh` hint for a loopback embed miss, and use a remote unreachable hint otherwise

## 3. Embed sign-in links

- [x] 3.1 Stop Data Access and Project from using `localhost` as the production Morph sign-in link when no origin is configured

## 4. Blueprint prompts and runbook

- [x] 4.1 Add the embed origin prompts on `morph-utils` only, with `sync: false` and no value, and update the contract checks that forbade those keys
- [x] 4.2 Document the env matrix, cookie name, `SameSite=Lax`, no `Domain`, the `*.onrender.com` failure modes, and that `bk` and invite-signup are not required

## 5. Verify

- [x] 5.1 Run the MorphUtils frontend tests and production build, the touched container contract checks, and `go test` / `cargo test` only if those modules changed
