## Purpose

Makes in-stack AI chats answer visually first: mermaid diagrams and charts in the bubble, with occasional pixel art, instead of long prose.

## ADDED Requirements

### Requirement: Visual-first chat contract

In-stack AI chats (Morph AI, Event Logs assistant, Content Maker assistant) MUST prefer a short caption plus a diagram or chart whenever the answer is structure, process, comparison, or quantities. Programming-like content MUST use mermaid flow, sequence, or class diagrams. Quantity breakdowns MUST use a chart (pie or xy). Single Morph Data records MUST stay compact labeled fields, not fake charts.

#### Scenario: Process or code-shaped answer

- **WHEN** the operator asks how a flow, API, or program is structured
- **THEN** the assistant reply includes a mermaid diagram
- **AND** prose is limited to a short caption (not a long essay)

#### Scenario: Numeric breakdown

- **WHEN** the assistant has counts, shares, or a comparison across categories
- **THEN** the reply includes a chart
- **AND** does not paste the same numbers as a long paragraph instead of the chart

#### Scenario: One Morph Data record

- **WHEN** the assistant summarizes a single person, place, or similar record
- **THEN** it uses compact labeled fields
- **AND** it does not invent a chart that adds no information

### Requirement: Chat renders diagrams and occasional pixel art

Chat UIs that show assistant markdown MUST draw mermaid fences as diagrams (dark theme). Invalid mermaid MUST fall back to the source text. Pixel art MUST appear when the image-generator agent is used, or when the assistant emits a compact pixel grid fence in ordinary chat. The system MUST NOT call an image API on every chat turn.

#### Scenario: Mermaid fence in Morph AI chat

- **WHEN** an assistant message contains a mermaid code fence
- **THEN** the operator sees a rendered dark-theme diagram, not only a code block

#### Scenario: Broken mermaid

- **WHEN** mermaid source cannot be drawn
- **THEN** the operator still sees the source text

#### Scenario: Pixel art without image API spam

- **WHEN** the operator uses the image-generator agent for a picture
- **THEN** the result is shown as an image in the chat (pixel-art style when that is the prompt)
- **WHEN** an ordinary chat includes a compact pixel grid fence
- **THEN** the chat draws that grid as pixel art
- **AND** other turns do not automatically generate images
