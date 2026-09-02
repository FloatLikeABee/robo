## Purpose

Gives Morph AI a collapsible session list, a split chat-and-workspace shell, and a taller composer so the conversation has room next to agent work.

## ADDED Requirements

### Requirement: Session list collapses in place
On the main Morph AI chat (not embedded Morph Data), the left session menu MUST be collapsible in the layout so chat and workspace gain width. When expanded, the user MUST still create, switch, rename, and delete sessions. Collapse state MUST persist for the browser session.

#### Scenario: Collapse to free width
- **WHEN** the user collapses the session menu
- **THEN** the session titles no longer occupy the previous sidebar width
- **AND** the chat and workspace panes expand into that space
- **AND** a control remains to expand the list again

#### Scenario: Expanded list still works
- **WHEN** the session menu is expanded
- **THEN** the user can start a new chat, switch sessions, rename, and delete (except the default session)

#### Scenario: Embedded chat unchanged
- **WHEN** Morph AI is shown as embedded Morph Data chat (`singleSession`)
- **THEN** the session rail and collapse control MUST NOT appear

### Requirement: Chat and workspace split
The main Morph AI panel MUST split into two columns of approximately equal width: conversation on the left (messages plus composer) and workspace on the right. The composer MUST sit in the left column only, to the left of the workspace.

#### Scenario: Two columns
- **WHEN** the user is on main Morph AI
- **THEN** messages and the input bar occupy the left column
- **AND** the workspace occupies the right column
- **AND** the input bar does not stretch under the workspace

#### Scenario: Narrow viewport
- **WHEN** the viewport is too narrow for two readable columns
- **THEN** the workspace MUST stack below the chat or be reachable via a tab/toggle
- **AND** the composer MUST remain with the chat column

### Requirement: Taller composer
The Morph AI composer on the main agent shell MUST be about three times the current default height so the user can paste longer context without the field staying a single line.

#### Scenario: Default height
- **WHEN** the user focuses the chat input on main Morph AI
- **THEN** the visible input area is approximately three times the previous default height (about three lines of text at rest)
- **AND** the field still grows with content up to a larger max height
