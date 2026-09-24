## Purpose

Keep the Morph AI transcript and composer readable and tappable on a phone, including while the soft keyboard is open.

## ADDED Requirements

### Requirement: Transcript prose fits the phone column
At about 390px CSS width, ordinary message prose in a multi-message Morph AI transcript MUST wrap inside the chat column. The transcript MUST NOT force the document or the chat pane to scroll horizontally for that prose.

#### Scenario: Wrapped prose
- **WHEN** an operator views several chat messages of normal prose at about 390px width
- **THEN** each message wraps within the chat column
- **AND** the document scroll width is not greater than the layout viewport

### Requirement: Composer stays visible with the soft keyboard
While the soft keyboard is open on a phone, the Morph AI composer input MUST remain in the visible viewport, and the send control MUST NOT stay covered for the whole time the keyboard is open. The composer MUST also clear the left and right safe areas.

#### Scenario: Keyboard open
- **WHEN** the operator focuses the composer and the soft keyboard opens
- **THEN** the text input remains visible
- **AND** the send control is not left under the keyboard

#### Scenario: Side safe area
- **WHEN** the viewport reports non-zero left or right safe-area insets
- **THEN** the composer padding is at least those insets

### Requirement: Composer actions meet the tap size
On a phone, the send control and the attach and skills entries in the composer, when those entries are present, MUST be at least 44 by 44 CSS pixels, or have padding that makes the hit area that size. Cancel and clear, when shown in the composer, MUST meet the same size.

#### Scenario: Send and skills
- **WHEN** the composer is shown at about 390px width
- **THEN** send is at least 44 by 44 CSS pixels
- **AND** the skills entry is at least 44 by 44 CSS pixels

#### Scenario: Attach when file upload is enabled
- **WHEN** the composer shows an attach control
- **THEN** that control is at least 44 by 44 CSS pixels

### Requirement: The chat pane scrolls vertically only
A long Morph AI transcript MUST scroll inside the chat pane on the block axis. That pane MUST NOT scroll on the inline axis, and MUST NOT push the page shell sideways.

#### Scenario: Long transcript
- **WHEN** the operator scrolls a transcript longer than the phone viewport
- **THEN** the chat pane moves on the block axis
- **AND** the page shell does not gain a horizontal scrollbar

### Requirement: Empty and error states fit the phone
The empty chat welcome and a send or network error MUST be readable at about 390px and MUST NOT overflow the chat column.

#### Scenario: Empty chat
- **WHEN** the transcript has no messages at about 390px width
- **THEN** the welcome content fits the chat column

#### Scenario: Send or network error
- **WHEN** Morph AI shows an error message at about 390px width
- **THEN** the error copy wraps inside the chat column

### Requirement: Visual blocks stay in the column
Code, Mermaid, and pixel-art blocks in a message MUST either scale to the chat column width or keep an enlarge path that does not widen the page. A wide block MAY scroll inside its own box.

#### Scenario: Diagram on a phone
- **WHEN** a message contains a Mermaid diagram at about 390px width
- **THEN** the diagram scales within the column or can be enlarged without widening the document

#### Scenario: Code block
- **WHEN** a message contains a long code line at about 390px width
- **THEN** the code scrolls inside its block
- **AND** the chat pane and the document do not scroll sideways
