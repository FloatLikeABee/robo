## Context

See proposal.md for motivation.

AI Tools is embedded in Dorph as an iframe (`AiToolsWorkspaceDrawer`). Assistants are listed in `bk/frontend` `AssistantManager.js` with a Run icon that opens a query dialog (`POST /assistants/{id}/run`). Dorph chat (`SkoolAiChat.js`) already posts to `/api/chat` and the Dorph backend already applies AI Tools assistants when `agent_id` is `bk:<assistantId>` (`parseBKAssistantID` → `buildBKAssistantInstructions`). The frontend never sends `agent_id` today, so Run never reaches Dorph chat.

Constraints: cross-origin iframe (`localhost:3000` inside `localhost:3031`); no new Dorph API; keep `POST /assistants/{id}/run` on the bk backend for other callers.

## Goals / Non-Goals

**Goals:**

- Replace Run with Apply/Dismiss on the Assistants list.
- Bridge the iframe to Dorph chat so Apply sets the chat's `agent_id` and Dismiss clears it.
- Show applied state in both AI Tools and Dorph chat, kept in sync.

**Non-Goals:**

- Restoring or replacing the one-shot Run dialog.
- Stacking multiple assistants.
- New backend apply endpoints.
- Rebuilding a Dorph sidebar assistants picker (AI Tools is the apply surface).
- Changing how Dorph compiles BK assistant instructions once `agent_id` is present.

## Decisions

### D1: Dorph owns applied-assistant state; iframe signals apply/dismiss

**Choice:** Dorph chat holds `{ id, name } | null`. The Assistants page posts apply/dismiss; Dorph replies with current state so the iframe can show Apply vs Dismiss.

**Why:** Chat is what must send `agent_id`. The iframe is a guest and can be closed; state must survive the drawer closing.

**Alternatives:** Store only in the iframe (lost when the drawer unmounts). Store only in bk (Dorph would poll; extra API). Rejected.

### D2: `postMessage` plus `BroadcastChannel` for the same events

**Choice:** Shared event types:

- `dorph-ai:apply-assistant` — `{ id, name }` (bk raw id, not `bk:` prefix)
- `dorph-ai:dismiss-assistant`
- `dorph-ai:request-applied-assistant` — iframe asks Dorph for current state on load
- `dorph-ai:applied-assistant-state` — Dorph → iframe `{ id, name } | null`

When AI Tools is in Dorph's iframe, use `window.parent.postMessage` with a fixed origin (bk origin from `REACT_APP_BK_URL` / Dorph origin). Also post on `BroadcastChannel('dorph-ai-applied-assistant')` so Apply from "Open in tab" still reaches an open Dorph chat tab.

Ignore messages that fail origin / type checks.

**Why:** Iframe is the common path; BroadcastChannel covers the explicit "Open in tab" path without a backend.

**Alternatives:** `localStorage` events (same-origin only; iframe is cross-origin). Query-string on iframe src (cannot update after load). Rejected.

### D3: Chat payload uses existing `bk:<id>` agent id

**Choice:** Dorph maps applied `{ id }` → `agent_id: "bk:" + id` on JSON and multipart `/api/chat`. Dismiss omits `agent_id`.

**Why:** Backend already implements this. No API change.

**Alternatives:** Call `POST /assistants/{id}/run` from Dorph (loses Dorph sessions/tools). New proxy. Rejected.

### D4: Persist applied assistant in Dorph `sessionStorage`

**Choice:** Key `dorph-ai:applied-assistant`. Restore on Dorph chat mount. Clear on Dismiss.

**Why:** Drawer close and chat remount should not drop the applied assistant within the tab. Do not persist across tabs/sessions (sessionStorage).

**Alternatives:** `localStorage` (lingers across visits). Server-side per user (out of scope). Rejected.

### D5: UI — Apply/Dismiss on the card; chip in Dorph chat

**Choice:** On each assistant card, Apply (play/check icon + "Apply") when not applied; Dismiss when this card is the applied one. Other cards stay Apply. In Dorph chat, a compact chip (assistant name + Dismiss) near the header or composer, same visual language as the skills badge.

**Why:** Matches the request (Apply instead of Run, Dismiss to clear) and lets users dismiss without reopening AI Tools.

**Alternatives:** Apply-only in AI Tools, dismiss-only in Dorph. Rejected — user asked for dismiss on the assistant as well.

### D6: Standalone AI Tools with no Dorph listener

**Choice:** After Apply, if not in an iframe and no BroadcastChannel consumer is guaranteed, show a short toast: Dorph AI is required to apply this assistant to chat. Do not fake applied state in the list.

**Why:** Spec forbids silently treating Apply as success when Dorph cannot receive it.

## Risks / Trade-offs

- **[Cross-origin blocked / wrong origin]** → Pin allowed origins from env (`REACT_APP_BK_URL`, Dorph origin). Ignore unknown senders.
- **[Iframe not yet loaded when Dorph sends state]** → Iframe requests state on Assistants mount; Dorph answers.
- **[Apply from a tab with Dorph closed]** → Toast; user opens Dorph AI then Apply again (or Dorph already open in another tab via BroadcastChannel).
- **[Stale applied id after assistant deleted]** → Chat already returns "Unknown or unavailable AI tools assistant"; Dorph dismisses and shows that error once.
- **[Trade-off]** Run dialog is gone from this page; power users lose one-shot query. Mitigation: Dorph chat is the run surface.

## Migration Plan

- Ship frontend-only (bk Assistants page + Dorph chat + drawer message wiring).
- No data migration. Rollback: revert the two frontends; Run dialog returns; chat without `agent_id` is current behavior.

## Open Questions

None. Apply vs Dismiss labels, single applied assistant, and Dorph-owned `agent_id` are fixed by the request and existing chat API.
