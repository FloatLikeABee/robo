# Proposal

## Why

Publish actions wear the same pale accent stroke as the panel, so they read as labels. The source field stays a short box while the rest of the column is empty, and the two generate actions sit side by side with long labels, so generating from data does not read as the main action.

## What Changes

- The source text area grows into the unused height of the source column.
- Source actions stack full width, top to bottom. Generate from data is the last control and the only primary action in that column.
- Those labels shorten to Upload, From text, and From data.
- Every action in the publish module is a filled control. The panel accent stroke is no longer the thing that marks a button.
- Publish in the header stays the other primary action. Save, Apply, and the source secondaries are filled secondary buttons.

## Capabilities

### New Capabilities

- `content-publish-workspace`: How the publish workspace lays out the source column and how its actions look and read.

### Modified Capabilities

## Impact

- `composerx/frontend/src/components/ComposePublishPanel.svelte`
- `composerx/frontend/src/components/ComposePublishRecordsPanel.svelte` where it shares those action styles
- No API or data change
