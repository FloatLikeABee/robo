## Purpose

Operators can run a MorphNotes Research job from a prompt and optional files, watch twenty verified online research rounds, and get a stepwise Markdown plus publishable HTML conclusion.

## ADDED Requirements

### Requirement: Research is a MorphNotes module
MorphNotes MUST expose a Research module in the same admin navigation as Timelines and Big notes. The module MUST use a list of research jobs plus a detail pane. Opening Research with a non-empty list MUST select the first job.

#### Scenario: Open Research from MorphNotes nav
- **WHEN** the operator opens MorphNotes and chooses Research
- **THEN** they see a Research list/detail workspace
- **AND** they do not have to leave MorphNotes or open MorphUtils

#### Scenario: First research job is selected
- **WHEN** the operator opens Research and at least one job exists
- **THEN** the first list item is selected and its detail is shown

### Requirement: Start from a prompt and optional files
The operator MUST be able to start a research job with a text prompt on any subject. They MAY attach files of types txt, pdf, csv, and json. A prompt with no files MUST still start. Files other than those types MUST be rejected without starting the job.

#### Scenario: Prompt only
- **WHEN** the operator submits a non-empty prompt with no files
- **THEN** a research job is created and research begins

#### Scenario: Prompt plus allowed files
- **WHEN** the operator submits a prompt with txt, pdf, csv, and/or json files
- **THEN** those files are accepted as references for the job

#### Scenario: Disallowed file type
- **WHEN** the operator attaches a file that is not txt, pdf, csv, or json
- **THEN** the job is not started
- **AND** the operator is told the file type is not allowed

### Requirement: Uploaded files become searchable reference before research rounds
If the job has files, the system MUST extract their text and index it as retrieval-augmented reference for that job before the first online research round. Later rounds MUST be able to retrieve from those files. Indexing MUST work on the supported local stack (SQLite); Neo4j MUST NOT be required.

#### Scenario: Files are indexed before round one
- **WHEN** the operator starts a job with at least one allowed file
- **THEN** the file text is indexed as reference for that job
- **AND** round one does not start until that index step has finished or failed visibly

#### Scenario: Retrieval uses the uploaded files
- **WHEN** a later round needs reference material from an uploaded pdf or csv
- **THEN** the round can include retrieved snippets from those files in its working context

#### Scenario: No files skips ingest
- **WHEN** the operator starts a prompt-only job
- **THEN** research rounds still run
- **AND** no file-index failure is shown

### Requirement: Twenty online rounds with verification
Each research job MUST run exactly twenty online research rounds unless the operator cancels or the job fails. Each round MUST gather online sources, produce a conclusion piece for that step, and run a verification subagent that checks claims against the round’s sources. Pieces MUST be stored in round order and remain visible if the operator refreshes during the run.

#### Scenario: Twenty pieces after a complete run
- **WHEN** a job finishes all research rounds successfully
- **THEN** it has twenty stored pieces in order 1 through 20
- **AND** each piece includes the round’s research writing plus a verification result

#### Scenario: Progress is visible during the run
- **WHEN** a job is on round 7 of 20
- **THEN** the operator can see that round 7 is in progress
- **AND** pieces from rounds 1–6 are already readable

#### Scenario: Refresh keeps pieces
- **WHEN** the operator reloads Research while a job is running
- **THEN** completed pieces are still listed
- **AND** the job continues or resumes from the next unfinished round

### Requirement: Combined conclusion then final refine
After round 20, the system MUST combine the pieces in research-step order into one conclusion document, then run a final refine pass. The job MUST store both Markdown and HTML of the refined conclusion.

#### Scenario: Conclusion follows research steps
- **WHEN** the refine pass completes
- **THEN** the conclusion Markdown contains the twenty pieces in round order (refined, not shuffled)
- **AND** HTML for that document is also stored

#### Scenario: Refine is last
- **WHEN** round 20 has finished but refine has not
- **THEN** the job is not marked complete
- **AND** the operator can still see the twenty pieces

### Requirement: Markdown display and HTML publish
The conclusion Markdown MUST display as rendered Markdown with a Raw edit mode (same behavior as other MorphNotes stored Markdown). HTML MUST be publishable on a Morph-hosted public URL in the same family as Big notes / Timelines, without requiring Content Maker.

#### Scenario: Conclusion Markdown is rendered
- **WHEN** the operator opens a completed research job
- **THEN** the Markdown conclusion is shown rendered, not as raw source only
- **AND** they can switch to Raw to edit the stored Markdown

#### Scenario: Publish HTML
- **WHEN** the operator publishes a completed research job that has HTML
- **THEN** they receive a public URL for that HTML
- **AND** opening the URL shows the research document without signing in

### Requirement: Failure and empty prompt
An empty prompt MUST NOT start a job. If online gather or the language model fails for a round, the job MUST record the failure on that round and MUST NOT silently skip to a fake complete document. Cancel MUST stop further rounds.

#### Scenario: Empty prompt rejected
- **WHEN** the operator submits Research with a blank prompt
- **THEN** no job is created

#### Scenario: Round failure is visible
- **WHEN** round 4 cannot gather sources or call the model
- **THEN** the job records a failed piece for round 4 that the operator can read
- **AND** later rounds still run unless the operator cancelled the job
