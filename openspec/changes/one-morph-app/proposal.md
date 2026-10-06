## Why

Morph is still several products: Morph AI, a MorphUtils shell, and a Data Access module that should not be part of the product. Event Logs, Content Maker, and Project belong inside MorphNotes, and MorphNotes itself should open inside Morph AI the way AI tools already does. The Project screen in the screenshot is still mauve across the header, the selected row, and the HTML pill.

## What Changes

- **BREAKING:** Data Access is not a product module. It is not in any navigation, and `./start-all.sh` (including `all`) does not start it.
- **BREAKING:** MorphUtils is not a product shell. The Morph AI header has no MorphUtils chip. `./start-all.sh` does not start the MorphUtils UI. Event Logs, Content Maker, and Project are reached only from MorphNotes.
- MorphNotes opens from Morph AI as a large modal, same idea as AI tools. Both modals cover most of the window (about 96% of the width and height), not a 1200px-wide drawer.
- MorphNotes keeps Stories, Stick notes, Timelines, Research, and Generic data, and adds Event Logs, Content Maker, and Project as embedded modules. Those three stay their own servers. Data Access is not added.
- Project chrome is opaque dark blue. The header is not translucent. Selected rows and section pills use blue. The document preview gradient does not start from indigo.

## Capabilities

### New Capabilities

- `no-data-access-product`: Data Access is absent from the product and from the default start set.
- `morphnotes-module-host`: MorphNotes, inside Morph AI, hosts Event Logs, Content Maker, and Project. MorphUtils is not an entry.
- `large-shell-modals`: MorphNotes and AI tools modals cover most of the Morph AI window.
- `project-opaque-navy`: Project header, selection, and preview wash are opaque dark blue.

### Modified Capabilities

- `morph-utils-header-url`: The MorphUtils header chip is removed. MorphNotes opens in the modal instead.
- `morphutils-embed-auth`: The embed host is the MorphNotes modal, for Event Logs, Content Maker, and Project only. Data Access is not an embed.

## Impact

- Morph AI header and login copy, AI tools drawer width, new MorphNotes modal.
- MorphNotes navigation (`AppDrawer`) plus iframe panels for the three modules.
- Project frontend header, selected states, and preview gradient (`morph-engi`).
- `start-all.sh`: default/`all` drops `morph-utils-ui`, `sharpreport-api`, and `sharpreport-ui`. Event Logs, Content Maker, and Project servers still start.
- Header-link tests and the two archived specs above.
- Do not delete the SharpReport or `morph-utils` trees in this change. Do not restore a Morph AI Files workspace.
