## Purpose

Removes Video stories and the video generator from AI tools so that product only keeps Assistants, RAG, Documents, and System.

## ADDED Requirements

### Requirement: Video stories is not in AI tools
AI tools MUST NOT show a Video stories (or Video generator) nav item, route, or page. Requests to former video-story API paths MUST NOT serve the generator (404 or equivalent). Existing Assistants, RAG, Documents, and System surfaces MUST remain.

#### Scenario: Header has no Video stories
- **WHEN** an operator opens AI tools
- **THEN** the app header does not include Video stories
- **AND** `/video-stories` is not a working generator page

#### Scenario: API no longer lists video stories
- **WHEN** a client calls `GET /video-stories` on the AI tools API
- **THEN** the call does not return a video-story project list as a supported feature (not found or removed)
