## Purpose

Ensures pinned Morph AI workspace files actually supply text to the chat agent when Files include is on, with clear feedback when content cannot be read.

## ADDED Requirements

### Requirement: Pinned files contribute text on send

When Files include is on and one or more workspace files are pinned, Morph AI MUST send non-empty `content` for each readable pinned file in `pinned_files` on `POST /api/chat`. The backend MUST receive enough text for `AssemblePinnedBlob` to build a non-empty pinned-files context block.

#### Scenario: Pin then ask about file content

- **WHEN** the operator pins a readable workspace file
- **AND** Files include is on
- **AND** they send a message asking about that file
- **THEN** the request includes that file's text in `pinned_files[].content`
- **AND** the assistant reply is grounded in that text (not “I cannot see the file”)

#### Scenario: Snapshot workspace after refresh

- **WHEN** the operator refreshed Morph AI and Files shows a folder listing from session snapshot (no live handle yet)
- **AND** they had previously pinned files whose text was cached at pin time
- **AND** Files include is on
- **THEN** send still includes cached pinned text without requiring a live directory handle

### Requirement: Unreadable pins fail loudly

Morph AI MUST NOT silently send an empty pinned-files payload when pins are active and Files include is on. If no pinned file body can be read, send MUST be blocked with an operator-visible error explaining that folder access is required (e.g. Reconnect).

#### Scenario: Pin without readable body

- **WHEN** pinned paths exist but none can be read (no handle permission, no cached text)
- **AND** the operator tries to send with Files include on
- **THEN** Morph AI shows an inline error
- **AND** the chat request is not sent until the issue is resolved or pins are cleared / Files include is turned off

#### Scenario: Files chip does not imply success alone

- **WHEN** Files (N) is on but pinned content is unreadable
- **THEN** the UI indicates the problem (warning on chip and/or pinned row)
- **AND** toggling Files off allows send without pinned context

### Requirement: Pin action captures content when possible

When the operator pins a non-skipped file, Morph AI MUST attempt to read file text immediately and cache it on the session binding for later sends and refresh.

#### Scenario: Pin caches text

- **WHEN** the operator pins a readable file under 512 KiB
- **THEN** its text is stored in the session workspace binding keyed by path
- **AND** subsequent sends use that cache even if the in-memory `File` handle is gone

#### Scenario: Unpin clears cache

- **WHEN** the operator unpins a file
- **THEN** its cached text is removed from the session binding
