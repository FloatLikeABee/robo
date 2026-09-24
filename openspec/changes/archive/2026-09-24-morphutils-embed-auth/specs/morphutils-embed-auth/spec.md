## Purpose

Makes production MorphUtils load Event Logs, Content Maker, Data Access, and Project from configured HTTPS origins under one Morph sign-in, and sends an unauthenticated visitor back to Morph.

## ADDED Requirements

### Requirement: Configured embed iframes use the live origin
When MorphUtils is configured, the Event Logs, Content Maker, Data Access, and Project iframe `src` values MUST use those origins. Event Logs MUST keep the `/events-info` path on its origin. A production shell whose origin for a module is blank MUST NOT mount that iframe and MUST NOT substitute a loopback host. Local `npm run dev` with the variables unset MUST keep the current localhost defaults.

#### Scenario: HTTPS origins are the iframe sources
- **WHEN** Event Logs, Content Maker, Data Access, and Project origins are non-loopback `https` URLs
- **THEN** the Event Logs iframe `src` is that origin plus `/events-info`
- **AND** the other three iframe `src` values are those origins
- **AND** none of those `src` values use `localhost` or `127.0.0.1`

#### Scenario: Blank production origin does not iframe the shell
- **WHEN** the production shell has a blank origin for a module
- **THEN** that module has no iframe `src`
- **AND** the shell does not request `/events-info` on its own origin

### Requirement: A Morph session is handed to each embed as a bearer
A visitor who is signed in on Morph and then opens MorphUtils MUST be signed in on an embed without a second signup. The shell MUST pass the Morph bearer on the iframe `src` as `userspanel_token`. The shell MUST keep that bearer in a host-only cookie named `userspanel_session_token` with `SameSite=Lax` and MUST NOT set a `Domain` attribute. The shell MUST remove `userspanel_token` from its own address after storing it.

#### Scenario: Signed-in shell passes the bearer
- **WHEN** the shell holds a Morph bearer and an embed origin is configured
- **THEN** that embed iframe `src` includes `userspanel_token`
- **AND** the shell cookie is host-only with `SameSite=Lax` and no `Domain`

#### Scenario: Parent-domain cookies are not the production strategy
- **WHEN** an operator reads how a session crosses MorphUtils and an embed on `*.onrender.com`
- **THEN** the documented strategy is the bearer query handoff
- **AND** the notes say a cookie `Domain` of `onrender.com` is not set because that site is a public suffix

### Requirement: An unauthenticated visitor gets a path back to Morph
When MorphUtils has no Morph bearer, it MUST show a sign-in link to the configured Morph origin and MUST NOT mount embed iframes. `VITE_MORPH_AI_URL` is that origin when set. Otherwise `VITE_MORPH_API_URL` is that origin. When both are blank, the shell MUST say the Morph origin is not set and MUST NOT link to a loopback host. An embed login link built for production MUST NOT use `localhost` or `127.0.0.1`.

#### Scenario: Shell without a session links to Morph
- **WHEN** MorphUtils has no bearer and `VITE_MORPH_API_URL` is `https://morph.example`
- **THEN** the shell shows a sign-in link to `https://morph.example`
- **AND** it does not mount Event Logs, Content Maker, Data Access, or Project iframes

#### Scenario: Missing Morph origin is not a loopback link
- **WHEN** MorphUtils has no bearer and both Morph origin variables are blank
- **THEN** the shell does not link to `localhost` or `127.0.0.1`

#### Scenario: Production embed login is not localhost
- **WHEN** an embed builds its Morph sign-in link for a production page and no Morph origin is configured
- **THEN** that link does not use `localhost` or `127.0.0.1`

### Requirement: Docs list embed env and cookie failure modes
`deploy/README.md` MUST list `VITE_SHEETX_URL` (alias `VITE_FORMSX_URL`), `VITE_COMPOSERX_URL`, `VITE_DATAX_URL`, `VITE_PROJECTS_URL` (alias `VITE_MORPH_ENGI_URL`), and `VITE_MORPH_AI_URL` as MorphUtils embed origins read when the container starts. It MUST name the cookie `userspanel_session_token`, `SameSite=Lax`, and that no `Domain` is set. It MUST state these failure modes: `*.onrender.com` hosts are separate sites so a parent cookie is not sent; `SameSite=Lax` cookies are not sent on a cross-site iframe load; the bearer in the first request can appear in that request's logs before the page strips it; signing out on one host does not clear the cookie on the others; opening an embed origin directly does not see the Morph host cookie. It MUST say `bk` and invite-signup are not required. Example `onrender.com` hosts MUST stay labeled as examples and MUST NOT be required defaults in application source.

#### Scenario: A reader can configure origins and auth
- **WHEN** a reader follows the MorphUtils embed section in `deploy/README.md`
- **THEN** they see the embed origin variables, the cookie name, `SameSite=Lax`, and the public-suffix failure mode
- **AND** they see that `bk` and invite-signup are not required
