## MODIFIED Requirements

### Requirement: Configured embed iframes use the live origin
When MorphNotes embeds Event Logs, Content Maker, or Project, those iframe `src` values MUST use the configured origins. Event Logs MUST keep the `/events-info` path on its origin. Data Access MUST NOT be embedded. A production build whose origin for one of the three modules is blank MUST NOT mount that iframe and MUST NOT substitute a loopback host. Local development with the variables unset MUST keep the current localhost defaults for those three modules.

#### Scenario: HTTPS origins are the iframe sources
- **WHEN** Event Logs, Content Maker, and Project origins are non-loopback `https` URLs
- **THEN** the Event Logs iframe `src` is that origin plus `/events-info`
- **AND** the Content Maker and Project iframe `src` values are those origins
- **AND** none of those `src` values use `localhost` or `127.0.0.1`
- **AND** Data Access has no iframe

#### Scenario: Blank production origin does not iframe the shell
- **WHEN** the production MorphNotes modal has a blank origin for a module
- **THEN** that module has no iframe `src`
- **AND** MorphNotes does not request `/events-info` on its own origin

### Requirement: A Morph session is handed to each embed as a bearer
An operator who is signed in on Morph and then opens Event Logs, Content Maker, or Project from the MorphNotes modal MUST be signed in on that embed without a second signup. The modal MUST pass the Morph bearer on the iframe `src` as `userspanel_token`. The Morph session cookie MUST stay a host-only cookie named `userspanel_session_token` with `SameSite=Lax` and MUST NOT set a `Domain` attribute.

#### Scenario: Signed-in shell passes the bearer
- **WHEN** Morph holds a bearer and an embed origin is configured
- **THEN** that embed iframe `src` includes `userspanel_token`
- **AND** the Morph session cookie is host-only with `SameSite=Lax` and no `Domain`

#### Scenario: Parent-domain cookies are not the production strategy
- **WHEN** an operator reads how a session crosses MorphNotes and an embed on `*.onrender.com`
- **THEN** the documented strategy is the bearer query handoff
- **AND** the notes say a cookie `Domain` of `onrender.com` is not set because that site is a public suffix

### Requirement: An unauthenticated visitor gets a path back to Morph
Event Logs, Content Maker, and Project MUST NOT mount their main UI when they have no Morph bearer. A production embed login link MUST NOT use `localhost` or `127.0.0.1`. There is no MorphUtils sign-in page.

#### Scenario: Production embed login is not localhost
- **WHEN** an embed builds its Morph sign-in link for a production page and no Morph origin is configured
- **THEN** that link does not use `localhost` or `127.0.0.1`

### Requirement: Docs list embed env and cookie failure modes
`deploy/README.md` MUST list the Event Logs, Content Maker, and Project origin variables used by the MorphNotes modal, and MUST NOT list Data Access as a product embed. It MUST name the cookie `userspanel_session_token`, `SameSite=Lax`, and that no `Domain` is set. It MUST state these failure modes: `*.onrender.com` hosts are separate sites so a parent cookie is not sent; `SameSite=Lax` cookies are not sent on a cross-site iframe load; the bearer in the first request can appear in that request's logs before the page strips it; signing out on one host does not clear the cookie on the others; opening an embed origin directly does not see the Morph host cookie. Example `onrender.com` hosts MUST stay labeled as examples and MUST NOT be required defaults in application source.

#### Scenario: A reader can configure origins and auth
- **WHEN** a reader follows the embed section in `deploy/README.md`
- **THEN** they see the Event Logs, Content Maker, and Project origin variables, the cookie name, `SameSite=Lax`, and the public-suffix failure mode
- **AND** Data Access is not listed as a product embed
