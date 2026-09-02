## Context

See proposal.md for motivation.

AI Tools is embedded in Morph as an iframe (`AiToolsWorkspaceDrawer`). Assistants are listed in `bk/frontend` `AssistantManager.js` with a Run icon that opens a query dialog (`POST /assistants/{id}/run`). Morph chat (`SkoolAiChat.js`) already posts to `/api/chat` and the Morph backend already applies AI Tools assistants when `agent_id` is `bk:<assistantId>` (`parseBKAssistantID` → `buildBKAssistantInstructions`). The frontend never sends `agent_id` today, so Run never reaches Morph chat.

Constraints: cross-origin iframe (`localhost:3000` inside `localhost:3031`); no new Morph API; keep `POST /assistants/{id}/run` on the bk backend for other callers.

## Goals / Non-Goals

**Goals:**

- Replace Run with Apply/Dismiss on the Assistants list.
- Bridge the iframe to Morph chat so Apply sets the chat's `agent_id` and Dismiss clears it.
- Show applied state in both AI Tools and Morph chat, kept in sync.

**Non-Goals:**

- Restoring or replacing the one-shot Run dialog.
- Stacking multiple assistants.
- New backend apply endpoints.
- Rebuilding a Morph sidebar assistants picker (AI Tools is the apply surface).
- Changing how Morph compiles BK assistant instructions once `agent_id` is present.

## Decisions

### D1: Morph owns applied-assistant state; iframe signals apply/dismiss

**Choice:** Morph chat holds `{ id, name } | null`. The Assistants page posts apply/dismiss; Morph replies with current state so the iframe can show Apply vs Dismiss.

**Why:** Chat is what must send `agent_id`. The iframe is a guest and can be closed; state must survive the drawer closing.

**Alternatives:** Store only in the iframe (lost when the drawer unmounts). Store only in bk (Morph would poll; extra API). Rejected.

### D2: `postMessage` plus `BroadcastChannel` for the same events

**Choice:** Shared event types:

- `morph-ai:apply-assistant` — `{ id, name }` (bk raw id, not `bk:` prefix)
- `morph-ai:dismiss-assistant`
- `morph-ai:request-applied-assistant` — iframe asks Morph for current state on load
- `morph-ai:applied-assistant-state` — Morph → iframe `{ id, name } | null`

When AI Tools is in Morph's iframe, use `window.parent.postMessage` with a fixed origin (bk origin from `REACT_APP_BK_URL` / Morph origin). Also post on `BroadcastChannel('morph-ai-applied-assistant')` so Apply from "Open in tab" still reaches an open Morph chat tab.

Ignore messages that fail origin / type checks.

**Why:** Iframe is the common path; BroadcastChannel covers the explicit "Open in tab" path without a backend.

**Alternatives:** `localStorage` events (same-origin only; iframe is cross-origin). Query-string on iframe src (cannot update after load). Rejected.

### D3: Chat payload uses existing `bk:<id>` agent id

**Choice:** Morph maps applied `{ id }` → `agent_id: "bk:" + id` on JSON and multipart `/api/chat`. Dismiss omits `agent_id`.

**Why:** Backend already implements this. No API change.

**Alternatives:** Call `POST /assistants/{id}/run` from Morph (loses Morph sessions/tools). New proxy. Rejected.

### D4: Persist applied assistant in Morph `sessionStorage`

**Choice:** Key `morph-ai:applied-assistant`. Restore on Morph chat mount. Clear on Dismiss.

**Why:** Drawer close and chat remount should not drop the applied assistant within the tab. Do not persist across tabs/sessions (sessionStorage).

**Alternatives:** `localStorage` (lingers across visits). Server-side per user (out of scope). Rejected.

### D5: UI — Apply/Dismiss on the card; chip in Morph chat

**Choice:** On each assistant card, Apply (play/check icon + "Apply") when not applied; Dismiss when this card is the applied one. Other cards stay Apply. In Morph chat, a compact chip (assistant name + Dismiss) near the header or composer, same visual language as the skills badge.

**Why:** Matches the request (Apply instead of Run, Dismiss to clear) and lets users dismiss without reopening AI Tools.

**Alternatives:** Apply-only in AI Tools, dismiss-only in Morph. Rejected — user asked for dismiss on the assistant as well.

### D6: Standalone AI Tools with no Morph listener

**Choice:** After Apply, if not in an iframe and no BroadcastChannel consumer is guaranteed, show a short toast: Morph AI is required to apply this assistant to chat. Do not fake applied state in the list.

**Why:** Spec forbids silently treating Apply as success when Morph cannot receive it.

## Risks / Trade-offs

- **[Cross-origin blocked / wrong origin]** → Pin allowed origins from env (`REACT_APP_BK_URL`, Morph origin). Ignore unknown senders.
- **[Iframe not yet loaded when Morph sends state]** → Iframe requests state on Assistants mount; Morph answers.
- **[Apply from a tab with Morph closed]** → Toast; user opens Morph AI then Apply again (or Morph already open in another tab via BroadcastChannel).
- **[Stale applied id after assistant deleted]** → Chat already returns "Unknown or unavailable AI tools assistant"; Morph dismisses and shows that error once.
- **[Trade-off]** Run dialog is gone from this page; power users lose one-shot query. Mitigation: Morph chat is the run surface.

## Migration Plan

- Ship frontend-only (bk Assistants page + Morph chat + drawer message wiring).
- No data migration. Rollback: revert the two frontends; Run dialog returns; chat without `agent_id` is current behavior.

## Open Questions

None. Apply vs Dismiss labels, single applied assistant, and Morph-owned `agent_id` are fixed by the request and existing chat API.
