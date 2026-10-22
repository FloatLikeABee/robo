# Spec Delta

## ADDED Requirements

### Requirement: Default dorph_guide names html-anything
When `dorph_guide` is called with no topic, the returned markdown MUST name html-anything. The full playbook at `dorph://guide` MUST still stay identical to `docs/agents/16-dorph-for-agents.md`.

#### Scenario: Empty topic mentions the skill
- **WHEN** an agent calls `dorph_guide` with no topic
- **THEN** the excerpt contains `html-anything`
