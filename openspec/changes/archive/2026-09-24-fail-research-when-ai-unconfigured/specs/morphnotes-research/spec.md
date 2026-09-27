## Purpose

Make a MorphNotes Research job’s terminal status honest when AI is missing or rounds fail, so operators do not treat a failed run as a finished thesis.

## ADDED Requirements

### Requirement: Unconfigured AI does not start a Research job
When AI is not configured, a request to create a Research job with a valid prompt MUST be rejected with HTTP 503. The response MUST include a reason that tells the operator to set `MORPH_AI_API_KEY`. The system MUST NOT store a research job for that request. An empty prompt or an unsupported upload MUST still be rejected as a bad request even when AI is not configured.

#### Scenario: Create with a prompt and no usable AI
- **WHEN** the operator submits a non-empty research prompt and AI is not configured
- **THEN** the response status is 503
- **AND** the response tells the operator to set `MORPH_AI_API_KEY`
- **AND** no research job is stored

#### Scenario: Invalid create stays a bad request
- **WHEN** the operator submits an empty prompt, or a prompt with an unsupported file, and AI is not configured
- **THEN** the response status is 400
- **AND** no research job is stored

### Requirement: A Research job that cannot use AI ends failed
A Research job that runs while AI is not configured, or whose round writes all fail, MUST end with status `failed` and MUST NOT end with status `complete`. The job payload MUST include `error_text` that states the cause in plain language. The stored conclusion (`markdown_content`) MUST be empty. The system MUST NOT present concatenated “Round N failed” text as the conclusion. If AI is discovered to be not configured after an earlier round succeeded, the job MUST still end `failed` with that not-configured reason and an empty conclusion, and pieces already stored MUST remain.

#### Scenario: Run with AI not configured
- **WHEN** a Research job runs and AI is not configured
- **THEN** the job status is `failed`
- **AND** `error_text` tells the operator that AI is not configured
- **AND** the stored conclusion is empty

#### Scenario: Every round write fails
- **WHEN** a Research job runs and every round’s writing call fails
- **THEN** the job status is `failed`
- **AND** `error_text` states that the research rounds failed
- **AND** the stored conclusion is empty

### Requirement: Partial round success is distinct from a failed job
A round write that succeeds is a successful round, including when verification is unavailable. When at least one round succeeds, at least one round fails, and thesis synthesis produces a conclusion, the job status MUST be `complete`, the stored conclusion MUST be that synthesized document, and `error_text` MUST state how many rounds failed and which rounds they were. When at least one round succeeds and synthesis produces no conclusion, the job status MUST be `failed`, `error_text` MUST state that the thesis could not be synthesized, the stored conclusion MUST be empty, and the successful round pieces MUST remain stored. A run where every round write succeeds and synthesis produces a conclusion MUST end `complete` with empty `error_text`.

#### Scenario: Some rounds fail and synthesis succeeds
- **WHEN** some round writes succeed, some round writes fail, and synthesis returns a thesis
- **THEN** the job status is `complete`
- **AND** `error_text` names the failed rounds and how many failed
- **AND** the stored conclusion is that thesis

#### Scenario: Synthesis produces nothing after a successful round
- **WHEN** at least one round write succeeds and synthesis produces no conclusion
- **THEN** the job status is `failed`
- **AND** `error_text` states that the thesis could not be synthesized
- **AND** the stored conclusion is empty
- **AND** the successful round pieces are still stored

#### Scenario: Every round and synthesis succeed
- **WHEN** every round write succeeds and synthesis returns a thesis
- **THEN** the job status is `complete`
- **AND** `error_text` is empty
- **AND** the stored conclusion is that thesis

### Requirement: The Research UI shows the failed state and the reason
The Research list and the open job MUST show status `failed` and the `error_text` cause. A failed job with an empty conclusion MUST NOT show a thesis document in the editor. A `complete` job whose `error_text` names failed rounds MUST show the thesis and a warning that names those rounds. Publishing a job whose status is `failed` MUST be rejected.

#### Scenario: Operator opens a failed job
- **WHEN** the operator opens a Research job whose status is `failed` and whose `error_text` explains the cause
- **THEN** the list row shows `failed` and the cause
- **AND** the detail view shows the cause
- **AND** the editor does not contain a thesis

#### Scenario: Operator opens a complete job with failed rounds
- **WHEN** the operator opens a Research job whose status is `complete` and whose `error_text` names failed rounds
- **THEN** the thesis is shown
- **AND** a warning shows that rounds failed

#### Scenario: Publish of a failed job is rejected
- **WHEN** a client publishes a Research job whose status is `failed`
- **THEN** the response status is 409
- **AND** the job stays unpublished
