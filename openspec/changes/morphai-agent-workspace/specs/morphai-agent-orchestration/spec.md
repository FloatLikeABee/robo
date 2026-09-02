## Purpose

Lets Morph AI run as an agent: the user controls what context is included, unchanged context is cached, and tasks can spawn sub-agents whose results merge into one reply.

## ADDED Requirements

### Requirement: Context controls
The agent shell MUST show which context sources are included on the next send: pinned workspace files, Notes & TODOs, and Knowledge / HybridContext. The user MUST be able to include or exclude each source without deleting the underlying data. Excluded sources MUST NOT be sent as prompt context on the next turn.

#### Scenario: Exclude knowledge
- **WHEN** the user turns off Knowledge / HybridContext for the session
- **THEN** the next chat send does not include that knowledge in the model prompt
- **AND** knowledge files remain in the library

#### Scenario: Include pinned files
- **WHEN** the user has pinned workspace files and those pins are on
- **THEN** the next chat send includes those files (or their cached digest) as context

### Requirement: Context cache
The system MUST cache unchanged included context for the session so the same folder files and knowledge chunks are not fully re-uploaded on every turn. Changing a pin set, notes snapshot, or knowledge source MUST invalidate the affected cache entries. Cache MUST NOT leak another user’s session context.

#### Scenario: Repeat turn with same pins
- **WHEN** the user sends a second message with the same included files and knowledge as the previous turn
- **THEN** the system reuses the session context cache instead of re-reading every file from scratch
- **AND** the reply still reflects that cached context

#### Scenario: Pin change invalidates
- **WHEN** the user unpins a file or adds a new pin
- **THEN** the next send MUST NOT reuse a cache entry that still assumed the old pin set

### Requirement: Multi sub-agent by task
When a user message needs more than one kind of work (for example MorphData lookup plus file reading, or research plus notes), Morph AI MUST be able to run multiple sub-agents for those tasks. The UI MUST show that sub-agents are running. Results MUST merge into a single assistant reply in the parent session. A simple question MUST still complete with one agent and no required sub-agent fan-out.

#### Scenario: Task fans out
- **WHEN** the user asks something that needs both workspace files and MorphData
- **THEN** the system starts more than one sub-agent
- **AND** the user sees sub-agent progress
- **AND** one combined assistant message appears in the chat

#### Scenario: Simple chat stays single
- **WHEN** the user asks a short question that does not need extra tools or files
- **THEN** the system MAY skip sub-agents
- **AND** a single assistant reply still appears

### Requirement: Sub-agents cannot write the user’s folder
Sub-agents MUST NOT create, modify, or delete files in the user-opened workspace folder. They MAY read pinned files and MAY write only to Morph stores (notes, HybridContext, knowledge, chat messages) already allowed today.

#### Scenario: No disk writes
- **WHEN** a sub-agent finishes a file-related task
- **THEN** the local workspace folder contents on disk are unchanged
- **AND** any new artifact is stored in Morph (chat, notes, or knowledge), not as a silent overwrite of the user’s files
