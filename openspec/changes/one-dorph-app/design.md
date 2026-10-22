## Context

See proposal.md for why. Dorph AI and DorphNotes are one CRA (`/` and `/dorphdata`). DorphUtils on port 3040 is a separate shell that iframes Event Logs, Content Maker, Data Access, and Project. AI tools is already a drawer in Dorph AI, capped at `min(96vw, 1200px)`. The Project header uses `bg-surface/30`, so it is translucent. Selected pills use `bg-accent/20`. The document preview gradient starts at indigo `#1e1b4b`. That combination is the mauve in the screenshot.

## Goals / Non-Goals

**Goals:**

- Dorph AI is the only app the operator opens.
- DorphNotes and AI tools are the same near-fullscreen modal size.
- Event Logs, Content Maker, and Project are sections inside DorphNotes.
- Data Access and the DorphUtils shell are not part of the running product.
- Project chrome is opaque navy.

**Non-Goals:**

- Rewriting Event Logs, Content Maker, or Project into the Dorph CRA.
- Deleting the SharpReport or `dorph-utils` trees, or removing their container specs.
- Changing Event Logs away from `/events-info`, or restoring an Info Sheets tab.
- Restoring a Dorph AI Files workspace.
- A light/dark switch.

## Decisions

### 1. One shell, two modals

Dorph AI stays on screen. DorphNotes and AI tools each open a modal that is 96vw by 96dvh, sharing one size class. The current AI tools cap of 1200px is what makes that drawer feel like a side panel.

**Why not navigate to `/dorphdata`:** Leaving the chat is the separate-app feeling this removes. The modal closes back to the same conversation.

**Why a same-origin iframe for DorphNotes:** `/dorphdata` is already the DorphNotes UI in this CRA. Iframing it reuses the drawer, routes, and session cookie without mounting a second router over the chat. Cross-origin modules stay iframes inside that UI.

### 2. Move the three modules, do not merge their codebases

DorphNotes navigation gains Event Logs, Content Maker, and Project. Each panel iframes the existing server, with the same bearer query (`userspanel_token`) DorphUtils used, and the same down hint when the origin refuses. Event Logs keeps `/events-info`.

**Why not delete DorphUtils and SharpReport now:** The product stops linking and starting them. Deleting those trees, Render services, and container specs is a second change. An explicit `start sharpreport` or `start dorph-utils` MAY remain for the leftover code. `all` and a bare `./start-all.sh` do not start them. Event Logs, Content Maker, and Project servers still start, because DorphNotes embeds them.

### 3. Data Access is omitted, not hidden

No nav item, no iframe, no default process. Do not keep a disabled tile.

### 4. Opaque navy, not a translucent header

The mauve header is `bg-surface/30` over the page, and the preview gradient starts at `#1e1b4b`. Selected pills go through `bg-accent/20`, which is reading as pink. Set the header to opaque `#0c1220`. Paint selected tabs, the selected project, and the HTML/Markdown pill with `#3b82f6` at low opacity. Start the preview gradient from `#0c1220`, not indigo.

**Why not another token rename:** Renaming `--color-violet` to blue did not change the translucent header or the indigo preview. Those are the pixels in the screenshot.

## Risks / Trade-offs

- [DorphNotes inside an iframe inside a modal scrolls twice] → The modal body is the iframe; DorphNotes fills that frame. Do not add a third scroll container around it.
- [Cross-origin embeds still need the bearer on the URL] → Keep the existing `userspanel_token` handoff. Do not invent a shared parent-domain cookie.
- [Leftover DorphUtils and SharpReport code can still be started by name] → They are not in `all` and not linked. A later change can delete the trees.
- [Direct `/dorphdata` bookmarks open DorphNotes without the chat modal] → Those URLs keep working. The header control uses the modal.

## Migration Plan

- No data migration. Operators use Dorph AI. Old DorphUtils bookmarks on port 3040 are unused unless someone starts that UI on purpose.
- Rollback is restoring the header link and the start-all entries.

## Open Questions

None. Modal size is 96vw by 96dvh. Data Access and DorphUtils stay in the repo and off the default start.
