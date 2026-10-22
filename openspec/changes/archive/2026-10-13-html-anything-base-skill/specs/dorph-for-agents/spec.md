# Spec Delta

## ADDED Requirements

### Requirement: Playbook names the html-anything base skill
`docs/agents/16-dorph-for-agents.md` MUST tell agents that Dorph's HTML base skill is the vendored html-anything skill, that Ge's Arial dark multi-accent overlay applies, and where the vendor files live. The embedded MCP guide MUST stay identical to that file.

#### Scenario: Agent reads writing or HTML guidance
- **WHEN** an agent reads `docs/agents/16-dorph-for-agents.md`
- **THEN** the file names html-anything
- **AND** it names `skills/vendor/html-anything/`
- **AND** it says HTML follows Arial, dark mode, and varied colours

### Requirement: Skills pack exposes html-anything
The dorph-mcp skills pack MUST expose html-anything as a skill agents can load. The Dorph notes skill or a sibling html-anything skill MUST point at chapter 16 and the vendor path. Pack files MUST stay under 150 lines except the vendored upstream SKILL.md.

#### Scenario: Pack lists the skill
- **WHEN** an agent lists Dorph skills
- **THEN** html-anything is present
- **AND** the pack text names `docs/agents/16-dorph-for-agents.md` or the vendor path
