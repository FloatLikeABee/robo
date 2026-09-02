## Purpose

Keeps Content Maker Compose content work on screen when the operator switches to Published contents and returns, so in-progress HTML is not replaced by the starter page.

## ADDED Requirements

### Requirement: Compose draft survives inner navigation
Switching from **Compose content** to **Published contents** (Content Maker header/nav tabs) MUST NOT discard the unsaved compose session: page name, HTML body, and AI chat messages. Returning to Compose content MUST show the same values without requiring Save draft or Publish.

#### Scenario: Edit then open Published contents then return
- **WHEN** the operator replaces the starter HTML with distinct body text (and optionally a name)
- **AND** they open Published contents
- **AND** they return to Compose content
- **THEN** the compose editor still shows that body text (and name if they set one)

#### Scenario: Starter page is not restored over real work
- **WHEN** the compose HTML is no longer the empty starter document
- **AND** the operator leaves Compose content and comes back in the same Content Maker session
- **THEN** the UI MUST NOT replace that HTML with the default “Start composing” starter

### Requirement: Explicit save still works
Publish and Save draft MUST continue to persist to the server. Surviving a tab switch MUST NOT by itself create a published page.

#### Scenario: Tab switch does not publish
- **WHEN** the operator has unpublished compose HTML and opens Published contents
- **THEN** no new published page is created solely from that navigation
