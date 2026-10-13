# html-anything-base-skill Specification

## Purpose

Gives Morph one vendored html-anything skill, with license and Ge style overlay, loaded in compact form on every HTML-producing AI path.

## Requirements

### Requirement: Vendored skill keeps license and attribution
The repository MUST keep html-anything skill text and needed assets under `skills/vendor/html-anything/`. That tree MUST include the upstream Apache-2.0 LICENSE and an attribution notice that names nexu-io/html-anything.

#### Scenario: Vendor tree is present
- **WHEN** a reviewer opens `skills/vendor/html-anything/`
- **THEN** the directory contains LICENSE, an attribution notice, and SKILL.md
- **AND** the LICENSE text is Apache-2.0
- **AND** the notice names https://github.com/nexu-io/html-anything

### Requirement: Compact injection stays inside the Ollama budget
The text Morph injects for this skill MUST be a condensed form plus Ge's style overlay. It MUST stay at or under 512 tokens, counted as Unicode runes divided by four rounding up.

#### Scenario: Budget test
- **WHEN** the compact injected skill text is measured
- **THEN** the token estimate is 512 or fewer
- **AND** the text still names html-anything, Arial or a sans stack, a dark theme, and more than one accent color

### Requirement: Every HTML-producing path loads the skill
Note AI output, generated HTML or markdown, grill final content, and agent-facing skill docs MUST include the compact html-anything skill and Ge overlay. Ordinary chat that is not producing HTML MUST NOT receive the full vendor SKILL.md.

#### Scenario: Note AI and generated documents
- **WHEN** Morph builds a note-AI, generated HTML/MD, or form-HTML prompt
- **THEN** that prompt includes the compact skill text
- **AND** it includes Ge's Arial, dark, multi-accent overlay

#### Scenario: Grill final content
- **WHEN** the grill builds the write prompt for the draft
- **THEN** the prompt includes the compact skill text

#### Scenario: Ordinary chat stays compact
- **WHEN** Morph builds a normal chat reply prompt that is not asking for HTML
- **THEN** it does not append the full vendor SKILL.md

### Requirement: Compact skill uses hex colours only
The compact html-anything injection MUST name accent colours as hex values only. It MUST NOT place a colour word immediately before a hex value (for example `cyan #38bdf8`). The injected text MUST still stay at or under 512 tokens (Unicode runes divided by four, rounding up).

#### Scenario: Hex only under budget
- **WHEN** the compact injected skill text is inspected
- **THEN** it contains `#38bdf8` and `#a78bfa`
- **AND** it does not contain `cyan #38bdf8` or `violet #a78bfa`
- **AND** the token estimate is 512 or fewer
