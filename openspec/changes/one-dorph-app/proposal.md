## Why

Dorph is still several products: Dorph AI, a DorphUtils shell, and a Data Access module that should not be part of the product. Event Logs, Content Maker, and Project belong inside DorphNotes, and DorphNotes itself should open inside Dorph AI the way AI tools already does. The Project screen in the screenshot is still mauve across the header, the selected row, and the HTML pill.

## What Changes

- **BREAKING:** Data Access is not a product module. It is not in any navigation, and `./start-all.sh` (including `all`) does not start it.
- **BREAKING:** DorphUtils is not a product shell. The Dorph AI header has no DorphUtils chip. `./start-all.sh` does not start the DorphUtils UI. Event Logs, Content Maker, and Project are reached only from DorphNotes.
- DorphNotes opens from Dorph AI as a large modal, same idea as AI tools. Both modals cover most of the window (about 96% of the width and height), not a 1200px-wide drawer.
- DorphNotes keeps Stories, Stick notes, Timelines, Research, and Generic data, and adds Event Logs, Content Maker, and Project as embedded modules. Those three stay their own servers. Data Access is not added.
- Project chrome is opaque dark blue. The header is not translucent. Selected rows and section pills use blue. The document preview gradient does not start from indigo.

## Capabilities

### New Capabilities

- `no-data-access-product`: Data Access is absent from the product and from the default start set.
- `dorphnotes-module-host`: DorphNotes, inside Dorph AI, hosts Event Logs, Content Maker, and Project. DorphUtils is not an entry.
- `large-shell-modals`: DorphNotes and AI tools modals cover most of the Dorph AI window.
- `project-opaque-navy`: Project header, selection, and preview wash are opaque dark blue.

### Modified Capabilities

- `dorph-utils-header-url`: The DorphUtils header chip is removed. DorphNotes opens in the modal instead.
- `dorphutils-embed-auth`: The embed host is the DorphNotes modal, for Event Logs, Content Maker, and Project only. Data Access is not an embed.

## Impact

- Dorph AI header and login copy, AI tools drawer width, new DorphNotes modal.
- DorphNotes navigation (`AppDrawer`) plus iframe panels for the three modules.
- Project frontend header, selected states, and preview gradient (`dorph-engi`).
- `start-all.sh`: default/`all` drops `dorph-utils-ui`, `sharpreport-api`, and `sharpreport-ui`. Event Logs, Content Maker, and Project servers still start.
- Header-link tests and the two archived specs above.
- Do not delete the SharpReport or `dorph-utils` trees in this change. Do not restore a Dorph AI Files workspace.
