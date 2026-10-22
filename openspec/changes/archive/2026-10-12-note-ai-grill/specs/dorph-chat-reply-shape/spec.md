## MODIFIED Requirements

### Requirement: Plain questions get plain answers
Dorph AI chat MUST answer the operator's question in concise plain language. A diagram, chart, table, or code block MUST be included only when the question is about structure, quantities, or code and that form makes the answer clearer. An ordinary question MUST NOT be answered as a caption plus a diagram, a pixel grid, or a code or architecture write-up. Note and TODO AI assist MUST NOT treat the title as a chat question. That path follows the note grill and inserts body text only after the operator confirms.

#### Scenario: Identity question
- **WHEN** the operator asks which model is in use
- **THEN** the reply is short prose
- **AND** it does not include a diagram, a caption label, or a JSON metadata block

#### Scenario: Structure question
- **WHEN** the operator asks how a process or architecture is structured
- **THEN** the reply MAY include one diagram or code block that explains that structure

#### Scenario: Note title is not a chat question
- **WHEN** the operator runs AI assist on a note or TODO title
- **THEN** the assistant does not insert a direct answer to that title
- **AND** the body changes only after the note grill confirms a draft

### Requirement: Model identity is not invented
When the operator asks Dorph AI chat which model or version is in use, the reply MUST name the configured provider and model. If the provider or model is not configured, the reply MUST say that part is unknown. The reply MUST NOT invent a product version, model name, or update date. Note and TODO AI assist does not answer that question. It follows the note grill, and the note body MUST NOT introduce the assistant or invent a model identity.

#### Scenario: Configured provider and model
- **WHEN** Dorph is configured with a provider and a model and the operator asks which model is in use
- **THEN** the reply names that provider and that model
- **AND** it does not add a version or date that was not configured

#### Scenario: Unknown configuration
- **WHEN** no provider and no model can be resolved and the operator asks which model is in use
- **THEN** the reply says the model is unknown

#### Scenario: Note body does not introduce the assistant
- **WHEN** AI assist finishes a note or TODO draft
- **THEN** the draft does not introduce the assistant or name a model the operator did not supply
