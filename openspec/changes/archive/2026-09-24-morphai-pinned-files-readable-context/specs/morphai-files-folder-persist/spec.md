## ADDED Requirements

### Requirement: Session snapshot may store pinned file text

For pinned workspace files that are readable and within the existing per-file size cap, Morph AI MUST persist their text in the per-session IndexedDB binding alongside path metadata so chat context survives refresh without a live directory handle.

#### Scenario: Refresh retains pinned text for chat

- **WHEN** the operator pinned readable files, sent chat successfully, then reloaded Morph AI
- **AND** Files shows the folder listing from snapshot
- **THEN** pinned file text remains available from the session binding for chat context
- **AND** the operator is not required to re-pin those files
