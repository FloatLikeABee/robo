# Design

## Context

See `proposal.md`. DorphNotes drawer lists Generic data (`/dorphdata/generic-data`, in-app grid) and Content Maker (`/dorphdata/content-maker`, composerx iframe). Analysis lives on `generic_data.ai_analysis`. HTML publish already exists as `POST /publishes` (`name`, `theme`, `html_content`) behind Dorph’s `/api/composerx/*` proxy. The Dorph AI title uses `-webkit-text-fill-color: transparent` plus an accent gradient, which disappears into twilight glass.

## Goals / Non-Goals

**Goals**

- One nav label: Contents & Data.
- One page: Data pane + Publish pane.
- Publish-from-record uses composerx, not a new renderer.
- Header title is light ink.

**Non-Goals**

- Renaming `composerx/` or deleting generic-data APIs.
- Event Logs merge, new markdown engine, theme toggle.

## Decisions

### 1. How to combine the two tools (grilled)

| Approach | Pros | Cons |
|----------|------|------|
| **Shell page with Data + Publish tabs (recommended)** | Reuses GenericData.js and EmbeddedModule; one nav | Extra wrapper |
| Iframe Generic data inside composerx | One app chrome | Cross-origin, two backends, worse |
| Drop composerx from nav; only a publish button | Smaller nav | Owner still needs the HTML editor |

**Choice:** wrapper route `/dorphdata/contents-data` (aliases: `content-maker`, `generic-data`). Tabs: **Data** (existing GenericData) and **Publish** (existing `EmbeddedModule id=content-maker`). Default tab Data; `?tab=publish` opens the iframe.

Agent Commons board: same — one destination, row action, no third renderer, no grid-in-iframe.

### 2. Publish-from-analysis

**Choice:** On a record, **Publish as HTML** builds a small HTML document (title, analysis, optional table excerpt from `detail.payload`) and `POST /api/composerx/publishes`. Then switch to the Publish tab (or show the public path). Reuse proxy; no new Go route unless the proxy cannot send HTML (it can).

Rejected: asking publish-chat LLM to restyle every time (slow, needs composerx up + extra tokens). Rejected: storing HTML only in generic_data (then it is not a Content Maker page).

### 3. Title color

**Choice:** `.app-title` uses `color: var(--chat-text)` and drops gradient text-fill. Keep `::before` accent bar. Do not invent a new theme.

Rejected: another glass pass; orange title (competes with send button).

### 4. Labels and tests

- `messages.js`: `contentMaker` → “Contents & Data”; drop drawer use of `genericData` (string may remain for in-page headings as “Data”).
- Embed list label in `dorphNotesEmbeds.js` matches.
- Default DorphNotes index: Contents & Data (still the generic-data surface as the Data tab).

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| ComposerX down | Analyze/import still work; publish shows existing not-configured/down copy |
| Fat HTML from huge tables | Cap excerpt rows (e.g. 50) with a `ponytail` note |
| Bookmark to Generic data | Redirect into combined page |

## Migration Plan

Ship frontend shell + redirects + publish button. No data migration. Rollback: restore two drawer rows.

## Open Questions

None blocking. Exact tab control (MUI Tabs vs segmented buttons) is apply-time.
