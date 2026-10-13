# note-ai-grill Specification

## Purpose

Turns Notes & TODOs AI assist into a short interview. The operator answers a few on-topic questions, reviews the draft, and only then inserts it. The same page is the MCP grill app.

## Requirements

### Requirement: AI assist opens a grill before writing the body
On a note or TODO, AI assist MUST open a modal conversation when the title is non-empty. The first step MUST be a question about that title. Existing body text MUST be context for the grill and MUST NOT be replaced until the operator confirms a draft. AI assist MUST NOT start when the title is empty.

#### Scenario: New note with a title
- **WHEN** the operator opens a new note, sets the title, and clicks AI assist
- **THEN** a modal conversation opens
- **AND** the first question is about that title
- **AND** the note body stays empty until the operator confirms a draft

#### Scenario: Edit keeps the current body until confirm
- **WHEN** the operator clicks AI assist on a note or TODO that already has body text
- **THEN** the modal opens with that text as context
- **AND** the editor body is unchanged until the operator confirms

#### Scenario: Empty title
- **WHEN** the title is empty
- **THEN** AI assist does not open the grill and does not call the model

### Requirement: Question count follows the topic and stops at four
The grill MUST plan fewer questions for a short single-intent title and more for a longer or multi-part title. It MUST NOT plan more than four questions. It MUST ask one question at a time. Each planned question MUST be about the title. When the model returns no on-topic question, the grill MUST still ask a question that names the title.

#### Scenario: Short title
- **WHEN** the title is a short single intent such as "buy milk"
- **THEN** the grill plans one question

#### Scenario: Multi-part title
- **WHEN** the title is a longer or multi-part topic such as "make this morphAI a platform for agents, actually agent notebook"
- **THEN** the grill plans more than one question
- **AND** it plans at most four

#### Scenario: A long body does not add questions
- **WHEN** the title is a short single intent and the existing body is long
- **THEN** the grill still plans one question

#### Scenario: Model returns a chatty plan
- **WHEN** the model returns a greeting and no question about the title
- **THEN** the question shown to the operator names the title
- **AND** the greeting is not shown as the question

### Requirement: The operator can skip or finish early
The operator MUST be able to skip the current question or stop the interview and write the draft immediately. A skip MUST NOT be stored as a fact. An answer that asks to stop, including "enough" or "write it now", MUST finish the interview instead of being stored as a fact. The grill MUST call the model at most twice: once to plan questions and once to write the draft.

#### Scenario: Skip
- **WHEN** the operator skips the current question
- **THEN** the grill moves on without using that skip as content
- **AND** it does not call the model for the skip itself

#### Scenario: Finish early
- **WHEN** the operator chooses to write now, or answers with a stop phrase such as "enough, write it now"
- **THEN** the grill stops asking
- **AND** it shows a draft built from the title, the context, and the answers given so far

#### Scenario: A longer answer is not a stop phrase
- **WHEN** the operator answers with a sentence that mentions writing but is not only a stop phrase
- **THEN** the grill stores that sentence as the answer
- **AND** it continues the interview

#### Scenario: Empty answer
- **WHEN** the operator submits an empty answer
- **THEN** the grill stays on the same question

### Requirement: Confirmed draft is body text
After the interview, the grill MUST show a draft and MUST NOT insert it until the operator confirms. The draft MUST be the note or TODO body. It MUST NOT be a greeting, a self-introduction, or an offer of further help. A TODO draft MUST be actionable subtasks. Confirming in the Morph SPA MUST replace the editor body and MUST NOT save the item by itself. Cancelling MUST leave the editor body unchanged. In the MCP app, confirming MUST store a note for the verified session through the existing note create tool, and only after the draft is shown.

#### Scenario: Chatty model draft
- **WHEN** the model answers the title as an assistant ("Great! I'm Morph AI...")
- **THEN** the draft shown for review is note or TODO content about the title and the answers
- **AND** that assistant reply is not the inserted text

#### Scenario: Confirm in the SPA
- **WHEN** the operator confirms the draft in the Notes & TODOs modal
- **THEN** the editor body becomes that draft
- **AND** the item is not saved until the existing create or save control

#### Scenario: Cancel
- **WHEN** the operator closes the modal without confirming
- **THEN** the editor body is unchanged

#### Scenario: Model unavailable
- **WHEN** the model call fails
- **THEN** the editor body is unchanged
- **AND** answers already given are not discarded

#### Scenario: TODO draft
- **WHEN** the item is a TODO and the operator confirms
- **THEN** the editor body is a short list of subtasks about the title

### Requirement: Grill sessions belong to the signed-in user
The HTTP grill MUST require the same Morph session as other private `/api/tran/` routes. A missing session MUST be rejected and MUST NOT call the model. A grill id MUST be readable only by the session that started it. Another session MUST get the same not-found result it would for a missing id, and MUST NOT receive the draft.

#### Scenario: No session
- **WHEN** a client calls the grill HTTP route without a Morph session
- **THEN** the response is unauthorized
- **AND** no model call is made

#### Scenario: Other session
- **WHEN** a different signed-in user requests a grill id they did not start
- **THEN** the result is not found
- **AND** the draft is not included

### Requirement: One grill page for the SPA and MCP hosts
The conversation UI MUST be one HTML document served as an MCP app resource and rendered inside the Morph SPA AI assist modal. The page MUST use a dark background, light text, and an Arial or equivalent sans-serif stack. It MUST NOT contain an external URL, an HTML-sink assignment, or a session token. It MUST talk to the host only through the MCP app bridge.

#### Scenario: SPA modal renders the grill page
- **WHEN** the operator opens AI assist in Notes & TODOs
- **THEN** the modal renders that grill HTML document
- **AND** the page is dark and uses Arial or an equivalent sans-serif font

#### Scenario: Closed document
- **WHEN** the grill HTML is read
- **THEN** it contains no `http:` or `https:` URL and no HTML-sink assignment
- **AND** it contains no session token

### Requirement: Grill start accepts every note kind
`note_grill` start MUST accept `kind` values `note`, `todo`, `big_note`, `timeline`, `research`, `stick`, and `task`. An unknown kind MUST be `invalid kind`. The interview MUST stay in memory and MUST NOT insert a row. The draft MUST be body text for that kind (a TODO stays bullets; a timeline draft lists dated entries; research names findings). Another user's grill id MUST still be not found.

#### Scenario: Start a timeline grill
- **WHEN** a client starts `note_grill` with kind `timeline` and a title
- **THEN** the result is grill state for that session user
- **AND** the state kind is `timeline`
- **AND** no timeline row is written

#### Scenario: Unknown kind
- **WHEN** a client starts `note_grill` with kind `banana`
- **THEN** the result is an invalid-kind error
- **AND** no session is stored
