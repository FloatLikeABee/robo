# agent-lesson-control Specification

## Purpose

Operators can see, disable, and delete the session lessons that belong to them, and later chats follow only the enabled lessons of the signed-in user.

## Requirements

### Requirement: Lessons are enabled by default
The system MUST store an enabled flag on each agent lesson. Existing rows MUST remain enabled after migration. New lessons MUST be stored enabled.

#### Scenario: Existing row stays enabled
- **WHEN** the lesson table is migrated and a row was stored before the enabled flag existed
- **THEN** that row is enabled

#### Scenario: New lesson starts enabled
- **WHEN** a significant session distills a lesson for the signed-in user
- **THEN** the stored lesson is enabled

### Requirement: Legacy lessons are claimed once
Lessons stored before ownership MUST stay unowned until the system makes a one-shot decision. When platform accounts are missing or there are zero accounts, the system MUST leave those rows unowned and MUST try again on a later startup. The first time at least one account exists, if there is exactly one account the system MUST assign every still-unowned lesson to that account and MUST NOT assign again later. If that account already has a lesson for the same source session, the system MUST keep the account's lesson and MUST delete the conflicting unowned row. The claim MUST finish without failing startup. If more than one account exists at that decision, the system MUST leave unowned rows unowned forever. A database that has lessons but no platform-account table MUST still start, and those lessons MUST stay unowned.

#### Scenario: Single account keeps historical lessons
- **WHEN** migration runs and exactly one platform account exists
- **AND** unowned lessons exist
- **THEN** those lessons belong to that account
- **AND** a later startup does not assign any newer unowned lesson

#### Scenario: Several accounts do not receive historical lessons
- **WHEN** the first ownership decision runs and more than one platform account exists
- **THEN** unowned lessons stay unowned
- **AND** after accounts are later reduced to one, those lessons stay unowned

#### Scenario: Harvested lesson wins a session collision
- **WHEN** migration has already run while there are zero accounts
- **AND** the only account then stores a lesson for a source session that an unowned lesson also uses
- **AND** migration runs again
- **THEN** startup succeeds
- **AND** the account keeps its lesson
- **AND** the unowned row is gone

#### Scenario: Missing account table does not block startup
- **WHEN** the lesson table is migrated and the platform-account table is absent
- **THEN** migration succeeds
- **AND** existing lessons stay enabled and unowned

### Requirement: Each user has their own lesson for a session
The system MUST allow two users to each store one lesson for the same source session id, including the shared default session. The system MUST NOT store a second lesson for the same user and source session.

#### Scenario: Two users share the default session id
- **WHEN** two users each finish a significant chat in the session id "default"
- **THEN** each user has their own lesson for that session

### Requirement: Caller lists only their lessons
`GET /api/agent-lessons` MUST require a verified bearer token for an existing platform user. It MUST return that user's lessons, including disabled ones, and MUST NOT return another user's lessons. A request with no bearer token MUST be rejected with 401 even when the lesson store is unavailable. Through the auth middleware, a request the middleware does not accept MUST be 401 before the handler runs. A bearer token presented to the handler when the store cannot check it MUST be 503. A client `X-User-ID` header MUST NOT select the user. `auth_user_id` MUST NOT select the user, because the middleware still copies a client header into that value when no bearer was verified.

#### Scenario: Owner sees enabled and disabled lessons
- **WHEN** an authenticated user lists lessons
- **AND** they own one enabled lesson and one disabled lesson
- **THEN** the response includes both
- **AND** it does not include another user's lesson

#### Scenario: Middleware rejects a request the store cannot authenticate
- **WHEN** the lesson store is unavailable
- **AND** a request reaches the lesson route through the auth middleware with a bearer token and no `X-User-ID`
- **THEN** the response is 401

#### Scenario: Header is not identity
- **WHEN** a request sets `X-User-ID` to another user and does not send a bearer token
- **THEN** the response is 401
- **AND** that user's lessons are unchanged

### Requirement: Owner can disable or delete only their lesson
`PATCH /api/agent-lessons/:id` with `enabled` MUST change that flag only when the caller owns the lesson. The PATCH body MUST NOT contain any field other than `enabled`. `DELETE /api/agent-lessons/:id` MUST remove the lesson only when the caller owns it. A missing id and another user's id MUST both return 404 with the same error. The caller MUST be a verified bearer user. Disabling or deleting a lesson MUST change the next management-chat answer for the same prompt so a previously cached reply is not returned. The management-chat exact cache MUST be keyed by the verified bearer user and that user's enabled lessons. A client `X-User-ID` MUST NOT read or write that cache.

#### Scenario: Owner disables a lesson
- **WHEN** the owner patches the lesson with enabled false
- **THEN** the lesson is disabled
- **AND** it remains in the owner's list

#### Scenario: Unknown patch field is rejected
- **WHEN** the owner patches a lesson with `enabled` and another field
- **THEN** the response is 400
- **AND** the lesson is unchanged

#### Scenario: Disabled lesson is not served from cache
- **WHEN** a verified user has a cached management-chat reply for a prompt while a lesson is enabled
- **AND** they disable that lesson
- **AND** they send the same prompt
- **THEN** the reply is not the cached reply from before the disable

#### Scenario: Header does not read the exact cache
- **WHEN** a verified user's prompt is in the management-chat exact cache
- **AND** a later request presents that user's id only in `X-User-ID`
- **THEN** the cached reply is not returned

#### Scenario: Other user cannot change or delete
- **WHEN** a different authenticated user patches or deletes the lesson
- **THEN** the response is 404
- **AND** the lesson is unchanged

### Requirement: Prompts and the skills catalog use enabled lessons only
Chat prompt injection and the lessons embedded in `GET /api/skills` MUST include only enabled lessons owned by the verified bearer user. Disabled lessons and other users' lessons MUST be omitted. When no verified bearer user is present, prompt injection MUST include no lessons and the skills lessons array MUST be empty. Harvest MUST record the bearer user's id and MUST NOT record a lesson for a header-only caller.

#### Scenario: Disabled lesson is not injected
- **WHEN** the signed-in user has an enabled lesson and a disabled lesson
- **THEN** the chat prompt contains the enabled rule
- **AND** it does not contain the disabled rule

#### Scenario: Skills catalog hides disabled and foreign lessons
- **WHEN** the signed-in user requests the skills catalog
- **THEN** the embedded lessons include their enabled lesson
- **AND** they do not include their disabled lesson or another user's lesson
