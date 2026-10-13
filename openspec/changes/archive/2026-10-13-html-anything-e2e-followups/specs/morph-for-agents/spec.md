# Spec Delta

## ADDED Requirements

### Requirement: Default morph_guide names html-anything
When `morph_guide` is called with no topic, the returned markdown MUST name html-anything. The full playbook at `morph://guide` MUST still stay identical to `docs/agents/16-morph-for-agents.md`.

#### Scenario: Empty topic mentions the skill
- **WHEN** an agent calls `morph_guide` with no topic
- **THEN** the excerpt contains `html-anything`
