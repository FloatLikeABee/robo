## REMOVED Requirements

### Requirement: Twenty online rounds with verification
**Reason:** Twenty rounds is too slow and produces more fragments than a useful conclusion needs.
**Migration:** New jobs run five verified rounds. Jobs already in progress keep the round target stored on that job. Completed twenty-round jobs keep their stored pieces; operators still open them in Research.

### Requirement: Combined conclusion then final refine
**Reason:** Concatenating pieces in round order (even after a light refine) is not a finished document.
**Migration:** After the last round, the system writes a thesis-style synthesis from all pieces and verifications, then a critic polish pass. Per-round pieces remain on the Rounds tab.

## ADDED Requirements

### Requirement: Five online rounds with verification
Each new research job MUST run exactly five online research rounds unless the operator cancels or the job fails. Each round MUST gather online sources, produce a conclusion piece for that step, and run a verification pass that checks claims against that round’s sources. Pieces MUST be stored in round order and remain visible if the operator refreshes during the run. Jobs created before this change MUST keep the round count they started with.

#### Scenario: Five pieces after a complete run
- **WHEN** a new job finishes all research rounds successfully
- **THEN** it has five stored pieces in order 1 through 5
- **AND** each piece includes the round’s research writing plus a verification result

#### Scenario: Progress is visible during the run
- **WHEN** a new job is on round 3 of 5
- **THEN** the operator can see that round 3 is in progress
- **AND** pieces from rounds 1–2 are already readable

#### Scenario: In-flight jobs keep their original round target
- **WHEN** a job was created to run twenty rounds and is still running after this change
- **THEN** that job continues until its original round target is done
- **AND** new jobs started after the change run five rounds

### Requirement: Thesis-quality synthesized conclusion
After the last research round, the system MUST produce one stored Markdown conclusion (and HTML from it) that reads as a single professional document: argument and structure chosen for the subject, absorbing the best-supported material from every round. The conclusion MUST NOT be a concatenation of round pieces, MUST NOT use research-round order as its outline, and MUST NOT be organized as Round 1…Round N headings. Uncertain or unverified claims MUST remain qualified. Per-round pieces MUST stay available separately from this conclusion. The job MUST NOT be marked complete until that synthesized conclusion is stored.

#### Scenario: Conclusion is a unified document
- **WHEN** synthesis finishes
- **THEN** the stored Markdown is one coherent document answering the original prompt
- **AND** it is not a sequence of Round 1, Round 2, … section headers for each piece
- **AND** HTML for that document is also stored

#### Scenario: Essence over dump
- **WHEN** two rounds overlap or one round’s claims failed verification
- **THEN** the conclusion keeps the well-supported substance once, in a natural structure
- **AND** it does not paste both round write-ups in full in round order

#### Scenario: Synthesis is last
- **WHEN** the last round has finished but synthesis has not
- **THEN** the job is not marked complete
- **AND** the operator can still read the per-round pieces
