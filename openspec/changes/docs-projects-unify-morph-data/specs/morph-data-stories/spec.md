## Purpose

Defines Morph Data Stories as a task-like note workspace: tile cards, detail on click, and notes instead of post comments.

## ADDED Requirements

### Requirement: Stories replaces Story Board
Morph Data navigation SHALL label the former Story Board as **Stories**. User-facing “Story Board” as the primary nav label MUST NOT remain.

#### Scenario: Drawer shows Stories
- **WHEN** an authenticated user views Morph Data navigation
- **THEN** the story surface is labeled Stories

### Requirement: Stories shown as tile cards
Stories list view SHALL present stories as a grid/tile of cards similar to Tasks card layouts. Each card MUST show enough identity (title and/or preview) to choose a story.

#### Scenario: Tile grid listing
- **WHEN** a user opens Stories
- **THEN** stories appear as cards/tiles rather than only as a chronological post feed

### Requirement: Click opens story detail
Selecting a story card SHALL open a detail view for that story (drawer, page, or modal). The detail view MUST show the story content for reading/editing according to existing permissions.

#### Scenario: Open story from card
- **WHEN** a user clicks a story card
- **THEN** the story detail view opens for that story

### Requirement: Stories are notes; comments become notes
Stories SHALL be treated as notes (not social posts). The former comments affordance on a story MUST be labeled and presented as **notes** (create/list/delete of note entries on the story).

#### Scenario: Add a note on a story
- **WHEN** a user opens a story detail and adds a note
- **THEN** the note is saved against that story
- **AND** the UI labels the affordance as notes (not comments/posts)
