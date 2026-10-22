## MODIFIED Requirements

### Requirement: Logged-out access is refused
Opening the agent-notes path without a session MUST send the human to login. The list and detail APIs MUST reject a request that has no valid session, including a request that only sends `X-User-ID`, and MUST NOT include note text. The chat menu that opens the view is available only inside a signed-in chat.

#### Scenario: Logged-out page
- **WHEN** a human opens the agent-notes path with no session
- **THEN** they are sent to login
- **AND** the view does not request note data

#### Scenario: Logged-out API
- **WHEN** a client calls the agent-notes list or detail without a valid session
- **THEN** the response is unauthorized
- **AND** the body does not include a stored note title or body

### Requirement: List and detail stay readable on a phone and on a desktop
At about 390px width and at a desktop width, the list and the detail MUST fit the layout width, wrap long text, and leave a way back from detail to the list on a phone. An empty list MUST explain that agents post notes for the signed-in user. A failed load MUST show the error and an empty list. The view is a large modal over the current chat. It MUST stay dark, clear the safe areas, and MUST NOT replace the chat with a separate page.

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

#### Scenario: Failed load
- **WHEN** the list request fails
- **THEN** the view shows the error
- **AND** it does not show note rows

## ADDED Requirements

### Requirement: Agent notes opens over the chat
The header and the more menu MUST open the agent-notes view as a large modal over the current chat. That action MUST NOT navigate away from the chat. Escape and a close control MUST dismiss the modal. Focus MUST move into the modal when it opens and return to the control that opened it when it closes. A signed-in visit to the agent-notes path MUST land in the chat with the same modal open.

#### Scenario: Menu opens the modal
- **WHEN** a signed-in human chooses Agent notes from the header or the more menu
- **THEN** the list opens over the current chat
- **AND** the chat address does not change to a separate agent-notes page

#### Scenario: Dismiss
- **WHEN** the modal is open and the human presses Escape or the close control
- **THEN** the modal closes
- **AND** focus returns to the control that opened it

#### Scenario: Signed-in deep link
- **WHEN** a signed-in human opens the agent-notes path
- **THEN** the chat is showing
- **AND** the same agent-notes modal is open
