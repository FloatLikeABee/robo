## Purpose

Let a signed-in Morph human read the notes their agents posted under that same account, without seeing another user's notes or any notes while logged out.

## ADDED Requirements

### Requirement: Signed-in human sees only their agent notes
A signed-in human who opens the agent-notes view MUST see notes stored for the one active Tran user linked to that session, and only those notes that carry the agent marker in the title or the body. The list MUST show title, time, and status. Opening a row MUST show that note's body. Notes that lack both markers, todos, and rows owned by anyone else MUST NOT appear.

#### Scenario: Own agent note is listed
- **WHEN** the signed-in user has an agent-authored note and opens the agent-notes view
- **THEN** the list shows that note's title, time, and status
- **AND** the detail view shows the body

#### Scenario: A marker that survives an edit still counts
- **WHEN** an agent note has had its title changed but the body still starts with the agent source line, or the title still starts with the agent title mark and the body line was removed
- **THEN** the note still appears for that user

#### Scenario: Human notes and todos stay out
- **WHEN** the same user also has a note or todo that does not carry an agent marker
- **THEN** that row is absent from the agent-notes list and detail

### Requirement: Another user cannot read these notes
The agent-notes list and detail MUST scope to the verified session subject. A client-supplied user id, including `user_id` and `X-User-ID`, MUST NOT select the owner. When the session maps to no active Tran user, or to more than one, the response MUST NOT return any note and MUST NOT fall back to user id 1. A request for another user's note MUST be the same not-found result as a missing id and MUST NOT include that note's title or body.

#### Scenario: User B requests user A's note
- **WHEN** user B is signed in and requests the list or user A's agent-note id
- **THEN** A's title and body are absent
- **AND** the detail result is not found

#### Scenario: Query user id is ignored
- **WHEN** a signed-in user passes another user's id as `user_id`
- **THEN** the response still contains only the caller's agent notes

#### Scenario: No single notes user
- **WHEN** the session has no active Tran user, or two active Tran users share the session email
- **THEN** no note text is returned
- **AND** user id 1's notes are not returned

### Requirement: Logged-out access is refused
Opening the agent-notes view without a session MUST send the human to login. The list and detail APIs MUST reject a request that has no valid session, including a request that only sends `X-User-ID`, and MUST NOT include note text.

#### Scenario: Logged-out page
- **WHEN** a human opens the agent-notes view with no session
- **THEN** they are sent to login
- **AND** the view does not request note data

#### Scenario: Logged-out API
- **WHEN** a client calls the agent-notes list or detail without a valid session
- **THEN** the response is unauthorized
- **AND** the body does not include a stored note title or body

### Requirement: List and detail stay readable on a phone and on a desktop
At about 390px width and at a desktop width, the list and the detail MUST fit the layout width, wrap long text, and leave a way back from detail to the list on a phone. An empty list MUST explain that agents post notes for the signed-in user. The page MUST stay dark and clear the safe areas.

#### Scenario: Phone width
- **WHEN** the human browses the list and opens a note at about 390px width
- **THEN** the title, time, status, and body are readable without a horizontal page scroll
- **AND** a control returns to the list

#### Scenario: Desktop width
- **WHEN** the human browses the list and a note at a desktop width
- **THEN** the list and the detail are both readable

#### Scenario: Empty list
- **WHEN** the signed-in user has no agent notes
- **THEN** the view explains that agents post notes under this account
